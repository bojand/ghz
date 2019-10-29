package api

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/bojand/ghz/web/model"
	"github.com/labstack/echo"
)

// ReportDatabase interface for encapsulating database access.
type ReportDatabase interface {
	CountReports() (uint, error)
	CountReportsForProject(uint) (uint, error)
	FindReportByID(uint) (*model.Report, error)
	FindPreviousReport(uint) (*model.Report, error)
	DeleteReport(*model.Report) error
	DeleteReportBulk([]uint) (int, error)
	ListReports(limit, page uint, sortField, order string) ([]*model.Report, error)
	ListReportsForProject(pid, limit, page uint, sortField, order string) ([]*model.Report, error)
}

// The ReportAPI provides handlers for managing reports.
type ReportAPI struct {
	DB ReportDatabase
}

// ReportList response
type ReportList struct {
	Total uint            `json:"total"`
	Data  []*model.Report `json:"data"`
}

// DeleteReportBulkRequest is the request to delete bulk reports
type DeleteReportBulkRequest struct {
	IDs []uint `json:"ids"`
}

// ListReportsForProject lists reports for a project
func (api *ReportAPI) ListReportsForProject(ctx echo.Context) error {
	var projectID uint64
	var err error

	pid := ctx.Param("pid")
	if pid == "" {
		return echo.NewHTTPError(http.StatusNotFound, "")
	}

	if projectID, err = strconv.ParseUint(pid, 10, 32); err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}

	return api.listReports(true, uint(projectID), ctx)
}

// ListReportsAll gets a list of all reports
func (api *ReportAPI) ListReportsAll(ctx echo.Context) error {
	return api.listReports(false, 0, ctx)
}

func (api *ReportAPI) listReports(forProject bool, projectID uint, ctx echo.Context) error {
	var page uint64
	var sort, order string
	var err error

	if page, err = strconv.ParseUint(ctx.QueryParam("page"), 10, 32); err != nil {
		page = 0
	}

	if sort = ctx.QueryParam("sort"); sort == "" {
		sort = "id"
	}

	if order = ctx.QueryParam("order"); order == "" {
		order = "desc"
	}

	sort = strings.ToLower(sort)
	order = strings.ToLower(order)

	limit := uint(20)

	countCh := make(chan uint, 1)
	dataCh := make(chan []*model.Report, 1)
	errCh := make(chan error, 2)
	defer close(errCh)

	go func() {
		var count uint
		if forProject {
			count, err = api.DB.CountReportsForProject(projectID)
		} else {
			count, err = api.DB.CountReports()
		}
		errCh <- err
		countCh <- count
		close(countCh)
	}()

	go func() {
		var reports []*model.Report
		var err error
		if forProject {
			reports, err = api.DB.ListReportsForProject(uint(projectID), limit, uint(page), sort, order)
		} else {
			reports, err = api.DB.ListReports(limit, uint(page), sort, order)
		}
		errCh <- err
		dataCh <- reports
		close(dataCh)
	}()

	count, data, err1, err2 := <-countCh, <-dataCh, <-errCh, <-errCh

	if err1 != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Bad Request: "+err1.Error())
	}

	if err2 != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Bad Request: "+err2.Error())
	}

	rl := &ReportList{Total: count, Data: data}

	return ctx.JSON(http.StatusOK, rl)
}

// GetReport gets a report
func (api *ReportAPI) GetReport(ctx echo.Context) error {
	var id uint64
	var report *model.Report
	var err error

	if id, err = getReportID(ctx); err != nil {
		return err
	}

	if report, err = api.DB.FindReportByID(uint(id)); err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}

	return ctx.JSON(http.StatusOK, report)
}

// DeleteReport deletes a report
func (api *ReportAPI) DeleteReport(ctx echo.Context) error {
	var id uint64
	var report *model.Report
	var err error

	if id, err = getReportID(ctx); err != nil {
		return err
	}

	if report, err = api.DB.FindReportByID(uint(id)); err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}

	err = api.DB.DeleteReport(report)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return ctx.JSON(http.StatusOK, report)
}

// DeleteReportBulk deletes bulk reports
func (api *ReportAPI) DeleteReportBulk(ctx echo.Context) error {
	del := new(DeleteReportBulkRequest)
	var err error

	if err := ctx.Bind(del); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if ctx.Echo().Validator != nil {
		if err := ctx.Validate(del); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
	}

	n, err := api.DB.DeleteReportBulk(del.IDs)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	m := make(map[string]int)
	m["deleted"] = n

	return ctx.JSON(http.StatusOK, m)
}

// GetPreviousReport gets a previous report
func (api *ReportAPI) GetPreviousReport(ctx echo.Context) error {
	var id uint64
	var report *model.Report
	var err error

	if id, err = getReportID(ctx); err != nil {
		return err
	}

	if report, err = api.DB.FindPreviousReport(uint(id)); err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}

	return ctx.JSON(http.StatusOK, report)
}

func getReportID(ctx echo.Context) (uint64, error) {
	var id uint64
	var err error

	rid := ctx.Param("rid")
	if rid == "" {
		return 0, echo.NewHTTPError(http.StatusNotFound, "")
	}

	if id, err = strconv.ParseUint(rid, 10, 32); err != nil {
		return 0, echo.NewHTTPError(http.StatusNotFound, err.Error())
	}

	return id, err
}
