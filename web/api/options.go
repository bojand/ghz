package api

import (
	"net/http"

	"github.com/bojand/ghz/web/model"
	"github.com/labstack/echo"
)

// OptionsDatabase interface for encapsulating database access.
type OptionsDatabase interface {
	GetOptionsForReport(uint) (*model.Options, error)
}

// The OptionsAPI provides handlers
type OptionsAPI struct {
	DB OptionsDatabase
}

// GetOptions gets options for a report
func (api *OptionsAPI) GetOptions(ctx echo.Context) error {
	var id uint64
	var o *model.Options
	var err error

	if id, err = getReportID(ctx); err != nil {
		return err
	}

	if o, err = api.DB.GetOptionsForReport(uint(id)); err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}

	return ctx.JSON(http.StatusOK, o)
}
