package database

import (
	"github.com/bojand/ghz/web/model"
)

// FindReportByID gets the report by id
func (d *Database) FindReportByID(id uint) (*model.Report, error) {
	p := new(model.Report)
	err := d.DB.First(p, id).Error
	if err != nil {
		p = nil
	}
	return p, err
}

// CountReports returns the number of reports
func (d *Database) CountReports() (uint, error) {
	p := new(model.Report)
	count := uint(0)
	err := d.DB.Model(p).Count(&count).Error
	return count, err
}

// CountReportsForProject returns the number of reports
func (d *Database) CountReportsForProject(pid uint) (uint, error) {
	count := uint(0)
	err := d.DB.Model(&model.Report{}).Where("project_id = ?", pid).Count(&count).Error
	return count, err
}

// CreateReport creates a new report
func (d *Database) CreateReport(r *model.Report) error {
	return d.DB.Create(r).Error
}

// FindPreviousReport find previous report for the report id
func (d *Database) FindPreviousReport(rid uint) (*model.Report, error) {
	report, err := d.FindReportByID(rid)
	if err != nil {
		return nil, err
	}

	previous := new(model.Report)

	orderSQL := "date desc"
	whereSQL := "project_id = ? AND date < ?"

	err = d.DB.Debug().Where(whereSQL, report.ProjectID, report.Date).Order(orderSQL).Limit(1).Find(&previous).Error

	return previous, err
}

// ListReports lists reports using sorting
func (d *Database) ListReports(limit, page uint, sortField, order string) ([]*model.Report, error) {
	return d.listReports(false, 0, limit, page, sortField, order)
}

// ListReportsForProject lists reports using sorting
func (d *Database) ListReportsForProject(pid, limit, page uint, sortField, order string) ([]*model.Report, error) {
	return d.listReports(true, pid, limit, page, sortField, order)
}

func (d *Database) listReports(byProject bool, pid, limit, page uint, sortField, order string) ([]*model.Report, error) {
	if sortField != "id" && sortField != "date" {
		sortField = "id"
	}

	if order != "asc" && order != "desc" {
		order = "desc"
	}

	offset := uint(0)
	if page >= 0 && limit >= 0 {
		offset = page * limit
	}

	orderSQL := sortField + " " + string(order)

	if byProject {
		p := &model.Project{}
		p.ID = pid

		s := make([]*model.Report, limit)

		err := d.DB.Order(orderSQL).Offset(offset).Limit(limit).Model(p).Related(&s).Error

		return s, err
	}

	s := make([]*model.Report, limit)

	err := d.DB.Order(orderSQL).Offset(offset).Limit(limit).Find(&s).Error

	return s, err
}
