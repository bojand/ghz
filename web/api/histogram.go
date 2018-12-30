package api

import (
	"net/http"
	"strconv"

	"github.com/bojand/ghz/web/model"
	"github.com/labstack/echo"
)

// HistogramDatabase interface for encapsulating database access.
type HistogramDatabase interface {
	GetHistogramForReport(uint) (*model.Histogram, error)
}

// The HistogramAPI provides handlers for managing projects.
type HistogramAPI struct {
	DB HistogramDatabase
}

// GetHistogram gets a project
func (api *HistogramAPI) GetHistogram(ctx echo.Context) error {
	var id uint64
	var h *model.Histogram
	var err error

	rid := ctx.Param("rid")
	if rid == "" {
		return echo.NewHTTPError(http.StatusNotFound, "")
	}

	if id, err = strconv.ParseUint(rid, 10, 32); err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}

	if h, err = api.DB.GetHistogramForReport(uint(id)); err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}

	return ctx.JSON(http.StatusOK, h)
}
