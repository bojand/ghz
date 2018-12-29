package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/bojand/ghz/runner"
	"github.com/jinzhu/gorm"
)

// LatencyDistribution holds latency distribution data
type LatencyDistribution runner.LatencyDistribution

// LatencyDistributionList is a slice of LatencyDistribution pointers
type LatencyDistributionList []*LatencyDistribution

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

	var lds []LatencyDistribution
	if err := json.Unmarshal(sourceByte, &lds); err != nil {
		return err
	}

	for index := range lds {
		*ld = append(*ld, &lds[index])
	}

	return nil
}

// Bucket holds histogram data
type Bucket runner.Bucket

// BucketList is a slice of buckets
type BucketList []*Bucket

// Value converts struct to a database value
func (bl BucketList) Value() (driver.Value, error) {
	v, err := json.Marshal(bl)
	if err != nil {
		return nil, err
	}
	return string(v), nil
}

// Scan converts database value to a struct
func (bl *BucketList) Scan(src interface{}) error {
	var sourceStr string
	sourceByte, ok := src.([]byte)
	if !ok {
		sourceStr, ok = src.(string)
		if !ok {
			return errors.New("type assertion from string / byte")
		}
		sourceByte = []byte(sourceStr)
	}

	var buckets []Bucket
	if err := json.Unmarshal(sourceByte, &buckets); err != nil {
		return err
	}

	for index := range buckets {
		*bl = append(*bl, &buckets[index])
	}

	return nil
}

// Options represents the report options
type Options runner.Options

// Value converts options struct to a database value
func (o Options) Value() (driver.Value, error) {
	v, err := json.Marshal(o)
	if err != nil {
		return nil, err
	}
	return string(v), nil
}

// Scan converts database value to an Options struct
func (o *Options) Scan(src interface{}) error {
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

// Report represents a project
type Report struct {
	Model

	ProjectID uint     `json:"projectID" gorm:"type:integer REFERENCES projects(id)"`
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

	Options *Options `json:"options,omitempty" gorm:"type:varchar(512)"`

	ErrorDist      StringIntMap `json:"errorDistribution,omitempty" gorm:"type:varchar(512)"`
	StatusCodeDist StringIntMap `json:"statusCodeDistribution,omitempty" gorm:"type:varchar(512)"`

	LatencyDistribution LatencyDistributionList `json:"latencyDistribution" gorm:"type:varchar(512)"`
	Histogram           BucketList              `json:"histogram"  gorm:"type:varchar(512)"`
}

// BeforeSave is called by GORM before save
func (r *Report) BeforeSave(scope *gorm.Scope) error {
	if r.ProjectID == 0 && r.Project == nil {
		return errors.New("Report must belong to a project")
	}

	r.Status = StatusOK

	if scope != nil {
		scope.SetColumn("status", r.Status)
	}

	return nil
}
