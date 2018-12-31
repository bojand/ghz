package api

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/bojand/ghz/runner"
	"github.com/bojand/ghz/web/model"
	"github.com/labstack/echo"
)

// ExportDatabase interface for encapsulating database access.
type ExportDatabase interface {
	FindReportByID(uint) (*model.Report, error)
	GetHistogramForReport(uint) (*model.Histogram, error)
	GetOptionsForReport(uint) (*model.Options, error)
	ListAllDetailsForReport(uint) ([]*model.Detail, error)
}

// The ExportAPI provides handlers.
type ExportAPI struct {
	DB ExportDatabase
}

// JSONExportRespose is the response to JSON export
type JSONExportRespose struct {
	model.Report

	Options *model.OptionsInfo `json:"options,omitempty"`

	Histogram model.BucketList `json:"histogram"`

	Details []*runner.ResultDetail `json:"details"`
}

// GetExport does export for the report
func (api *ExportAPI) GetExport(ctx echo.Context) error {
	var id uint64
	var report *model.Report
	var err error

	format := strings.ToLower(ctx.QueryParam("format"))
	if format != "csv" && format != "json" {
		return echo.NewHTTPError(http.StatusBadRequest, "Unsupported format: "+format)
	}

	rid := ctx.Param("rid")
	if rid == "" {
		return echo.NewHTTPError(http.StatusNotFound, "")
	}

	if id, err = strconv.ParseUint(rid, 10, 32); err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}

	if report, err = api.DB.FindReportByID(uint(id)); err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}

	var options *model.Options
	if options, err = api.DB.GetOptionsForReport(uint(id)); err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}

	var histogram *model.Histogram
	if histogram, err = api.DB.GetHistogramForReport(uint(id)); err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}

	var details []*model.Detail
	if details, err = api.DB.ListAllDetailsForReport(uint(id)); err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}

	jsonRes := JSONExportRespose{
		Report: *report,
	}

	jsonRes.Options = options.Info

	jsonRes.Histogram = histogram.Buckets

	jsonRes.Details = make([]*runner.ResultDetail, len(details))
	for i := range details {
		jsonRes.Details[i] = &details[i].ResultDetail
	}

	return ctx.JSONPretty(http.StatusOK, jsonRes, "  ")
}
