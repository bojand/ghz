package database

import (
	"os"
	"testing"
	"time"

	"github.com/bojand/ghz/runner"
	"github.com/bojand/ghz/web/model"
	"github.com/stretchr/testify/assert"
)

func TestDatabase_Detail(t *testing.T) {
	os.Remove(dbName)

	defer os.Remove(dbName)

	db, err := New("sqlite3", dbName, false)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	defer db.Close()

	var rid, rid2 uint

	t.Run("new report", func(t *testing.T) {
		p := model.Project{
			Name:        "Test Proj 111 ",
			Description: "Test Description Asdf ",
		}

		r := model.Report{
			Project:   &p,
			Name:      "Test report",
			EndReason: "normal",
			Date:      time.Date(2018, 12, 1, 8, 0, 0, 0, time.UTC),
			Count:     200,
			Total:     time.Duration(2 * time.Second),
			Average:   time.Duration(10 * time.Millisecond),
			Fastest:   time.Duration(1 * time.Millisecond),
			Slowest:   time.Duration(100 * time.Millisecond),
			Rps:       2000,
		}

		err := db.CreateReport(&r)

		assert.NoError(t, err)
		assert.NotZero(t, p.ID)
		assert.NotZero(t, r.ID)

		rid = r.ID
	})

	t.Run("new report 2", func(t *testing.T) {
		p := model.Project{
			Name:        "Test Proj 222 ",
			Description: "Test Description 222 ",
		}

		r := model.Report{
			Project:   &p,
			Name:      "Test report 2",
			EndReason: "normal",
			Date:      time.Date(2018, 12, 1, 8, 0, 0, 0, time.UTC),
			Count:     300,
			Total:     time.Duration(2 * time.Second),
			Average:   time.Duration(10 * time.Millisecond),
			Fastest:   time.Duration(1 * time.Millisecond),
			Slowest:   time.Duration(100 * time.Millisecond),
			Rps:       3000,
		}

		err := db.CreateReport(&r)

		assert.NoError(t, err)
		assert.NotZero(t, p.ID)
		assert.NotZero(t, r.ID)

		rid2 = r.ID
	})

	t.Run("CreateDetailsBatch()", func(t *testing.T) {
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

	t.Run("CreateDetailsBatch() 2", func(t *testing.T) {
		M := 300
		s := make([]*model.Detail, M)

		for n := 0; n < M; n++ {
			nd := &model.Detail{
				ReportID: rid2,
				ResultDetail: runner.ResultDetail{
					Timestamp: time.Now(),
					Latency:   time.Duration(1 * time.Millisecond),
					Status:    "OK",
				},
			}

			s[n] = nd
		}

		created, errored := db.CreateDetailsBatch(rid2, s)

		assert.Equal(t, M, int(created))
		assert.Equal(t, 0, int(errored))
	})

	t.Run("CreateDetailsBatch() for unknown", func(t *testing.T) {
		M := 100
		s := make([]*model.Detail, M)

		for n := 0; n < M; n++ {
			nd := &model.Detail{
				ReportID: 43232,
				ResultDetail: runner.ResultDetail{
					Timestamp: time.Now(),
					Latency:   time.Duration(1 * time.Millisecond),
					Status:    "OK",
				},
			}

			s[n] = nd
		}

		created, errored := db.CreateDetailsBatch(43232, s)

		assert.Equal(t, 0, int(created))
		assert.Equal(t, M, int(errored))
	})

	t.Run("ListAllDetailsForReport", func(t *testing.T) {
		details, err := db.ListAllDetailsForReport(rid)

		assert.NoError(t, err)
		assert.Len(t, details, 200)
	})

	t.Run("ListAllDetailsForReport 2", func(t *testing.T) {
		details, err := db.ListAllDetailsForReport(rid2)

		assert.NoError(t, err)
		assert.Len(t, details, 300)
	})

	t.Run("ListAllDetailsForReport unknown", func(t *testing.T) {
		details, err := db.ListAllDetailsForReport(43332)

		assert.NoError(t, err)
		assert.Len(t, details, 0)
	})
}
