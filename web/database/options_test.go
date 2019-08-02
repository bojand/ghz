package database

import (
	"os"
	"testing"
	"time"

	"github.com/bojand/ghz/runner"
	"github.com/bojand/ghz/web/model"
	"github.com/stretchr/testify/assert"
)

func TestDatabase_Options(t *testing.T) {
	os.Remove(dbName)

	defer os.Remove(dbName)

	db, err := New("sqlite3", dbName, false)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	defer db.Close()

	var rid, oid uint

	t.Run("new", func(t *testing.T) {
		t.Run("test new", func(t *testing.T) {
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

			o := model.Options{
				Report: &r,
				Info: &model.OptionsInfo{
					Call:  "helloworld.Greeter.SayHi",
					Proto: "greeter.proto",
				},
			}

			err := db.CreateOptions(&o)

			assert.NoError(t, err)
			assert.NotZero(t, p.ID)
			assert.NotZero(t, r.ID)
			assert.NotZero(t, o.ID)

			rid = r.ID
			oid = o.ID
		})
	})

	t.Run("GetOptionsForReport", func(t *testing.T) {
		o, err := db.GetOptionsForReport(rid)

		assert.NoError(t, err)

		assert.Equal(t, rid, o.ReportID)
		assert.Equal(t, oid, o.ID)
		assert.Nil(t, o.Report)
		assert.NotZero(t, o.CreatedAt)
		assert.NotZero(t, o.UpdatedAt)

		assert.NotNil(t, o.Info)
		assert.Equal(t, "helloworld.Greeter.SayHi", o.Info.Call)
		assert.Equal(t, "greeter.proto", o.Info.Proto)
	})

	t.Run("GetOptionsForReport invalid report id", func(t *testing.T) {
		o, err := db.GetOptionsForReport(12321)

		assert.Error(t, err)
		assert.Nil(t, o)
	})
}
