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

func TestHistogramAPI(t *testing.T) {
	os.Remove(dbName)

	defer os.Remove(dbName)

	db, err := database.New("sqlite3", dbName, false)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	defer db.Close()

	api := HistogramAPI{DB: db}

	var rid, hid uint

	t.Run("Create Histogram", func(t *testing.T) {
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
		hid = h.ID
	})

	t.Run("GetHistogram", func(t *testing.T) {
		id := strconv.FormatUint(uint64(rid), 10)

		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/"+id+"/histogram", strings.NewReader(""))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		rec := httptest.NewRecorder()

		c := e.NewContext(req, rec)
		c.SetParamNames("rid")
		c.SetParamValues(id)

		if assert.NoError(t, api.GetHistogram(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)

			h := new(model.Histogram)
			err = json.NewDecoder(rec.Body).Decode(h)

			assert.NoError(t, err)

			assert.Equal(t, rid, h.ReportID)
			assert.Equal(t, hid, h.ID)
			assert.Nil(t, h.Report)
			assert.NotZero(t, h.CreatedAt)
			assert.NotZero(t, h.UpdatedAt)

			assert.NotNil(t, h.Buckets)
			assert.Len(t, h.Buckets, 5)
			assert.Equal(t, &runner.Bucket{
				Mark:      0.01,
				Count:     1,
				Frequency: 0.005,
			}, h.Buckets[0])
			assert.Equal(t, &runner.Bucket{
				Mark:      0.1,
				Count:     15,
				Frequency: 0.07,
			}, h.Buckets[4])
		}
	})

	t.Run("GetOptions 404 for unknown", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/12332198/histogram", strings.NewReader(""))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		rec := httptest.NewRecorder()

		c := e.NewContext(req, rec)
		c.SetParamNames("rid")
		c.SetParamValues("12332198")

		err := api.GetHistogram(c)
		if assert.Error(t, err) {
			httpError, ok := err.(*echo.HTTPError)
			assert.True(t, ok)
			assert.Equal(t, http.StatusNotFound, httpError.Code)
		}
	})

	t.Run("GetOptions 404 for invalid", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/asdf/histogram", strings.NewReader(""))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		rec := httptest.NewRecorder()

		c := e.NewContext(req, rec)
		c.SetParamNames("rid")
		c.SetParamValues("asdf")

		err := api.GetHistogram(c)
		if assert.Error(t, err) {
			httpError, ok := err.(*echo.HTTPError)
			assert.True(t, ok)
			assert.Equal(t, http.StatusNotFound, httpError.Code)
		}
	})

	t.Run("GetOptions 404 for empty", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/histogram", strings.NewReader(""))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		rec := httptest.NewRecorder()

		c := e.NewContext(req, rec)

		err := api.GetHistogram(c)
		if assert.Error(t, err) {
			httpError, ok := err.(*echo.HTTPError)
			assert.True(t, ok)
			assert.Equal(t, http.StatusNotFound, httpError.Code)
		}
	})
}
