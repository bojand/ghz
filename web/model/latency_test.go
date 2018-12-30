package model

import (
	"os"
	"testing"
	"time"

	"github.com/bojand/ghz/runner"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
)

func TestLatency_BeforeSave(t *testing.T) {
	var hs = []struct {
		name        string
		in          *LatencyDistribution
		expected    *LatencyDistribution
		expectError bool
	}{
		{"no report id", &LatencyDistribution{}, &LatencyDistribution{}, true},
		{"with report id", &LatencyDistribution{ReportID: 123}, &LatencyDistribution{ReportID: 123}, false},
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

func TestLatencyDistribution(t *testing.T) {
	defer os.Remove(dbName)

	db, err := gorm.Open("sqlite3", dbName)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	defer db.Close()

	db.LogMode(true)

	db.Exec("PRAGMA foreign_keys = ON;")
	db.AutoMigrate(&Project{}, &Report{}, &LatencyDistribution{})

	var rid, lid uint

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

		ld := LatencyDistribution{
			Report: &r,
			List: []*runner.LatencyDistribution{
				&runner.LatencyDistribution{
					Percentage: 25,
					Latency:    time.Duration(1 * time.Millisecond),
				},
				&runner.LatencyDistribution{
					Percentage: 50,
					Latency:    time.Duration(5 * time.Millisecond),
				},
				&runner.LatencyDistribution{
					Percentage: 75,
					Latency:    time.Duration(10 * time.Millisecond),
				},
				&runner.LatencyDistribution{
					Percentage: 90,
					Latency:    time.Duration(15 * time.Millisecond),
				},
				&runner.LatencyDistribution{
					Percentage: 95,
					Latency:    time.Duration(20 * time.Millisecond),
				},
				&runner.LatencyDistribution{
					Percentage: 99,
					Latency:    time.Duration(25 * time.Millisecond),
				},
			},
		}

		err := db.Create(&ld).Error

		assert.NoError(t, err)
		assert.NotZero(t, p.ID)
		assert.NotZero(t, r.ID)
		assert.NotZero(t, ld.ID)

		lid = ld.ID
		rid = r.ID
	})

	t.Run("read", func(t *testing.T) {
		ld := new(LatencyDistribution)
		err = db.First(ld, lid).Error

		assert.NoError(t, err)

		assert.Equal(t, rid, ld.ReportID)
		assert.Equal(t, lid, ld.ID)
		assert.Nil(t, ld.Report)
		assert.NotNil(t, ld.CreatedAt)
		assert.NotNil(t, ld.UpdatedAt)
		assert.Nil(t, ld.DeletedAt)

		assert.NotNil(t, ld.List)
		assert.Len(t, ld.List, 6)
		assert.Equal(t, &runner.LatencyDistribution{
			Percentage: 25,
			Latency:    time.Duration(1 * time.Millisecond),
		}, ld.List[0])
		assert.Equal(t, &runner.LatencyDistribution{
			Percentage: 99,
			Latency:    time.Duration(25 * time.Millisecond),
		}, ld.List[5])
	})

	t.Run("test create with report id", func(t *testing.T) {
		ld := LatencyDistribution{
			ReportID: rid,
			List: []*runner.LatencyDistribution{
				&runner.LatencyDistribution{
					Percentage: 25,
					Latency:    time.Duration(2 * time.Millisecond),
				},
				&runner.LatencyDistribution{
					Percentage: 50,
					Latency:    time.Duration(6 * time.Millisecond),
				},
				&runner.LatencyDistribution{
					Percentage: 75,
					Latency:    time.Duration(11 * time.Millisecond),
				},
				&runner.LatencyDistribution{
					Percentage: 90,
					Latency:    time.Duration(16 * time.Millisecond),
				},
				&runner.LatencyDistribution{
					Percentage: 95,
					Latency:    time.Duration(21 * time.Millisecond),
				},
				&runner.LatencyDistribution{
					Percentage: 99,
					Latency:    time.Duration(26 * time.Millisecond),
				},
			},
		}

		err := db.Create(&ld).Error

		assert.NoError(t, err)
		assert.NotZero(t, ld.ID)

		l2 := new(LatencyDistribution)
		err = db.First(l2, ld.ID).Error

		assert.NoError(t, err)

		assert.Equal(t, rid, l2.ReportID)
		assert.Equal(t, ld.ID, l2.ID)
		assert.Nil(t, l2.Report)
		assert.NotNil(t, l2.CreatedAt)
		assert.NotNil(t, l2.UpdatedAt)
		assert.Nil(t, l2.DeletedAt)

		assert.NotNil(t, l2.List)
		assert.Len(t, l2.List, 6)
		assert.Equal(t, &runner.LatencyDistribution{
			Percentage: 25,
			Latency:    time.Duration(2 * time.Millisecond),
		}, l2.List[0])
		assert.Equal(t, &runner.LatencyDistribution{
			Percentage: 75,
			Latency:    time.Duration(11 * time.Millisecond),
		}, l2.List[2])
		assert.Equal(t, &runner.LatencyDistribution{
			Percentage: 99,
			Latency:    time.Duration(26 * time.Millisecond),
		}, l2.List[5])
	})

	t.Run("fail create with unknown report id", func(t *testing.T) {
		ld := LatencyDistribution{
			ReportID: 34242,
			List: []*runner.LatencyDistribution{
				&runner.LatencyDistribution{
					Percentage: 25,
					Latency:    time.Duration(2 * time.Millisecond),
				},
				&runner.LatencyDistribution{
					Percentage: 50,
					Latency:    time.Duration(6 * time.Millisecond),
				},
			},
		}

		err := db.Create(&ld).Error

		assert.Error(t, err)
	})
}
