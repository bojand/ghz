package api

import (
	"net/http"
	"strconv"

	"github.com/bojand/ghz/web/model"
	"github.com/labstack/echo"
)

// LatencyDatabase interface for encapsulating database access.
type LatencyDatabase interface {
	GetLatencyDistributionForReport(uint) (*model.LatencyDistribution, error)
}

// The LatencyAPI provides handlers for managing projects.
type LatencyAPI struct {
	DB LatencyDatabase
}

// GetLatencyDistribution gets a project
func (api *LatencyAPI) GetLatencyDistribution(ctx echo.Context) error {
	var id uint64
	var ld *model.LatencyDistribution
	var err error

	rid := ctx.Param("rid")
	if rid == "" {
		return echo.NewHTTPError(http.StatusNotFound, "")
	}

	if id, err = strconv.ParseUint(rid, 10, 32); err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}

	if ld, err = api.DB.GetLatencyDistributionForReport(uint(id)); err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}

	return ctx.JSON(http.StatusOK, ld)
}
