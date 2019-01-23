package api

import (
	"net/http"

	"github.com/bojand/ghz/runner"
	"github.com/bojand/ghz/web/model"
	"github.com/labstack/echo"
)

// IngestDatabase interface for encapsulating database access.
type IngestDatabase interface {
	CreateProject(*model.Project) error
	CreateReport(*model.Report) error
	CreateHistogram(*model.Histogram) error
	CreateOptions(*model.Options) error
	FindProjectByID(uint) (*model.Project, error)
	FindLatestReportForProject(uint) (*model.Report, error)
	CreateDetailsBatch(uint, []*model.Detail) (uint, uint)
	UpdateProjectStatus(uint, model.Status) error
}

// IngestResponse is the response to the ingest endpoint
type IngestResponse struct {
	// Created project
	Project *model.Project `json:"project"`

	// Created report
	Report *model.Report `json:"report"`

	// Created Options
	Options *model.Options `json:"options"`

	// Created Histogram
	Histogram *model.Histogram `json:"histogram"`

	// The summary of created details
	Details *DetailsCreated `json:"details"`
}

// DetailsCreated summary of how many details got created and how many failed
type DetailsCreated struct {
	// Number of successfully created detail objects
	Success uint `json:"success"`

	// Number of failed detail objects
	Fail uint `json:"fail"`
}

// The IngestAPI provides handlers for ingesting and processing reports.
type IngestAPI struct {
	DB IngestDatabase
}

// IngestRequest is the raw report
type IngestRequest runner.Report

// Ingest creates data from raw report
func (api *IngestAPI) Ingest(ctx echo.Context) error {
	ir := new(IngestRequest)

	if err := bindAndValidateInput(ctx, ir); err != nil {
		return err
	}

	// Project
	p := new(model.Project)
	if err := api.DB.CreateProject(p); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return api.ingestToProject(p, ir, ctx)
}

// IngestToProject ingests data into a specific project
func (api *IngestAPI) IngestToProject(ctx echo.Context) error {
	var project *model.Project
	var err error

	if project, err = findProject(api.DB.FindProjectByID, ctx); err != nil {
		return err
	}

	ir := new(IngestRequest)

	if err := bindAndValidateInput(ctx, ir); err != nil {
		return err
	}

	return api.ingestToProject(project, ir, ctx)
}

func (api *IngestAPI) ingestToProject(p *model.Project, ir *IngestRequest, ctx echo.Context) error {

	// first get latest (we'll need it later)
	latest, _ := api.DB.FindLatestReportForProject(p.ID)

	// Report

	report := convertIngestToReport(p.ID, ir)
	if err := api.DB.CreateReport(report); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	// Options

	o := &model.Options{
		ReportID: report.ID,
	}

	opts := model.OptionsInfo(ir.Options)
	o.Info = &opts

	if err := api.DB.CreateOptions(o); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	// Histogram
	h := &model.Histogram{
		ReportID: report.ID,
	}

	h.Buckets = make(model.BucketList, len(ir.Histogram))
	for i := range ir.Histogram {
		h.Buckets[i] = &ir.Histogram[i]
	}

	if err := api.DB.CreateHistogram(h); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	// Details

	details := make([]*model.Detail, len(ir.Details))
	for i, v := range ir.Details {
		det := model.Detail{ReportID: report.ID, ResultDetail: v}
		details[i] = &det
	}

	created, errored := api.DB.CreateDetailsBatch(report.ID, details)

	// Update project status if needed
	if latest == nil || report.Date.After(latest.Date) {
		if err := api.DB.UpdateProjectStatus(p.ID, report.Status); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		p.Status = report.Status
	}

	// Response

	rres := &IngestResponse{
		Project:   p,
		Report:    report,
		Options:   o,
		Histogram: h,
		Details: &DetailsCreated{
			Success: created,
			Fail:    errored,
		},
	}

	return ctx.JSON(http.StatusCreated, rres)
}

func convertIngestToReport(pid uint, ir *IngestRequest) *model.Report {
	r := new(model.Report)
	r.ProjectID = pid

	r.Name = ir.Name
	r.EndReason = ir.EndReason.String()

	r.Date = ir.Date

	r.Count = ir.Count
	r.Total = ir.Total
	r.Average = ir.Average
	r.Fastest = ir.Fastest
	r.Slowest = ir.Slowest
	r.Rps = ir.Rps

	r.ErrorDist = ir.ErrorDist

	r.StatusCodeDist = ir.StatusCodeDist

	r.Tags = ir.Tags

	r.LatencyDistribution = make(model.LatencyDistributionList, len(ir.LatencyDistribution))
	for i := range ir.LatencyDistribution {
		r.LatencyDistribution[i] = &ir.LatencyDistribution[i]
	}

	// status
	r.Status = model.StatusOK

	if len(r.ErrorDist) > 0 {
		r.Status = model.StatusFail
	}

	return r
}

func bindAndValidateInput(ctx echo.Context, ir *IngestRequest) error {
	if err := ctx.Bind(ir); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if ctx.Echo().Validator != nil {
		if err := ctx.Validate(ir); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
	}

	return nil
}
