package api

import (
	"net/http"

	"github.com/bojand/ghz/web/model"
	"github.com/labstack/echo"
)

// HistogramDatabase interface for encapsulating database access.
type HistogramDatabase interface {
	GetHistogramForReport(uint) (*model.Histogram, error)
}

// The HistogramAPI provides handlers.
type HistogramAPI struct {
	DB HistogramDatabase
}

// GetHistogram gets a histogram for the report
func (api *HistogramAPI) GetHistogram(ctx echo.Context) error {
	var id uint64
	var h *model.Histogram
	var err error

	if id, err = getReportID(ctx); err != nil {
		return err
	}

	if h, err = api.DB.GetHistogramForReport(uint(id)); err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}

	return ctx.JSON(http.StatusOK, h)
}
