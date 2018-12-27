package database

import (
	"errors"

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

// UpdateProject creates a new project
func (d *Database) UpdateProject(p *model.Project) error {
	return d.DB.Create(p).Error
}

// UpdateProjectStatus updates the project's status
func (d *Database) UpdateProjectStatus(pid uint, status model.Status) error {
	p := new(model.Project)
	p.ID = pid
	return d.DB.Model(p).Update("status", status).Error
}

// ListProjects lists projects
func (d *Database) ListProjects(limit, page uint) ([]*model.Project, error) {
	offset := uint(0)
	if page >= 0 && limit >= 0 {
		offset = page * limit
	}

	s := make([]*model.Project, limit)

	err := d.DB.Offset(offset).Limit(limit).Order("name desc").Find(&s).Error

	return s, err
}

// ListProjectsSorted lists projects using sorting
func (d *Database) ListProjectsSorted(limit, page uint, sortField string, order Order) ([]*model.Project, error) {
	if sortField != "name" && sortField != "id" {
		return nil, errors.New("Invalid sort parameters")
	}

	offset := uint(0)
	if page >= 0 && limit >= 0 {
		offset = page * limit
	}

	orderSQL := sortField + " " + string(order)

	s := make([]*model.Project, limit)

	err := d.DB.Order(orderSQL).Offset(offset).Limit(limit).Find(&s).Error

	return s, err
}
