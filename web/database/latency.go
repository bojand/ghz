package database

import (
	"github.com/bojand/ghz/web/model"
)

// CreateLatencyDistribution creates a new report
func (d *Database) CreateLatencyDistribution(ld *model.LatencyDistribution) error {
	return d.DB.Create(ld).Error
}

// GetLatencyDistributionForReport creates a new report
func (d *Database) GetLatencyDistributionForReport(rid uint) (*model.LatencyDistribution, error) {
	r := &model.Report{}
	r.ID = rid
	ld := new(model.LatencyDistribution)
	err := d.DB.Model(r).Related(&ld).Error
	return ld, err
}
