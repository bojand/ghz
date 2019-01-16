package api

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/bojand/ghz/web/model"
	"github.com/labstack/echo"
)

// ProjectDatabase interface for encapsulating database access.
type ProjectDatabase interface {
	CreateProject(project *model.Project) error
	FindProjectByID(id uint) (*model.Project, error)
	UpdateProject(*model.Project) error
	DeleteProject(*model.Project) error
	CountProjects() (uint, error)
	ListProjects(limit, page uint, sortField, order string) ([]*model.Project, error)
}

// The ProjectAPI provides handlers for managing projects.
type ProjectAPI struct {
	DB ProjectDatabase
}

// ProjectList response
type ProjectList struct {
	Total uint             `json:"total"`
	Data  []*model.Project `json:"data"`
}

// CreateProject creates a project
func (api *ProjectAPI) CreateProject(ctx echo.Context) error {
	p := new(model.Project)

	if err := api.bindAndValidate(ctx, p); err != nil {
		return err
	}

	if err := api.DB.CreateProject(p); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return ctx.JSON(http.StatusCreated, p)
}

// UpdateProject updates a project
func (api *ProjectAPI) UpdateProject(ctx echo.Context) error {
	var project *model.Project
	var err error

	if project, err = findProject(api.DB.FindProjectByID, ctx); err != nil {
		return err
	}

	newVal := new(model.Project)

	if err := api.bindAndValidate(ctx, newVal); err != nil {
		return err
	}

	project.Name = newVal.Name
	project.Description = newVal.Description

	err = api.DB.UpdateProject(project)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return ctx.JSON(http.StatusOK, project)
}

// GetProject gets a project
func (api *ProjectAPI) GetProject(ctx echo.Context) error {
	var project *model.Project
	var err error

	if project, err = findProject(api.DB.FindProjectByID, ctx); err != nil {
		return err
	}

	return ctx.JSON(http.StatusOK, project)
}

// ListProjects lists projects
func (api *ProjectAPI) ListProjects(ctx echo.Context) error {
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
	dataCh := make(chan []*model.Project, 1)
	errCh := make(chan error, 2)
	defer close(errCh)

	go func() {
		count, err := api.DB.CountProjects()
		errCh <- err
		countCh <- count
		close(countCh)
	}()

	go func() {
		var projects []*model.Project
		var err error
		projects, err = api.DB.ListProjects(limit, uint(page), sort, order)
		errCh <- err
		dataCh <- projects
		close(dataCh)
	}()

	count, data, err1, err2 := <-countCh, <-dataCh, <-errCh, <-errCh

	if err1 != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Bad Request: "+err1.Error())
	}

	if err2 != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Bad Request: "+err2.Error())
	}

	pl := &ProjectList{Total: count, Data: data}

	return ctx.JSON(http.StatusOK, pl)
}

// DeleteProject deletes a project
func (api *ProjectAPI) DeleteProject(ctx echo.Context) error {
	var project *model.Project
	var err error

	if project, err = findProject(api.DB.FindProjectByID, ctx); err != nil {
		return err
	}

	err = api.DB.DeleteProject(project)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return ctx.JSON(http.StatusOK, project)
}

func findProject(FindProjectByID func(id uint) (*model.Project, error), ctx echo.Context) (*model.Project, error) {
	var id uint64
	var project *model.Project
	var err error

	pid := ctx.Param("pid")
	if pid == "" {
		return nil, echo.NewHTTPError(http.StatusNotFound, "")
	}

	if id, err = strconv.ParseUint(pid, 10, 32); err != nil {
		return nil, echo.NewHTTPError(http.StatusNotFound, err.Error())
	}

	if project, err = FindProjectByID(uint(id)); err != nil {
		return nil, echo.NewHTTPError(http.StatusNotFound, err.Error())
	}

	return project, err
}

func (api *ProjectAPI) bindAndValidate(ctx echo.Context, p *model.Project) error {
	if err := ctx.Bind(p); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if ctx.Echo().Validator != nil {
		if err := ctx.Validate(p); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
	}

	return nil
}
