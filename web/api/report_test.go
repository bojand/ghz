package api

import (
	"encoding/json"
	"fmt"
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

func TestReportAPI(t *testing.T) {
	os.Remove(dbName)

	defer os.Remove(dbName)

	db, err := database.New("sqlite3", dbName, false)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	defer db.Close()

	api := ReportAPI{DB: db}

	var projectID, projectID2 uint
	var pid, pid2 string
	var reportID uint
	var rid string
	var reportID9, reportID10, p2reportID5, p2reportID6, p2ridLatest uint
	var rid10 string

	t.Run("Create Reports", func(t *testing.T) {
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

		err := db.CreateReport(&r)

		assert.NoError(t, err)
		assert.NotZero(t, p.ID)
		assert.NotZero(t, r.ID)

		projectID = p.ID
		pid = strconv.FormatUint(uint64(projectID), 10)
		reportID = r.ID
		rid = strconv.FormatUint(uint64(reportID), 10)

		N := 20

		for n := 0; n < N; n++ {
			r := model.Report{
				ProjectID: projectID,
				Name:      "Test report " + strconv.FormatUint(uint64(n), 10),
				EndReason: "normal",
				Date:      time.Date(2018, 12, 1, 2+n, 0, 0, 0, time.UTC),
				Count:     uint64(200) + uint64(n),
				Total:     time.Duration(time.Duration(20+n) * time.Second),
				Average:   time.Duration(time.Duration(10+n) * time.Millisecond),
				Fastest:   time.Duration(time.Duration(1+n) * time.Millisecond),
				Slowest:   time.Duration(time.Duration(100+n) * time.Millisecond),
				Rps:       float64(2000 + n),
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

			err := db.CreateReport(&r)

			assert.NoError(t, err)
			assert.NotZero(t, p.ID)
			assert.NotZero(t, r.ID)

			if n == 9 {
				reportID9 = r.ID
			}

			if n == 10 {
				reportID10 = r.ID
				rid10 = strconv.FormatUint(uint64(reportID10), 10)
			}
		}
	})

	t.Run("Create Reports Project 2", func(t *testing.T) {
		p := model.Project{
			Name:        "Test Proj 222 ",
			Description: "Test Description 2 ",
		}

		r := model.Report{
			Project:   &p,
			Name:      "Test report",
			EndReason: "normal",
			Date:      time.Date(2018, 12, 1, 1, 30, 0, 0, time.UTC),
			Count:     300,
			Total:     time.Duration(3 * time.Second),
			Average:   time.Duration(30 * time.Millisecond),
			Fastest:   time.Duration(3 * time.Millisecond),
			Slowest:   time.Duration(300 * time.Millisecond),
			Rps:       3000,
		}

		r.ErrorDist = map[string]int{
			"rpc error: code = Internal desc = Internal error.":            1,
			"rpc error: code = DeadlineExceeded desc = Deadline exceeded.": 4}

		r.StatusCodeDist = map[string]int{
			"OK":               195,
			"Internal":         1,
			"DeadlineExceeded": 4}

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

		err := db.CreateReport(&r)

		assert.NoError(t, err)
		assert.NotZero(t, p.ID)
		assert.NotZero(t, r.ID)

		projectID2 = p.ID
		pid2 = strconv.FormatUint(uint64(projectID2), 10)

		N := 10

		for n := 0; n < N; n++ {
			r := model.Report{
				ProjectID: projectID2,
				Name:      "Test report " + strconv.FormatUint(uint64(n), 10),
				EndReason: "normal",
				Date:      time.Date(2018, 12, 1, 2+n, 30, 0, 0, time.UTC),
				Count:     uint64(300) + uint64(n),
				Total:     time.Duration(time.Duration(30+n) * time.Second),
				Average:   time.Duration(time.Duration(30+n) * time.Millisecond),
				Fastest:   time.Duration(time.Duration(3+n) * time.Millisecond),
				Slowest:   time.Duration(time.Duration(300+n) * time.Millisecond),
				Rps:       float64(3000 + n),
			}

			r.ErrorDist = map[string]int{
				"rpc error: code = Internal desc = Internal error.":            1,
				"rpc error: code = DeadlineExceeded desc = Deadline exceeded.": 4}

			r.StatusCodeDist = map[string]int{
				"OK":               195,
				"Internal":         1,
				"DeadlineExceeded": 4}

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

			err := db.CreateReport(&r)

			assert.NoError(t, err)
			assert.NotZero(t, p.ID)
			assert.NotZero(t, r.ID)

			if n == 5 {
				p2reportID5 = r.ID
			}

			if n == 6 {
				p2reportID6 = r.ID
			}

			if n == 9 {
				p2ridLatest = r.ID
			}
		}
	})

	t.Run("GetReport", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/"+rid, strings.NewReader(""))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		rec := httptest.NewRecorder()

		c := e.NewContext(req, rec)
		c.SetParamNames("rid")
		c.SetParamValues(rid)

		if assert.NoError(t, api.GetReport(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)

			r := new(model.Report)
			err = json.NewDecoder(rec.Body).Decode(r)

			assert.NoError(t, err)

			assert.NotZero(t, r.ID)
			assert.NotZero(t, r.CreatedAt)
			assert.NotZero(t, r.UpdatedAt)
			assert.Equal(t, projectID, r.ProjectID)
			assert.Nil(t, r.Project)
			assert.Equal(t, "Test report", r.Name)
			assert.Equal(t, uint64(200), r.Count)
		}
	})

	t.Run("GetReport 404 for unknown", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/12332198", strings.NewReader(""))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		rec := httptest.NewRecorder()

		c := e.NewContext(req, rec)
		c.SetParamNames("rid")
		c.SetParamValues("12332198")

		err := api.GetReport(c)
		if assert.Error(t, err) {
			httpError, ok := err.(*echo.HTTPError)
			assert.True(t, ok)
			assert.Equal(t, http.StatusNotFound, httpError.Code)
		}
	})

	t.Run("GetReport 404 for invalid rid", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/asdf", strings.NewReader(""))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		rec := httptest.NewRecorder()

		c := e.NewContext(req, rec)
		c.SetParamNames("rid")
		c.SetParamValues("asdf")

		err := api.GetReport(c)
		if assert.Error(t, err) {
			httpError, ok := err.(*echo.HTTPError)
			assert.True(t, ok)
			assert.Equal(t, http.StatusNotFound, httpError.Code)
		}
	})

	t.Run("GetReport 404 for empty rid", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/", strings.NewReader(""))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		rec := httptest.NewRecorder()

		c := e.NewContext(req, rec)

		err := api.GetReport(c)
		if assert.Error(t, err) {
			httpError, ok := err.(*echo.HTTPError)
			assert.True(t, ok)
			assert.Equal(t, http.StatusNotFound, httpError.Code)
		}
	})

	t.Run("GetPreviousReport", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/"+rid10+"/previous", strings.NewReader(""))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		rec := httptest.NewRecorder()

		c := e.NewContext(req, rec)
		c.SetParamNames("rid")
		c.SetParamValues(rid10)

		if assert.NoError(t, api.GetPreviousReport(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)

			r := new(model.Report)
			err = json.NewDecoder(rec.Body).Decode(r)

			assert.NoError(t, err)

			assert.NotZero(t, r.ID)
			assert.Equal(t, reportID9, r.ID)
			assert.NotZero(t, r.CreatedAt)
			assert.NotZero(t, r.UpdatedAt)
			assert.Equal(t, projectID, r.ProjectID)
			assert.Nil(t, r.Project)
			assert.Equal(t, "Test report 9", r.Name)
			assert.Equal(t, uint64(209), r.Count)
		}
	})

	t.Run("GetPreviousReport 404 invalid report id", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/1232112/previous", strings.NewReader(""))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		rec := httptest.NewRecorder()

		c := e.NewContext(req, rec)
		c.SetParamNames("rid")
		c.SetParamValues("1232112")

		err := api.GetPreviousReport(c)
		if assert.Error(t, err) {
			httpError, ok := err.(*echo.HTTPError)
			assert.True(t, ok)
			assert.Equal(t, http.StatusNotFound, httpError.Code)
		}
	})

	t.Run("GetPreviousReport 404 no previous", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/"+rid+"/previous", strings.NewReader(""))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		rec := httptest.NewRecorder()

		c := e.NewContext(req, rec)
		c.SetParamNames("rid")
		c.SetParamValues(rid)

		err := api.GetPreviousReport(c)
		if assert.Error(t, err) {
			httpError, ok := err.(*echo.HTTPError)
			assert.True(t, ok)
			assert.Equal(t, http.StatusNotFound, httpError.Code)
		}
	})

	t.Run("ListReportsAll", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/", strings.NewReader(`{}`))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		rec := httptest.NewRecorder()

		c := e.NewContext(req, rec)

		if assert.NoError(t, api.ListReportsAll(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)

			list := new(ReportList)
			err = json.NewDecoder(rec.Body).Decode(list)

			assert.NoError(t, err)
			assert.Equal(t, uint(32), list.Total)
			assert.Len(t, list.Data, 20)
			assert.NotZero(t, list.Data[0].ID)
			assert.NotEmpty(t, list.Data[0].Name)

			// by default we sort by desc id
			assert.Equal(t, p2ridLatest, list.Data[0].ID)
		}
	})

	t.Run("ListReportsAll page 1", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/?page=1", strings.NewReader(`{}`))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		rec := httptest.NewRecorder()

		c := e.NewContext(req, rec)

		if assert.NoError(t, api.ListReportsAll(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)

			list := new(ReportList)
			err = json.NewDecoder(rec.Body).Decode(list)

			assert.NoError(t, err)
			assert.Equal(t, uint(32), list.Total)
			assert.Len(t, list.Data, 12)
			assert.NotZero(t, list.Data[11].ID)
			assert.NotEmpty(t, list.Data[11].Name)

			// by default we sort by desc id
			// so the last one should be the first one

			assert.Equal(t, reportID, list.Data[11].ID)
		}
	})

	t.Run("ListReportsAll page 2 empty", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/?page=2", strings.NewReader(`{}`))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		rec := httptest.NewRecorder()

		c := e.NewContext(req, rec)

		if assert.NoError(t, api.ListReportsAll(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)

			list := new(ReportList)
			err = json.NewDecoder(rec.Body).Decode(list)

			assert.NoError(t, err)
			assert.Equal(t, uint(32), list.Total)
			assert.Len(t, list.Data, 0)
			assert.Empty(t, list.Data)
		}
	})

	t.Run("ListReportsForProject p1", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/"+pid+"/reports", strings.NewReader(`{}`))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		rec := httptest.NewRecorder()

		c := e.NewContext(req, rec)
		c.SetParamNames("pid")
		c.SetParamValues(pid)

		if assert.NoError(t, api.ListReportsForProject(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)

			list := new(ReportList)
			err = json.NewDecoder(rec.Body).Decode(list)

			assert.NoError(t, err)
			assert.Equal(t, uint(21), list.Total)
			assert.Len(t, list.Data, 20)
			assert.NotZero(t, list.Data[0].ID)
			assert.NotEmpty(t, list.Data[0].Name)

			// by default we sort by desc id
			assert.Equal(t, reportID9, list.Data[10].ID)
			assert.Equal(t, reportID10, list.Data[9].ID)
		}
	})

	t.Run("ListReportsForProject p2", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/"+pid2+"/reports", strings.NewReader(`{}`))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		rec := httptest.NewRecorder()

		c := e.NewContext(req, rec)
		c.SetParamNames("pid")
		c.SetParamValues(pid2)

		if assert.NoError(t, api.ListReportsForProject(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)

			list := new(ReportList)
			err = json.NewDecoder(rec.Body).Decode(list)

			assert.NoError(t, err)
			assert.Equal(t, uint(11), list.Total)
			assert.Len(t, list.Data, 11)
			assert.NotZero(t, list.Data[0].ID)
			assert.NotEmpty(t, list.Data[0].Name)

			assert.Equal(t, p2ridLatest, list.Data[0].ID)
			assert.Equal(t, p2reportID5, list.Data[4].ID)
			assert.Equal(t, p2reportID6, list.Data[3].ID)
		}
	})

	t.Run("ListReportsForProject p1 page 1", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/"+pid+"/reports?page=1", strings.NewReader(`{}`))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		rec := httptest.NewRecorder()

		c := e.NewContext(req, rec)
		c.SetParamNames("pid")
		c.SetParamValues(pid)

		if assert.NoError(t, api.ListReportsForProject(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)

			list := new(ReportList)
			err = json.NewDecoder(rec.Body).Decode(list)

			assert.NoError(t, err)
			assert.Equal(t, uint(21), list.Total)
			assert.Len(t, list.Data, 1)
			assert.NotZero(t, list.Data[0].ID)
			assert.NotEmpty(t, list.Data[0].Name)

			assert.Equal(t, reportID, list.Data[0].ID)
		}
	})

	t.Run("ListReportsForProject p2 page 1 empty", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/"+pid2+"/reports?page=1", strings.NewReader(`{}`))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		rec := httptest.NewRecorder()

		c := e.NewContext(req, rec)
		c.SetParamNames("pid")
		c.SetParamValues(pid2)

		if assert.NoError(t, api.ListReportsForProject(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)

			list := new(ReportList)
			err = json.NewDecoder(rec.Body).Decode(list)

			assert.NoError(t, err)
			assert.Equal(t, uint(11), list.Total)
			assert.Len(t, list.Data, 0)
			assert.Empty(t, list.Data)
		}
	})

	t.Run("DeleteReportBulk p2", func(t *testing.T) {
		e := echo.New()
		idsStr := fmt.Sprintf(`[%+v, 123, %+v]`, p2reportID5, p2reportID6)
		reqJSON := `{"ids":` + idsStr + `}`
		req := httptest.NewRequest(http.MethodPost, "/bulk_delete", strings.NewReader(reqJSON))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		rec := httptest.NewRecorder()

		c := e.NewContext(req, rec)

		if assert.NoError(t, api.DeleteReportBulk(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)

			m := make(map[string]int)
			err = json.NewDecoder(rec.Body).Decode(&m)

			assert.NoError(t, err)
			assert.Equal(t, 2, m["deleted"])
		}
	})
}
