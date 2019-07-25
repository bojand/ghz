package api

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/alecthomas/template"
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

const (
	csvTmpl = `
duration (ms),status,error{{ range $i, $v := . }}
{{ formatDuration .Latency 1000000 }},{{ .Status }},{{ .Error }}{{ end }}
`
)

var tmplFuncMap = template.FuncMap{
	"formatDuration": formatDuration,
}

func formatDuration(duration time.Duration, div int64) string {
	durationNano := duration.Nanoseconds()
	return fmt.Sprintf("%4.2f", float64(durationNano/div))
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

	if id, err = getReportID(ctx); err != nil {
		return err
	}

	var details []*model.Detail
	if details, err = api.DB.ListAllDetailsForReport(uint(id)); err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}

	if format == "csv" {
		outputTmpl := csvTmpl

		buf := &bytes.Buffer{}
		templ := template.Must(template.New("tmpl").Funcs(tmplFuncMap).Parse(outputTmpl))
		if err := templ.Execute(buf, &details); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Bad Request: "+err.Error())
		}

		return ctx.Blob(http.StatusOK, "text/csv", buf.Bytes())
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
