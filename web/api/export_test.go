package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/bojand/ghz/runner"
	"github.com/bojand/ghz/web/database"
	"github.com/bojand/ghz/web/model"
	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
)

func TestExportAPI(t *testing.T) {
	os.Remove(dbName)

	defer os.Remove(dbName)

	db, err := database.New("sqlite3", dbName, false)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	defer db.Close()

	api := ExportAPI{DB: db}

	var rid uint

	t.Run("Create Report", func(t *testing.T) {
		p := model.Project{
			Name:        "Test Proj 111 ",
			Description: "Test Description Asdf ",
		}

		r := model.Report{
			Project:   &p,
			Name:      "Test report",
			EndReason: "normal",
			Date:      time.Date(2018, 12, 1, 1, 0, 0, 0, time.UTC),
			Count:     200,
			Total:     time.Duration(2 * time.Second),
			Average:   time.Duration(10 * time.Millisecond),
			Fastest:   time.Duration(1 * time.Millisecond),
			Slowest:   time.Duration(100 * time.Millisecond),
			Rps:       2000,
		}

		r.ErrorDist = map[string]int{
			"rpc error: code = Internal desc = Internal error.":            3,
			"rpc error: code = DeadlineExceeded desc = Deadline exceeded.": 2}

		r.StatusCodeDist = map[string]int{
			"OK":               195,
			"Internal":         3,
			"DeadlineExceeded": 2}

		r.LatencyDistribution = []*runner.LatencyDistribution{
			{
				Percentage: 25,
				Latency:    time.Duration(1 * time.Millisecond),
			},
			{
				Percentage: 50,
				Latency:    time.Duration(5 * time.Millisecond),
			},
			{
				Percentage: 75,
				Latency:    time.Duration(10 * time.Millisecond),
			},
			{
				Percentage: 90,
				Latency:    time.Duration(15 * time.Millisecond),
			},
			{
				Percentage: 95,
				Latency:    time.Duration(20 * time.Millisecond),
			},
			{
				Percentage: 99,
				Latency:    time.Duration(25 * time.Millisecond),
			},
		}

		h := model.Histogram{
			Report: &r,
			Buckets: []*runner.Bucket{
				{
					Mark:      0.01,
					Count:     1,
					Frequency: 0.005,
				},
				{
					Mark:      0.02,
					Count:     10,
					Frequency: 0.01,
				},
				{
					Mark:      0.03,
					Count:     50,
					Frequency: 0.1,
				},
				{
					Mark:      0.05,
					Count:     60,
					Frequency: 0.15,
				},
				{
					Mark:      0.1,
					Count:     15,
					Frequency: 0.07,
				},
			},
		}

		err := db.CreateHistogram(&h)

		assert.NoError(t, err)
		assert.NotZero(t, p.ID)
		assert.NotZero(t, r.ID)
		assert.NotZero(t, h.ID)

		rid = r.ID

		o := model.Options{
			ReportID: r.ID,
			Info: &model.OptionsInfo{
				Call:  "helloworld.Greeter.SayHi",
				Proto: "greeter.proto",
			},
		}

		err = db.CreateOptions(&o)

		assert.NoError(t, err)
		assert.NotZero(t, o.ID)

		M := 200
		s := make([]*model.Detail, M)

		for n := 0; n < M; n++ {
			nd := &model.Detail{
				ReportID: rid,
				ResultDetail: runner.ResultDetail{
					Timestamp: time.Now(),
					Latency:   time.Duration(1 * time.Millisecond),
					Status:    "OK",
				},
			}

			s[n] = nd
		}

		created, errored := db.CreateDetailsBatch(rid, s)

		assert.Equal(t, M, int(created))
		assert.Equal(t, 0, int(errored))
	})

	t.Run("GetExport JSON", func(t *testing.T) {
		id := strconv.FormatUint(uint64(rid), 10)

		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/"+id+"/histogram?format=json", strings.NewReader(""))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		rec := httptest.NewRecorder()

		c := e.NewContext(req, rec)
		c.SetParamNames("rid")
		c.SetParamValues(id)

		if assert.NoError(t, api.GetExport(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)

			jsonExport := new(JSONExportRespose)
			err = json.NewDecoder(rec.Body).Decode(jsonExport)

			assert.NoError(t, err)

			assert.Equal(t, rid, jsonExport.ID)
			assert.NotZero(t, jsonExport.CreatedAt)
			assert.NotZero(t, jsonExport.UpdatedAt)

			assert.NotNil(t, jsonExport.Histogram)
			assert.Len(t, jsonExport.Histogram, 5)
			assert.Equal(t, &runner.Bucket{
				Mark:      0.01,
				Count:     1,
				Frequency: 0.005,
			}, jsonExport.Histogram[0])
			assert.Equal(t, &runner.Bucket{
				Mark:      0.1,
				Count:     15,
				Frequency: 0.07,
			}, jsonExport.Histogram[4])

			assert.NotNil(t, jsonExport.Options)
			assert.Equal(t, "helloworld.Greeter.SayHi", jsonExport.Options.Call)
			assert.Equal(t, "greeter.proto", jsonExport.Options.Proto)

			assert.NotNil(t, jsonExport.Details)
			assert.Len(t, jsonExport.Details, 200)
		}
	})

	t.Run("GetExport CSV", func(t *testing.T) {
		id := strconv.FormatUint(uint64(rid), 10)

		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/"+id+"/histogram?format=csv", strings.NewReader(""))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		rec := httptest.NewRecorder()

		c := e.NewContext(req, rec)
		c.SetParamNames("rid")
		c.SetParamValues(id)

		if assert.NoError(t, api.GetExport(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
			assert.NotZero(t, len(rec.Body.String()))
		}
	})

	t.Run("GetGetExport 400 for invalid format", func(t *testing.T) {
		id := strconv.FormatUint(uint64(rid), 10)

		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/"+id+"/histogram?format=asdf", strings.NewReader(""))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		rec := httptest.NewRecorder()

		c := e.NewContext(req, rec)
		c.SetParamNames("rid")
		c.SetParamValues(id)

		err := api.GetExport(c)
		if assert.Error(t, err) {
			httpError, ok := err.(*echo.HTTPError)
			assert.True(t, ok)
			assert.Equal(t, http.StatusBadRequest, httpError.Code)
		}
	})

	t.Run("GetGetExport 404 for unknown", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/12332198/export?format=json", strings.NewReader(""))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		rec := httptest.NewRecorder()

		c := e.NewContext(req, rec)
		c.SetParamNames("rid")
		c.SetParamValues("12332198")

		err := api.GetExport(c)
		if assert.Error(t, err) {
			httpError, ok := err.(*echo.HTTPError)
			assert.True(t, ok)
			assert.Equal(t, http.StatusNotFound, httpError.Code)
		}
	})

	t.Run("GetOptions 404 for invalid", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/asdf/export?format=json", strings.NewReader(""))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		rec := httptest.NewRecorder()

		c := e.NewContext(req, rec)
		c.SetParamNames("rid")
		c.SetParamValues("asdf")

		err := api.GetExport(c)
		if assert.Error(t, err) {
			httpError, ok := err.(*echo.HTTPError)
			assert.True(t, ok)
			assert.Equal(t, http.StatusNotFound, httpError.Code)
		}
	})

	t.Run("GetOptions 404 for empty", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/export?format=json", strings.NewReader(""))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		rec := httptest.NewRecorder()

		c := e.NewContext(req, rec)

		err := api.GetExport(c)
		if assert.Error(t, err) {
			httpError, ok := err.(*echo.HTTPError)
			assert.True(t, ok)
			assert.Equal(t, http.StatusNotFound, httpError.Code)
		}
	})
}
