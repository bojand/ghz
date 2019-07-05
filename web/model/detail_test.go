package model

import (
	"os"
	"testing"
	"time"

	"github.com/bojand/ghz/runner"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
)

func TestDetail_BeforeSave(t *testing.T) {
	var details = []struct {
		name        string
		in          *Detail
		expected    *Detail
		expectError bool
	}{
		{"no run", &Detail{ResultDetail: runner.ResultDetail{Latency: 12345}}, &Detail{ResultDetail: runner.ResultDetail{Latency: 12345}}, true},
		{"trim error", &Detail{ReportID: 1, ResultDetail: runner.ResultDetail{Latency: 12345, Error: " network error "}}, &Detail{ReportID: 1, ResultDetail: runner.ResultDetail{Latency: 12345, Error: "network error", Status: "OK"}}, false},
		{"trim status", &Detail{ReportID: 1, ResultDetail: runner.ResultDetail{Latency: 12345, Error: " network error ", Status: " OK "}}, &Detail{ReportID: 1, ResultDetail: runner.ResultDetail{Latency: 12345, Error: "network error", Status: "OK"}}, false},
	}

	for _, tt := range details {
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

func TestDetail_UnmarshalJSON(t *testing.T) {
	expectedTime, err := time.Parse("2006-01-02T15:04:05-0700", "2018-08-08T13:00:00-0300")
	assert.NoError(t, err)

	var details = []struct {
		name        string
		in          string
		expected    *Detail
		expectError bool
	}{
		{"RFC3339",
			`{"timestamp":"2018-08-08T13:00:00.000000000-03:00","latency":123,"error":"","status":"OK"}`,
			&Detail{ResultDetail: runner.ResultDetail{Timestamp: expectedTime, Latency: 123, Error: "", Status: "OK"}},
			false},
		{"layoutISO2",
			`{"timestamp":"2018-08-08T13:00:00-0300","latency":123,"error":"","status":"OK"}`,
			&Detail{ResultDetail: runner.ResultDetail{Timestamp: expectedTime, Latency: 123, Error: "", Status: "OK"}},
			false},
	}

	for _, tt := range details {
		t.Run(tt.name, func(t *testing.T) {
			var d Detail
			err := d.UnmarshalJSON([]byte(tt.in))
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expected, &d)
		})
	}
}

func TestDetail(t *testing.T) {
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

	var rid, did uint

	t.Run("test create", func(t *testing.T) {
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

		d := Detail{
			Report: &r,
			ResultDetail: runner.ResultDetail{
				Timestamp: time.Now(),
				Latency:   time.Duration(1 * time.Millisecond),
				Status:    "OK",
			},
		}

		err := db.Create(&d).Error

		assert.NoError(t, err)
		assert.NotZero(t, p.ID)
		assert.NotZero(t, r.ID)
		assert.NotZero(t, d.ID)

		did = d.ID
		rid = r.ID
	})

	t.Run("read", func(t *testing.T) {
		d := new(Detail)
		err = db.First(d, did).Error

		assert.NoError(t, err)

		assert.Equal(t, rid, d.ReportID)
		assert.Equal(t, did, d.ID)
		assert.Nil(t, d.Report)
		assert.NotZero(t, d.Timestamp)
		assert.Equal(t, time.Duration(1*time.Millisecond), d.Latency)
		assert.NotZero(t, d.CreatedAt)
		assert.NotZero(t, d.UpdatedAt)
		assert.Equal(t, "OK", d.Status)
	})

	t.Run("test create with report id", func(t *testing.T) {
		detail := Detail{
			ReportID: rid,
			ResultDetail: runner.ResultDetail{
				Timestamp: time.Now(),
				Latency:   time.Duration(2 * time.Millisecond),
				Status:    "CANCELED",
			},
		}

		err := db.Create(&detail).Error

		assert.NoError(t, err)
		assert.NotZero(t, detail.ID)

		d := new(Detail)
		err = db.First(d, detail.ID).Error

		assert.NoError(t, err)

		assert.Equal(t, rid, d.ReportID)
		assert.Equal(t, detail.ID, d.ID)
		assert.Nil(t, d.Report)
		assert.NotZero(t, d.Timestamp)
		assert.Equal(t, time.Duration(2*time.Millisecond), d.Latency)
		assert.NotZero(t, d.CreatedAt)
		assert.NotZero(t, d.UpdatedAt)
		assert.Equal(t, "CANCELED", d.Status)
	})

	t.Run("fail create with unknown report id", func(t *testing.T) {
		detail := Detail{
			ReportID: 1233211,
			ResultDetail: runner.ResultDetail{
				Timestamp: time.Now(),
				Latency:   time.Duration(2 * time.Millisecond),
				Status:    "CANCELED",
			},
		}

		err := db.Create(&detail).Error

		assert.Error(t, err)
	})
}
