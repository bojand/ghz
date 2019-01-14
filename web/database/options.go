package database

import (
	"github.com/bojand/ghz/web/model"
)

// CreateOptions creates a new report
func (d *Database) CreateOptions(o *model.Options) error {
	return d.DB.Create(o).Error
}

// GetOptionsForReport creates a new report
func (d *Database) GetOptionsForReport(rid uint) (*model.Options, error) {
	r := &model.Report{}
	r.ID = rid
	o := new(model.Options)
	err := d.DB.Model(r).Related(&o).Error

	if err != nil {
		return nil, err
	}

	return o, err
}
