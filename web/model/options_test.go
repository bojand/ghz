package model

import (
	"os"
	"testing"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
)

func TestOptions_BeforeSave(t *testing.T) {
	var hs = []struct {
		name        string
		in          *Options
		expected    *Options
		expectError bool
	}{
		{"no report id", &Options{}, &Options{}, true},
		{"with report id", &Options{ReportID: 123}, &Options{ReportID: 123}, false},
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

func TestOptions(t *testing.T) {
	os.Remove(dbName)

	defer os.Remove(dbName)

	db, err := gorm.Open("sqlite3", dbName)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	defer db.Close()

	db.LogMode(true)

	db.Exec("PRAGMA foreign_keys = ON;")
	db.AutoMigrate(&Project{}, &Report{}, &Options{})

	var rid, oid uint

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

		o := Options{
			Report: &r,
			Info: &OptionsInfo{
				Call:  "helloworld.Greeter.SayHi",
				Proto: "greeter.proto",
			},
		}

		err := db.Create(&o).Error

		assert.NoError(t, err)
		assert.NotZero(t, p.ID)
		assert.NotZero(t, r.ID)
		assert.NotZero(t, o.ID)

		oid = o.ID
		rid = r.ID
	})

	t.Run("read", func(t *testing.T) {
		o := new(Options)
		err = db.First(o, oid).Error

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

	t.Run("fail create with unknown report id", func(t *testing.T) {
		o := Options{
			ReportID: 123213,
			Info: &OptionsInfo{
				Call:  "helloworld.Greeter.SayHi",
				Proto: "greeter.proto",
			},
		}

		err := db.Create(&o).Error

		assert.Error(t, err)
	})
}
