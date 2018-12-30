package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"

	"github.com/bojand/ghz/runner"
	"github.com/jinzhu/gorm"
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

// LatencyDistribution represents a histogram
type LatencyDistribution struct {
	Model

	ReportID uint    `json:"reportID" gorm:"type:integer REFERENCES reports(id);not null"`
	Report   *Report `json:"-"`

	List LatencyDistributionList `json:"list" gorm:"type:TEXT"`
}

// BeforeSave is called by GORM before save
func (ld *LatencyDistribution) BeforeSave(scope *gorm.Scope) error {
	if ld.ReportID == 0 && ld.Report == nil {
		return errors.New("LatencyDistribution must belong to a report")
	}

	return nil
}
