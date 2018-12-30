package database

import (
	"github.com/bojand/ghz/web/model"
)

// CreateHistogram creates a new report
func (d *Database) CreateHistogram(h *model.Histogram) error {
	return d.DB.Create(h).Error
}

// GetHistogramForReport creates a new report
func (d *Database) GetHistogramForReport(rid uint) (*model.Histogram, error) {
	r := &model.Report{}
	r.ID = rid
	h := new(model.Histogram)
	err := d.DB.Model(r).Related(&h).Error
	return h, err
}
