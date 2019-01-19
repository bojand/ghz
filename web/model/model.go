package model

import "time"

// Model base model definition. Copy of gorm.Model with custom tags
type Model struct {
	// The id
	ID uint `json:"id" gorm:"primary_key"`

	// The creation time
	CreatedAt time.Time `json:"createdAt"`

	// The updated time
	UpdatedAt time.Time `json:"updatedAt"`
}
