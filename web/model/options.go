package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"

	"github.com/bojand/ghz/runner"
	"github.com/jinzhu/gorm"
)

// Options represents a report detail
type Options struct {
	Model

	Report *Report `json:"-"`

	// Run id
	ReportID uint `json:"reportID" gorm:"type:integer REFERENCES reports(id) ON DELETE CASCADE;not null"`

	Info *OptionsInfo `json:"info,omitempty" gorm:"type:TEXT"`
}

// BeforeSave is called by GORM before save
func (o *Options) BeforeSave(scope *gorm.Scope) error {
	if o.ReportID == 0 && o.Report == nil {
		return errors.New("Options must belong to a report")
	}

	return nil
}

// OptionsInfo represents the report options
type OptionsInfo runner.Options

// Value converts options struct to a database value
func (o OptionsInfo) Value() (driver.Value, error) {
	v, err := json.Marshal(o)
	if err != nil {
		return nil, err
	}
	return string(v), nil
}

// Scan converts database value to an Options struct
func (o *OptionsInfo) Scan(src interface{}) error {
	var sourceStr string
	sourceByte, ok := src.([]byte)
	if !ok {
		sourceStr, ok = src.(string)
		if !ok {
			return errors.New("type assertion from string / byte")
		}
		sourceByte = []byte(sourceStr)
	}

	if err := json.Unmarshal(sourceByte, o); err != nil {
		return err
	}

	return nil
}
