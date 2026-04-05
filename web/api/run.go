package api

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/bojand/ghz/runner"
	"github.com/bojand/ghz/web/model"
	"github.com/labstack/echo"
)

// RunDatabase abstracts DB operations we need (reuse from ingest)
type RunDatabase interface {
	IngestDatabase
}

// RunRequest defines the JSON body accepted by /api/run/ endpoints
// Minimal initial version; can be extended.
type RunRequest struct {
	Call        string            `json:"call" validate:"required"`      // fully qualified method e.g. helloworld.Greeter.SayHello
	Host        string            `json:"host" validate:"required"`      // target host:port
	Concurrency uint              `json:"concurrency" validate:"gt=0"`   // -c
	Total       uint              `json:"total" validate:"gt=0"`         // -n total requests (exclusive with Duration)
	DurationSec uint              `json:"duration_sec" validate:"gte=0"` // run duration seconds (exclusive with Total)
	DataJSON    map[string]any    `json:"data"`                          // simple JSON payload (one request)
	Metadata    map[string]string `json:"metadata"`                      // metadata headers
	Tags        map[string]string `json:"tags"`                          // tags into report
	Insecure    bool              `json:"insecure"`
	ProtoFile   string            `json:"proto_file"` // single proto content (raw text)
	ProtoPath   []string          `json:"proto_path"` // import paths (not implemented yet)
	Name        string            `json:"name"`       // optional report name override
	Async       bool              `json:"async"`      // if true, run asynchronously and return job id
}

// RunResponse response after executing run and persisting like ingest
type RunResponse struct {
	IngestResponse
}

// RunAPI executes a ghz run directly from UI request
// It converts RunRequest -> runner options -> runner.Report -> persist (as ingest)
// For now only simple JSON body and single proto file content are supported.
type RunAPI struct {
	DB   RunDatabase
	Jobs *RunJobManager
}

// Run executes the test creating a new project implicitly
func (api *RunAPI) Run(ctx echo.Context) error {
	rr := new(RunRequest)
	if err := ctx.Bind(rr); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if ctx.Echo().Validator != nil {
		if err := ctx.Validate(rr); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
	}

	// create project (same as ingest)
	p := new(model.Project)
	if err := api.DB.CreateProject(p); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return api.execute(rr, p, ctx)
}

// RunToProject executes test into existing project
func (api *RunAPI) RunToProject(ctx echo.Context) error {
	var project *model.Project
	var err error
	if project, err = findProject(api.DB.FindProjectByID, ctx); err != nil {
		return err
	}
	rr := new(RunRequest)
	if err := ctx.Bind(rr); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if ctx.Echo().Validator != nil {
		if err := ctx.Validate(rr); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
	}
	return api.execute(rr, project, ctx)
}

// GetJob returns async job status
func (api *RunAPI) GetJob(ctx echo.Context) error {
	if api.Jobs == nil {
		return echo.NewHTTPError(http.StatusNotFound, "jobs not enabled")
	}
	jid := ctx.Param("jid")
	if jid == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "missing job id")
	}
	job, ok := api.Jobs.Get(jid)
	if !ok {
		return echo.NewHTTPError(http.StatusNotFound, "job not found")
	}
	return ctx.JSON(http.StatusOK, job)
}

