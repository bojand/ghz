package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"

	"github.com/bojand/ghz/web/database"
	"github.com/bojand/ghz/web/model"
)

const dbName = "../test/api_test.db"

func TestProjectAPI(t *testing.T) {
	os.Remove(dbName)

	defer os.Remove(dbName)

	db, err := database.New("sqlite3", dbName, false)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	defer db.Close()

	api := ProjectAPI{DB: db}

	var projectID uint
	var pid string

	t.Run("CreateProject", func(t *testing.T) {
		pJSON := `{"name":"Example Project","description":"Lorem Ipsum project description"}`

		e := echo.New()
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(pJSON))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		if assert.NoError(t, api.CreateProject(c)) {
			assert.Equal(t, http.StatusCreated, rec.Code)

			p := new(model.Project)
			err = json.NewDecoder(rec.Body).Decode(p)

			assert.NoError(t, err)

			assert.NotZero(t, p.ID)
			assert.NotZero(t, p.CreatedAt)
			assert.NotZero(t, p.UpdatedAt)
			assert.Equal(t, "Example Project", p.Name)
			assert.Equal(t, "Lorem Ipsum project description", p.Description)
			assert.Equal(t, model.StatusOK, p.Status)

			projectID = p.ID
			pid = strconv.FormatUint(uint64(projectID), 10)
		}
	})

	t.Run("CreateProject empty", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("{}"))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		if assert.NoError(t, api.CreateProject(c)) {
			assert.Equal(t, http.StatusCreated, rec.Code)

			p := new(model.Project)
			err = json.NewDecoder(rec.Body).Decode(p)

			assert.NoError(t, err)

			assert.NotZero(t, p.ID)
			assert.NotZero(t, p.CreatedAt)
			assert.NotZero(t, p.UpdatedAt)
			assert.NotEmpty(t, p.Name)
			assert.Equal(t, "", p.Description)
		}
	})

	t.Run("CreateProject with just description", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"description":"Some description for a project"}`))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		if assert.NoError(t, api.CreateProject(c)) {
			assert.Equal(t, http.StatusCreated, rec.Code)

			p := new(model.Project)
			err = json.NewDecoder(rec.Body).Decode(p)

			assert.NoError(t, err)

			assert.NotZero(t, p.ID)
			assert.NotZero(t, p.CreatedAt)
			assert.NotZero(t, p.UpdatedAt)
			assert.NotEmpty(t, p.Name)
			assert.Equal(t, "Some description for a project", p.Description)
		}
	})

	t.Run("GetProject", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/"+pid, strings.NewReader(""))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		rec := httptest.NewRecorder()

		c := e.NewContext(req, rec)
		c.SetParamNames("pid")
		c.SetParamValues(pid)

		if assert.NoError(t, api.GetProject(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)

			p := new(model.Project)
			err = json.NewDecoder(rec.Body).Decode(p)

			assert.NoError(t, err)

			assert.NotZero(t, p.ID)
			assert.NotZero(t, p.CreatedAt)
			assert.NotZero(t, p.UpdatedAt)
			assert.Equal(t, "Example Project", p.Name)
			assert.Equal(t, "Lorem Ipsum project description", p.Description)
		}
	})

	t.Run("GetProject 404 for unknown", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/12332198", strings.NewReader(""))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		rec := httptest.NewRecorder()

		c := e.NewContext(req, rec)
		c.SetParamNames("pid")
		c.SetParamValues("12332198")

		err := api.GetProject(c)
		if assert.Error(t, err) {
			httpError, ok := err.(*echo.HTTPError)
			assert.True(t, ok)
			assert.Equal(t, http.StatusNotFound, httpError.Code)
		}
	})

	t.Run("GetProject 404 for invalid", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/asdf", strings.NewReader(""))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		rec := httptest.NewRecorder()

		c := e.NewContext(req, rec)
		c.SetParamNames("pid")
		c.SetParamValues("asdf")

		err := api.GetProject(c)
		if assert.Error(t, err) {
			httpError, ok := err.(*echo.HTTPError)
			assert.True(t, ok)
			assert.Equal(t, http.StatusNotFound, httpError.Code)
		}
	})

	t.Run("GetProject 404 for empty pid", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/", strings.NewReader(""))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		rec := httptest.NewRecorder()

		c := e.NewContext(req, rec)
		c.SetParamNames("pid")
		c.SetParamValues("")

		err := api.GetProject(c)
		if assert.Error(t, err) {
			httpError, ok := err.(*echo.HTTPError)
			assert.True(t, ok)
			assert.Equal(t, http.StatusNotFound, httpError.Code)
		}
	})

	t.Run("UpdateProject 404 for unknown", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodPut, "/12332198", strings.NewReader("{}"))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		rec := httptest.NewRecorder()

		c := e.NewContext(req, rec)
		c.SetParamNames("pid")
		c.SetParamValues("12332198")

		err := api.GetProject(c)
		if assert.Error(t, err) {
			httpError, ok := err.(*echo.HTTPError)
			assert.True(t, ok)
			assert.Equal(t, http.StatusNotFound, httpError.Code)
		}
	})

	t.Run("UpdateProject", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodPut, "/"+pid, strings.NewReader(`{"status":"fail","name":"Updated name","description":"Updated desc"}`))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		rec := httptest.NewRecorder()

		c := e.NewContext(req, rec)
		c.SetParamNames("pid")
		c.SetParamValues(pid)

		if assert.NoError(t, api.UpdateProject(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)

			p := new(model.Project)
			err = json.NewDecoder(rec.Body).Decode(p)

			assert.NoError(t, err)

			assert.NotZero(t, p.ID)
			assert.NotZero(t, p.CreatedAt)
			assert.NotZero(t, p.UpdatedAt)
			assert.Equal(t, "Updated name", p.Name)
			assert.Equal(t, "Updated desc", p.Description)
			assert.Equal(t, model.StatusOK, p.Status) // we can only update name and desc
		}
	})

	t.Run("ListProjects", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/", strings.NewReader(`{}`))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		rec := httptest.NewRecorder()

		c := e.NewContext(req, rec)

		if assert.NoError(t, api.ListProjects(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)

			pl := new(ProjectList)
			err = json.NewDecoder(rec.Body).Decode(pl)

			assert.NoError(t, err)
			assert.Equal(t, uint(3), pl.Total)
			assert.Len(t, pl.Data, 3)
			assert.NotZero(t, pl.Data[0].ID)
			assert.NotEmpty(t, pl.Data[0].Name)
			assert.NotZero(t, pl.Data[1].ID)
			assert.NotEmpty(t, pl.Data[1].Name)
			assert.NotZero(t, pl.Data[2].ID)
			assert.NotEmpty(t, pl.Data[2].Name)

			// by default we sort by desc id
			// so the last one should be the first one
			assert.Equal(t, projectID, pl.Data[2].ID)
		}
	})

	t.Run("ListProjects sorted", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/?sort=id&order=asc", strings.NewReader(`{}`))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		rec := httptest.NewRecorder()

		c := e.NewContext(req, rec)

		if assert.NoError(t, api.ListProjects(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)

			pl := new(ProjectList)
			err = json.NewDecoder(rec.Body).Decode(pl)

			assert.NoError(t, err)
			assert.Equal(t, uint(3), pl.Total)
			assert.Len(t, pl.Data, 3)
			assert.NotZero(t, pl.Data[0].ID)
			assert.NotEmpty(t, pl.Data[0].Name)
			assert.NotZero(t, pl.Data[1].ID)
			assert.NotEmpty(t, pl.Data[1].Name)
			assert.NotZero(t, pl.Data[2].ID)
			assert.NotEmpty(t, pl.Data[2].Name)

			assert.Equal(t, projectID, pl.Data[0].ID)
		}
	})
}
