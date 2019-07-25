package model

import (
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/bojand/ghz/runner"
)

// Detail represents a report detail
type Detail struct {
	Model

	Report *Report `json:"-"`

	// Run id
	ReportID uint `json:"reportID" gorm:"type:integer REFERENCES reports(id) ON DELETE CASCADE;not null"`

	runner.ResultDetail
}

const layoutISO string = "2006-01-02T15:04:05.000Z"
const layoutISO2 string = "2006-01-02T15:04:05-0700"

// UnmarshalJSON for Detail
func (d *Detail) UnmarshalJSON(data []byte) error {
	type Alias Detail
	aux := &struct {
		Timestamp string `json:"timestamp"`
		*Alias
	}{
		Alias: (*Alias)(d),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	err := json.Unmarshal([]byte(aux.Timestamp), &d.Timestamp)
	if err != nil {
		d.Timestamp, err = time.Parse(time.RFC3339Nano, aux.Timestamp)
	}
	if err != nil {
		d.Timestamp, err = time.Parse(time.RFC3339, aux.Timestamp)
	}
	if err != nil {
		d.Timestamp, err = time.Parse(layoutISO, aux.Timestamp)
	}
	if err != nil {
		d.Timestamp, err = time.Parse(layoutISO2, aux.Timestamp)
	}

	return err
}

// BeforeSave is called by GORM before save
func (d *Detail) BeforeSave() error {
	if d.ReportID == 0 && d.Report == nil {
		return errors.New("Detail must belong to a report")
	}

	d.Error = strings.TrimSpace(d.Error)

	status := strings.TrimSpace(d.Status)
	if status == "" {
		status = "OK"
	}
	d.Status = status

	return nil
}