func (api *RunAPI) execute(rr *RunRequest, p *model.Project, ctx echo.Context) error {
	if rr.Async {
		if api.Jobs == nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "job manager not configured")
		}
		job := api.Jobs.NewJob(p.ID)
		api.Jobs.Start(job.ID)
		go func(jobID string, rrCpy *RunRequest, proj *model.Project) {
			resp, err := api.runAndPersist(rrCpy, proj)
			if err != nil {
				api.Jobs.Fail(jobID, err.Error())
			} else {
				api.Jobs.Succeed(jobID, resp)
			}
		}(job.ID, rr, p)
		return ctx.JSON(http.StatusAccepted, map[string]any{"job_id": job.ID, "status": job.Status})
	}

	resp, err := api.runAndPersist(rr, p)
	if err != nil {
		if he, ok := err.(*echo.HTTPError); ok {
			return he
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return ctx.JSON(http.StatusCreated, resp)
}

// runAndPersist executes the ghz run and persists report returning ingest-like response.
func (api *RunAPI) runAndPersist(rr *RunRequest, p *model.Project) (*IngestResponse, error) {
	options := []runner.Option{}

	if rr.Concurrency > 0 {
		options = append(options, runner.WithConcurrency(rr.Concurrency))
	}
	if rr.Total > 0 && rr.DurationSec == 0 {
		options = append(options, runner.WithTotalRequests(rr.Total))
	}
	if rr.DurationSec > 0 && rr.Total == 0 {
		options = append(options, runner.WithRunDuration(time.Duration(rr.DurationSec)*time.Second))
	}
	if rr.Insecure {
		options = append(options, runner.WithInsecure(true))
	}
	if rr.DataJSON != nil {
		if b, err := json.Marshal(rr.DataJSON); err == nil {
			options = append(options, runner.WithDataFromJSON(string(b)))
		} else {
			return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
	}
	if len(rr.Metadata) > 0 {
		options = append(options, runner.WithMetadata(rr.Metadata))
	}
	if len(rr.Tags) > 0 {
		options = append(options, runner.WithTags(rr.Tags))
	}
	if rr.ProtoFile != "" {
		f, err := os.CreateTemp("", "ghz-proto-*.proto")
		if err != nil {
			return nil, echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		defer os.Remove(f.Name())
		if _, err := f.WriteString(rr.ProtoFile); err != nil {
			return nil, echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		f.Close()
		options = append(options, runner.WithProtoFile(f.Name(), rr.ProtoPath))
	}

	rep, err := runner.Run(rr.Call, rr.Host, options...)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	if rr.Name != "" {
		rep.Name = rr.Name
	}

	ir := (*IngestRequest)(rep)
	// persist
	// reuse existing logic
	// we cannot use ctx here (async case) so call persist helper directly but adapt to return response
	ingestResp, err := api.persistRunToProjectReturn(p, ir)
	if err != nil {
		return nil, err
	}
	return ingestResp, nil
}

// helper like persistRunToProject but returns response instead of writing JSON
func (api *RunAPI) persistRunToProjectReturn(p *model.Project, ir *IngestRequest) (*IngestResponse, error) {
	latest, _ := api.DB.FindLatestReportForProject(p.ID)

	report := convertIngestToReport(p.ID, ir)
	if err := api.DB.CreateReport(report); err != nil {
		return nil, echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	o := &model.Options{ReportID: report.ID}
	opts := model.OptionsInfo(ir.Options)
	o.Info = &opts
	if err := api.DB.CreateOptions(o); err != nil {
		return nil, echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	h := &model.Histogram{ReportID: report.ID}
	h.Buckets = make(model.BucketList, len(ir.Histogram))
	for i := range ir.Histogram {
		h.Buckets[i] = &ir.Histogram[i]
	}
	if err := api.DB.CreateHistogram(h); err != nil {
		return nil, echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	details := make([]*model.Detail, len(ir.Details))
	for i, v := range ir.Details {
		det := model.Detail{ReportID: report.ID, ResultDetail: v}
		details[i] = &det
	}
	created, errored := api.DB.CreateDetailsBatch(report.ID, details)

	if latest == nil || report.Date.After(latest.Date) {
		if err := api.DB.UpdateProjectStatus(p.ID, report.Status); err != nil {
			return nil, echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		p.Status = report.Status
	}

	rres := &IngestResponse{ // reuse structure
		Project:   p,
		Report:    report,
		Options:   o,
		Histogram: h,
		Details:   &DetailsCreated{Success: created, Fail: errored},
	}
	return rres, nil
}

// persistRunToProject duplicates ingestToProject logic for run api to avoid export changes
// legacy synchronous path still uses old function name via execute->runAndPersist; keep wrapper if needed
func (api *RunAPI) persistRunToProject(p *model.Project, ir *IngestRequest, ctx echo.Context) error {
	resp, err := api.persistRunToProjectReturn(p, ir)
	if err != nil {
		if he, ok := err.(*echo.HTTPError); ok {
			return he
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return ctx.JSON(http.StatusCreated, resp)
}
