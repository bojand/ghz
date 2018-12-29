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
	FindProjectByID(uint) (*model.Project, error)
	CreateDetailsBatch(uint, []*model.Detail) (uint, uint)
}

// IngestResponse is the response to the ingest endpoint
type IngestResponse struct {
	// Created project
	Project *model.Project `json:"project"`

	// Created run
	Report *model.Report `json:"report"`

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

	if err := ctx.Bind(ir); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if ctx.Echo().Validator != nil {
		if err := ctx.Validate(ir); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
	}

	p := new(model.Project)

	if err := api.DB.CreateProject(p); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	report := convertIngestToReport(p.ID, ir)

	if err := api.DB.CreateReport(report); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	details := make([]*model.Detail, len(ir.Details))
	for i, v := range ir.Details {
		det := model.Detail{ReportID: report.ID, ResultDetail: v}
		details[i] = &det
	}

	created, errored := api.DB.CreateDetailsBatch(report.ID, details)

	rres := &IngestResponse{
		Project: p,
		Report:  report,
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

	opts := model.Options(*ir.Options)
	r.Options = &opts

	r.Date = ir.Date

	r.Count = ir.Count
	r.Total = ir.Total
	r.Average = ir.Average
	r.Fastest = ir.Fastest
	r.Slowest = ir.Slowest
	r.Rps = ir.Rps

	r.ErrorDist = ir.ErrorDist

	r.StatusCodeDist = ir.StatusCodeDist

	r.LatencyDistribution = make(model.LatencyDistributionList, len(ir.LatencyDistribution))
	for i, v := range ir.LatencyDistribution {
		r.LatencyDistribution[i] = &v
	}

	r.Histogram = make(model.BucketList, len(ir.Histogram))
	for i, v := range ir.Histogram {
		r.Histogram[i] = &v
	}

	r.Tags = ir.Tags

	return r
}
