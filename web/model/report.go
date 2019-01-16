package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/bojand/ghz/runner"
)

// LatencyDistributionList is a slice of LatencyDistribution pointers
type LatencyDistributionList []*runner.LatencyDistribution

// Value converts struct to a database value
func (ld LatencyDistributionList) Value() (driver.Value, error) {
	v, err := json.Marshal(ld)
	if err != nil {
		return nil, err
	}
	return string(v), nil
}

// Scan converts database value to a struct
func (ld *LatencyDistributionList) Scan(src interface{}) error {
	var sourceStr string
	sourceByte, ok := src.([]byte)
	if !ok {
		sourceStr, ok = src.(string)
		if !ok {
			return errors.New("type assertion from string / byte")
		}
		sourceByte = []byte(sourceStr)
	}

	var lds []runner.LatencyDistribution
	if err := json.Unmarshal(sourceByte, &lds); err != nil {
		return err
	}

	for index := range lds {
		*ld = append(*ld, &lds[index])
	}

	return nil
}

// StringIntMap is a map of string keys to int values
type StringIntMap map[string]int

// Value converts map to database value
func (m StringIntMap) Value() (driver.Value, error) {
	v, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return string(v), nil
}

// Scan converts database value to a map
func (m *StringIntMap) Scan(src interface{}) error {
	var sourceStr string
	sourceByte, ok := src.([]byte)
	if !ok {
		sourceStr, ok = src.(string)
		if !ok {
			return errors.New("type assertion from string / byte")
		}
		sourceByte = []byte(sourceStr)
	}

	if err := json.Unmarshal(sourceByte, m); err != nil {
		return err
	}

	return nil
}

// StringStringMap is a map of string keys to int values
type StringStringMap map[string]string

// Value converts map to database value
func (m StringStringMap) Value() (driver.Value, error) {
	v, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return string(v), nil
}

// Scan converts database value to a map
func (m *StringStringMap) Scan(src interface{}) error {
	var sourceStr string
	sourceByte, ok := src.([]byte)
	if !ok {
		sourceStr, ok = src.(string)
		if !ok {
			return errors.New("type assertion from string / byte")
		}
		sourceByte = []byte(sourceStr)
	}

	if err := json.Unmarshal(sourceByte, m); err != nil {
		return err
	}

	return nil
}

// Report represents a project
type Report struct {
	Model

	ProjectID uint     `json:"projectID" gorm:"type:integer REFERENCES projects(id) ON DELETE CASCADE;not null"`
	Project   *Project `json:"-"`

	Name      string    `json:"name,omitempty"`
	EndReason string    `json:"endReason,omitempty"`
	Date      time.Time `json:"date"`

	Count   uint64        `json:"count"`
	Total   time.Duration `json:"total"`
	Average time.Duration `json:"average"`
	Fastest time.Duration `json:"fastest"`
	Slowest time.Duration `json:"slowest"`
	Rps     float64       `json:"rps"`

	Status Status `json:"status" gorm:"not null"`

	ErrorDist      StringIntMap `json:"errorDistribution,omitempty" gorm:"type:TEXT"`
	StatusCodeDist StringIntMap `json:"statusCodeDistribution,omitempty" gorm:"type:TEXT"`

	LatencyDistribution LatencyDistributionList `json:"latencyDistribution" gorm:"type:TEXT"`

	Tags StringStringMap `json:"tags,omitempty" gorm:"type:TEXT"`
}

// BeforeSave is called by GORM before save
func (r *Report) BeforeSave() error {
	if r.ProjectID == 0 && r.Project == nil {
		return errors.New("Report must belong to a project")
	}

	if string(r.Status) == "" {
		r.Status = StatusOK
	}

	return nil
}
