package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/bojand/ghz/web/database"
	"github.com/bojand/ghz/web/model"
	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
)

func TestIngestAPI(t *testing.T) {
	os.Remove(dbName)

	defer os.Remove(dbName)

	db, err := database.New("sqlite3", dbName, false)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	defer db.Close()

	api := IngestAPI{DB: db}

	var projectID uint
	var projectName string
	var pid string

	t.Run("Ingest", func(t *testing.T) {

		dat, err := ioutil.ReadFile("../test/SayHello/report1.json")
		assert.NoError(t, err)

		e := echo.New()
		req := httptest.NewRequest(http.MethodPost, "/ingest", strings.NewReader(string(dat)))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		if assert.NoError(t, api.Ingest(c)) {
			assert.Equal(t, http.StatusCreated, rec.Code)

			r := new(IngestResponse)
			err = json.NewDecoder(rec.Body).Decode(r)

			assert.NoError(t, err)

			assert.NotNil(t, r.Project)
			assert.NotZero(t, r.Project.ID)
			assert.NotZero(t, r.Project.CreatedAt)
			assert.NotZero(t, r.Project.UpdatedAt)
			assert.NotEmpty(t, r.Project.Name)

			assert.NotNil(t, r.Report)
			assert.NotZero(t, r.Report.ID)
			assert.NotZero(t, r.Report.CreatedAt)
			assert.NotZero(t, r.Report.UpdatedAt)
			assert.Equal(t, r.Project.ID, r.Report.ProjectID)
			assert.Equal(t, "Greeter SayHello", r.Report.Name)
			assert.NotZero(t, r.Report.Date)
			assert.Equal(t, uint64(200), r.Report.Count)
			assert.NotZero(t, r.Report.EndReason)
			assert.NotZero(t, r.Report.Total)
			assert.NotZero(t, r.Report.Average)
			assert.NotZero(t, r.Report.Fastest)
			assert.NotZero(t, r.Report.Slowest)
			assert.NotZero(t, r.Report.Rps)

			assert.NotNil(t, r.Options)
			assert.NotNil(t, r.Options.Info)
			assert.NotEmpty(t, r.Options.Info.Name)
			assert.NotEmpty(t, r.Options.Info.Call)

			assert.NotNil(t, r.Histogram)
			assert.NotEmpty(t, r.Histogram.Buckets)

			assert.NotNil(t, r.Details)
			assert.NotEmpty(t, r.Details)

			projectID = r.Project.ID
			pid = strconv.FormatUint(uint64(r.Project.ID), 10)
			projectName = r.Project.Name
		}
	})

	t.Run("IngestToProject", func(t *testing.T) {

		dat, err := ioutil.ReadFile("../test/SayHello/report2.json")
		assert.NoError(t, err)

		e := echo.New()
		req := httptest.NewRequest(http.MethodPost, "/projects/"+pid+"/ingest", strings.NewReader(string(dat)))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		c := e.NewContext(req, rec)
		c.SetParamNames("pid")
		c.SetParamValues(pid)

		if assert.NoError(t, api.IngestToProject(c)) {
			assert.Equal(t, http.StatusCreated, rec.Code)

			r := new(IngestResponse)
			err = json.NewDecoder(rec.Body).Decode(r)

			assert.NoError(t, err)

			assert.NotNil(t, r.Project)
			assert.NotZero(t, r.Project.ID)
			assert.Equal(t, projectID, r.Project.ID)
			assert.Equal(t, projectName, r.Project.Name)
			assert.NotZero(t, r.Project.CreatedAt)
			assert.NotZero(t, r.Project.UpdatedAt)

			assert.NotNil(t, r.Report)
			assert.NotZero(t, r.Report.ID)
			assert.NotZero(t, r.Report.CreatedAt)
			assert.NotZero(t, r.Report.UpdatedAt)
			assert.Equal(t, r.Project.ID, r.Report.ProjectID)
			assert.Equal(t, "Greeter SayHello", r.Report.Name)
			assert.NotZero(t, r.Report.Date)
			assert.Equal(t, uint64(200), r.Report.Count)
			assert.NotZero(t, r.Report.EndReason)
			assert.NotZero(t, r.Report.Total)
			assert.NotZero(t, r.Report.Average)
			assert.NotZero(t, r.Report.Fastest)
			assert.NotZero(t, r.Report.Slowest)
			assert.NotZero(t, r.Report.Rps)

			assert.NotNil(t, r.Options)
			assert.NotNil(t, r.Options.Info)
			assert.NotEmpty(t, r.Options.Info.Name)

			assert.NotNil(t, r.Histogram)
			assert.NotEmpty(t, r.Histogram.Buckets)

			assert.NotNil(t, r.Details)
			assert.NotEmpty(t, r.Details)

			assert.Equal(t, r.Report.Status, r.Project.Status)
		}
	})

	t.Run("IngestToProject 404 for unknown project", func(t *testing.T) {

		dat, err := ioutil.ReadFile("../test/SayHello/report2.json")
		assert.NoError(t, err)

		e := echo.New()
		req := httptest.NewRequest(http.MethodPost, "/projects/123321/ingest", strings.NewReader(string(dat)))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		c := e.NewContext(req, rec)
		c.SetParamNames("pid")
		c.SetParamValues("123321")

		err = api.IngestToProject(c)
		if assert.Error(t, err) {
			httpError, ok := err.(*echo.HTTPError)
			assert.True(t, ok)
			assert.Equal(t, http.StatusNotFound, httpError.Code)
		}
	})

	t.Run("IngestToProject 404 for invalid project", func(t *testing.T) {

		dat, err := ioutil.ReadFile("../test/SayHello/report2.json")
		assert.NoError(t, err)

		e := echo.New()
		req := httptest.NewRequest(http.MethodPost, "/projects/asdf/ingest", strings.NewReader(string(dat)))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		c := e.NewContext(req, rec)
		c.SetParamNames("pid")
		c.SetParamValues("asdf")

		err = api.IngestToProject(c)
		if assert.Error(t, err) {
			httpError, ok := err.(*echo.HTTPError)
			assert.True(t, ok)
			assert.Equal(t, http.StatusNotFound, httpError.Code)
		}
	})

	t.Run("IngestToProject 404 for empty project", func(t *testing.T) {

		dat, err := ioutil.ReadFile("../test/SayHello/report2.json")
		assert.NoError(t, err)

		e := echo.New()
		req := httptest.NewRequest(http.MethodPost, "/projects/ingest", strings.NewReader(string(dat)))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		c := e.NewContext(req, rec)
		c.SetParamNames("pid")
		c.SetParamValues("")

		err = api.IngestToProject(c)
		if assert.Error(t, err) {
			httpError, ok := err.(*echo.HTTPError)
			assert.True(t, ok)
			assert.Equal(t, http.StatusNotFound, httpError.Code)
		}
	})

	t.Run("IngestToProject status update", func(t *testing.T) {
		dat, err := ioutil.ReadFile("../test/SayHello/report7.json")
		assert.NoError(t, err)

		e := echo.New()
		req := httptest.NewRequest(http.MethodPost, "/projects/"+pid+"/ingest", strings.NewReader(string(dat)))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		c := e.NewContext(req, rec)
		c.SetParamNames("pid")
		c.SetParamValues(pid)

		if assert.NoError(t, api.IngestToProject(c)) {
			assert.Equal(t, http.StatusCreated, rec.Code)

			r := new(IngestResponse)
			err = json.NewDecoder(rec.Body).Decode(r)

			assert.NoError(t, err)

			assert.Equal(t, model.Status("fail"), r.Project.Status)

			// now ingest an earlier run with no errors / status OK
			dat, err = ioutil.ReadFile("../test/SayHello/report3.json")
			assert.NoError(t, err)

			e = echo.New()
			req = httptest.NewRequest(http.MethodPost, "/projects/"+pid+"/ingest", strings.NewReader(string(dat)))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()

			c := e.NewContext(req, rec)
			c.SetParamNames("pid")
			c.SetParamValues(pid)

			if assert.NoError(t, api.IngestToProject(c)) {
				assert.Equal(t, http.StatusCreated, rec.Code)

				r := new(IngestResponse)
				err = json.NewDecoder(rec.Body).Decode(r)

				assert.NoError(t, err)

				assert.Equal(t, model.Status("fail"), r.Project.Status)
			}
		}
	})
}
