package model

import (
	"errors"
	"strings"

	"github.com/bojand/hri"
	"github.com/jinzhu/gorm"
)

// Project represents a project
type Project struct {
	Model
	Name        string `json:"name" gorm:"not null"`
	Description string `json:"description"`
	Status      Status `json:"status" gorm:"not null"`
}

// BeforeCreate is a GORM hook called when a model is created
func (p *Project) BeforeCreate() error {
	if p.Name == "" {
		p.Name = hri.Random()
	}

	if string(p.Status) == "" {
		p.Status = StatusOK
	}

	return nil
}

// BeforeUpdate is a GORM hook called when a model is updated
func (p *Project) BeforeUpdate() error {
	if p.Name == "" {
		return errors.New("Project name cannot be empty")
	}

	return nil
}

// BeforeSave is a GORM hook called when a model is created or updated
func (p *Project) BeforeSave(scope *gorm.Scope) error {
	p.Name = strings.TrimSpace(p.Name)
	p.Description = strings.TrimSpace(p.Description)

	if scope != nil {
		scope.SetColumn("name", p.Name)
		scope.SetColumn("description", p.Description)
	}

	return nil
}
