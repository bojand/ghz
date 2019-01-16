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

func TestDeleteAPI(t *testing.T) {
	os.Remove(dbName)

	defer os.Remove(dbName)

	db, err := database.New("sqlite3", dbName, false)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	defer db.Close()

	api := IngestAPI{DB: db}
	projectAPI := ProjectAPI{DB: db}
	reportAPI := ReportAPI{DB: db}

	var reportID, projectID string
	var pid uint
	var rid, rid2 uint
	var oid, oid2 uint

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

			assert.NotNil(t, r.Histogram)
			assert.NotEmpty(t, r.Histogram.Buckets)

			assert.NotNil(t, r.Details)
			assert.NotEmpty(t, r.Details)

			pid = r.Project.ID
			rid = r.Report.ID
			projectID = strconv.FormatUint(uint64(r.Project.ID), 10)
			reportID = strconv.FormatUint(uint64(r.Report.ID), 10)
			oid = r.Options.ID
		}
	})

	t.Run("IngestToProject", func(t *testing.T) {

		dat, err := ioutil.ReadFile("../test/SayHello/report2.json")
		assert.NoError(t, err)

		e := echo.New()
		req := httptest.NewRequest(http.MethodPost, "/projects/"+projectID+"/ingest", strings.NewReader(string(dat)))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		c := e.NewContext(req, rec)
		c.SetParamNames("pid")
		c.SetParamValues(projectID)

		if assert.NoError(t, api.IngestToProject(c)) {
			assert.Equal(t, http.StatusCreated, rec.Code)

			r := new(IngestResponse)
			err = json.NewDecoder(rec.Body).Decode(r)

			assert.NoError(t, err)

			assert.NotNil(t, r.Project)
			assert.NotZero(t, r.Project.ID)
			assert.Equal(t, pid, r.Project.ID)
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

			rid2 = r.Report.ID
			oid2 = r.Options.ID
		}
	})

	t.Run("DeleteReport 404 for unknown id", func(t *testing.T) {

		e := echo.New()
		req := httptest.NewRequest(http.MethodDelete, "/reports/123321", strings.NewReader(""))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		c := e.NewContext(req, rec)
		c.SetParamNames("rid")
		c.SetParamValues("123321")

		err := reportAPI.DeleteReport(c)
		if assert.Error(t, err) {
			httpError, ok := err.(*echo.HTTPError)
			assert.True(t, ok)
			assert.Equal(t, http.StatusNotFound, httpError.Code)
		}
	})

	t.Run("DeleteReport()", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodDelete, "/reports/"+reportID, strings.NewReader(""))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		c := e.NewContext(req, rec)
		c.SetParamNames("rid")
		c.SetParamValues(reportID)

		err := reportAPI.DeleteReport(c)
		if assert.NoError(t, err) {
			assert.Equal(t, http.StatusOK, rec.Code)

			r := new(model.Report)
			err = db.DB.First(r, rid).Error

			assert.Error(t, err)

			o := new(model.Options)
			err = db.DB.First(o, oid).Error

			assert.Error(t, err)

			p := new(model.Project)
			err = db.DB.First(p, pid).Error

			assert.NoError(t, err)
			assert.Equal(t, pid, p.ID)

			details, err := db.ListAllDetailsForReport(rid)

			assert.NoError(t, err)
			assert.Len(t, details, 0)

			details, err = db.ListAllDetailsForReport(rid2)

			assert.NoError(t, err)
			assert.True(t, len(details) > 0)
		}
	})

	t.Run("DeleteProject() 404 for unknown id", func(t *testing.T) {

		e := echo.New()
		req := httptest.NewRequest(http.MethodDelete, "/projects/123321", strings.NewReader(""))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		c := e.NewContext(req, rec)
		c.SetParamNames("pid")
		c.SetParamValues("123321")

		err := projectAPI.DeleteProject(c)
		if assert.Error(t, err) {
			httpError, ok := err.(*echo.HTTPError)
			assert.True(t, ok)
			assert.Equal(t, http.StatusNotFound, httpError.Code)
		}
	})

	t.Run("DeleteProject()", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodDelete, "/projects/"+projectID, strings.NewReader(""))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		c := e.NewContext(req, rec)
		c.SetParamNames("pid")
		c.SetParamValues(projectID)

		err := projectAPI.DeleteProject(c)
		if assert.NoError(t, err) {
			assert.Equal(t, http.StatusOK, rec.Code)

			r := new(model.Report)
			err = db.DB.First(r, rid2).Error

			assert.Error(t, err)

			o := new(model.Options)
			err = db.DB.First(o, oid2).Error

			assert.Error(t, err)

			p := new(model.Project)
			err = db.DB.First(p, pid).Error

			assert.Error(t, err)

			details, err := db.ListAllDetailsForReport(rid)

			assert.NoError(t, err)
			assert.Len(t, details, 0)

			details, err = db.ListAllDetailsForReport(rid2)

			assert.NoError(t, err)
			assert.Len(t, details, 0)
		}
	})
}
