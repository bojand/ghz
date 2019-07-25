package model

import (
	"os"
	"testing"
	"time"

	"github.com/bojand/ghz/runner"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
)

func TestHistogram_BeforeSave(t *testing.T) {
	var hs = []struct {
		name        string
		in          *Histogram
		expected    *Histogram
		expectError bool
	}{
		{"no report id", &Histogram{}, &Histogram{}, true},
		{"with report id", &Histogram{ReportID: 123}, &Histogram{ReportID: 123}, false},
	}

	for _, tt := range hs {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.in.BeforeSave(nil)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expected, tt.in)
		})
	}
}

func TestHistogram(t *testing.T) {
	os.Remove(dbName)

	defer os.Remove(dbName)

	db, err := gorm.Open("sqlite3", dbName)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	defer db.Close()

	db.LogMode(true)

	db.Exec("PRAGMA foreign_keys = ON;")
	db.AutoMigrate(&Project{}, &Report{}, &Histogram{})

	var rid, hid uint

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

		h := Histogram{
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

		err := db.Create(&h).Error

		assert.NoError(t, err)
		assert.NotZero(t, p.ID)
		assert.NotZero(t, r.ID)
		assert.NotZero(t, h.ID)

		hid = h.ID
		rid = r.ID
	})

	t.Run("read", func(t *testing.T) {
		h := new(Histogram)
		err = db.First(h, hid).Error

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
	})

	t.Run("test create with report id", func(t *testing.T) {
		ch := Histogram{
			ReportID: rid,
			Buckets: []*runner.Bucket{
				{
					Mark:      0.02,
					Count:     2,
					Frequency: 0.006,
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
					Mark:      0.2,
					Count:     16,
					Frequency: 0.08,
				},
			},
		}

		err := db.Create(&ch).Error

		assert.NoError(t, err)
		assert.NotZero(t, ch.ID)

		h := new(Histogram)
		err = db.First(h, ch.ID).Error

		assert.NoError(t, err)

		assert.Equal(t, rid, h.ReportID)
		assert.Equal(t, ch.ID, h.ID)
		assert.Nil(t, h.Report)
		assert.NotZero(t, h.CreatedAt)
		assert.NotZero(t, h.UpdatedAt)

		assert.NotNil(t, h.Buckets)
		assert.Len(t, h.Buckets, 5)
		assert.Equal(t, &runner.Bucket{
			Mark:      0.02,
			Count:     2,
			Frequency: 0.006,
		}, h.Buckets[0])
		assert.Equal(t, &runner.Bucket{
			Mark:      0.2,
			Count:     16,
			Frequency: 0.08,
		}, h.Buckets[4])
	})

	t.Run("fail create with unknown report id", func(t *testing.T) {
		ch := Histogram{
			ReportID: 123213,
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
			},
		}

		err := db.Create(&ch).Error

		assert.Error(t, err)
	})
}
