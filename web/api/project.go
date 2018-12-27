package api

import (
	"net/http"

	"github.com/bojand/ghz/web/model"
	"github.com/labstack/echo"
)

// ProjectDatabase interface for encapsulating database access.
type ProjectDatabase interface {
	CreateProject(project *model.Project) error
	FindProjectByID(id uint) (*model.Project, error)
	UpdateProject(project *model.Project) error
}

// The ProjectAPI provides handlers for managing projects.
type ProjectAPI struct {
	DB ProjectDatabase
}

// CreateProject creates a project
func (api *ProjectAPI) CreateProject(ctx echo.Context) error {
	p := new(model.Project)

	if err := ctx.Bind(p); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := ctx.Validate(p); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	err := api.DB.CreateProject(p)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return ctx.JSON(http.StatusCreated, p)

}
