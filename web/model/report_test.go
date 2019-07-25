package model

import (
	"os"
	"testing"
	"time"

	"github.com/bojand/ghz/runner"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
)

func TestReport_BeforeSave(t *testing.T) {
	var reports = []struct {
		name        string
		in          *Report
		expected    *Report
		expectError bool
	}{
		{"no project id", &Report{}, &Report{}, true},
		{"with project id", &Report{ProjectID: 123}, &Report{ProjectID: 123, Status: "ok"}, false},
	}

	for _, tt := range reports {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.in.BeforeSave()
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expected, tt.in)
		})
	}
}

func TestReport(t *testing.T) {
	os.Remove(dbName)

	defer os.Remove(dbName)

	db, err := gorm.Open("sqlite3", dbName)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	defer db.Close()

	db.LogMode(true)

	db.Exec("PRAGMA foreign_keys = ON;")
	db.AutoMigrate(&Project{}, &Report{}, &Detail{})

	var rid, pid uint

	t.Run("create", func(t *testing.T) {
		p := Project{
			Name:        "Test Project 111 ",
			Description: "Test Description Asdf ",
		}

		r := Report{
			Project:   &p,
			Name:      "Test report",
			EndReason: "normal",
			Date:      time.Now(),
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

		r.Tags = map[string]string{
			"env":        "staging",
			"created by": "Joe Developer",
		}

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

		err := db.Create(&r).Error

		assert.NoError(t, err)
		assert.NotZero(t, p.ID)
		assert.NotZero(t, r.ID)

		pid = p.ID
		rid = r.ID

		p2 := new(Project)
		err = db.First(p2, p.ID).Error

		assert.NoError(t, err)
		assert.Equal(t, p.Name, p2.Name)
		assert.Equal(t, "Test Description Asdf", p2.Description)
		assert.Equal(t, StatusOK, p2.Status)
		assert.NotZero(t, p2.CreatedAt)
		assert.NotZero(t, p2.UpdatedAt)
	})

	t.Run("read", func(t *testing.T) {
		r := new(Report)
		err = db.First(r, rid).Error

		assert.NoError(t, err)

		assert.Equal(t, pid, r.ProjectID)
		assert.NotZero(t, r.CreatedAt)
		assert.NotZero(t, r.UpdatedAt)
		assert.Equal(t, StatusOK, r.Status)

		assert.Equal(t, "Test report", r.Name)
		assert.Equal(t, "normal", r.EndReason)
		assert.NotZero(t, r.Date)

		assert.Equal(t, 3, r.ErrorDist["rpc error: code = Internal desc = Internal error."])
		assert.Equal(t, 2, r.ErrorDist["rpc error: code = DeadlineExceeded desc = Deadline exceeded."])

		assert.Equal(t, 195, r.StatusCodeDist["OK"])
		assert.Equal(t, 3, r.StatusCodeDist["Internal"])
		assert.Equal(t, 2, r.StatusCodeDist["DeadlineExceeded"])

		assert.NotNil(t, r.LatencyDistribution)
		assert.Len(t, r.LatencyDistribution, 6)
		assert.Equal(t, &runner.LatencyDistribution{
			Percentage: 25,
			Latency:    time.Duration(1 * time.Millisecond),
		}, r.LatencyDistribution[0])
		assert.Equal(t, &runner.LatencyDistribution{
			Percentage: 99,
			Latency:    time.Duration(25 * time.Millisecond),
		}, r.LatencyDistribution[5])

		assert.Equal(t, "staging", r.Tags["env"])
		assert.Equal(t, "Joe Developer", r.Tags["created by"])
	})

	t.Run("create with project id", func(t *testing.T) {
		r := Report{
			ProjectID: pid,
			Name:      "Test report 2",
			EndReason: "cancelled",
			Date:      time.Now(),
			Count:     300,
			Total:     time.Duration(3 * time.Second),
			Average:   time.Duration(11 * time.Millisecond),
			Fastest:   time.Duration(2 * time.Millisecond),
			Slowest:   time.Duration(120 * time.Millisecond),
			Rps:       2100,
		}

		err := db.Create(&r).Error

		assert.NoError(t, err)
		assert.NotZero(t, r.ID)

		r2 := new(Report)
		err = db.First(r2, r.ID).Error

		assert.NoError(t, err)
		assert.Equal(t, r.Name, r2.Name)
		assert.Equal(t, r.EndReason, r2.EndReason)
		assert.Equal(t, StatusOK, r2.Status)
		assert.NotZero(t, r2.CreatedAt)
		assert.NotZero(t, r2.UpdatedAt)
		assert.Equal(t, uint64(300), r2.Count)
		assert.Equal(t, float64(2100), r2.Rps)
	})

	t.Run("fail with invalid project id", func(t *testing.T) {
		r := Report{
			ProjectID: 123432,
			Name:      "Test report 2",
			EndReason: "cancelled",
			Date:      time.Now(),
			Count:     300,
			Total:     time.Duration(3 * time.Second),
			Average:   time.Duration(11 * time.Millisecond),
			Fastest:   time.Duration(2 * time.Millisecond),
			Slowest:   time.Duration(120 * time.Millisecond),
			Rps:       2100,
		}

		err := db.Create(&r).Error

		assert.Error(t, err)
	})
}
