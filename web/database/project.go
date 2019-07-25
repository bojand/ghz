package database

import (
	"github.com/bojand/ghz/web/model"
)

// FindProjectByID gets the project by id
func (d *Database) FindProjectByID(id uint) (*model.Project, error) {
	p := new(model.Project)
	err := d.DB.First(p, id).Error
	if err != nil {
		p = nil
	}
	return p, err
}

// CountProjects returns the number of projects
func (d *Database) CountProjects() (uint, error) {
	p := new(model.Project)
	count := uint(0)
	err := d.DB.Model(p).Count(&count).Error
	return count, err
}

// CreateProject creates a new project
func (d *Database) CreateProject(p *model.Project) error {
	return d.DB.Create(p).Error
}

// UpdateProject update a project
func (d *Database) UpdateProject(p *model.Project) error {
	return d.DB.Save(p).Error
}

// DeleteProject deletas an existing project
func (d *Database) DeleteProject(p *model.Project) error {
	return d.DB.Delete(p).Error
}

// UpdateProjectStatus updates the project's status
func (d *Database) UpdateProjectStatus(pid uint, status model.Status) error {
	p := new(model.Project)
	p.ID = pid

	// use UpdateColumn to circumvent update hooks and not modify updated at time
	return d.DB.Model(p).UpdateColumn("status", status).Error
}

// ListProjects lists projects using sorting
func (d *Database) ListProjects(limit, page uint, sortField, order string) ([]*model.Project, error) {
	if sortField != "name" && sortField != "id" {
		sortField = "id"
	}

	if order != "asc" && order != "desc" {
		order = "desc"
	}

	offset := uint(0)
	if page > 0 && limit > 0 {
		offset = page * limit
	}

	orderSQL := sortField + " " + string(order)

	s := make([]*model.Project, limit)

	err := d.DB.Order(orderSQL).Offset(offset).Limit(limit).Find(&s).Error

	return s, err
}
