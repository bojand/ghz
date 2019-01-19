package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"

	"github.com/bojand/ghz/runner"
	"github.com/jinzhu/gorm"
)

// BucketList is a slice of buckets
type BucketList []*runner.Bucket

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

	var buckets []runner.Bucket
	if err := json.Unmarshal(sourceByte, &buckets); err != nil {
		return err
	}

	for index := range buckets {
		*bl = append(*bl, &buckets[index])
	}

	return nil
}

// Histogram represents a histogram
type Histogram struct {
	Model

	ReportID uint    `json:"reportID" gorm:"type:integer REFERENCES reports(id) ON DELETE CASCADE;not null"`
	Report   *Report `json:"-"`

	Buckets BucketList `json:"buckets" gorm:"type:TEXT"`
}

// BeforeSave is called by GORM before save
func (h *Histogram) BeforeSave(scope *gorm.Scope) error {
	if h.ReportID == 0 && h.Report == nil {
		return errors.New("Histogram must belong to a report")
	}

	return nil
}
