package model

import (
	"os"
	"testing"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/stretchr/testify/assert"
)

const dbName = "../test/model_test.db"

func TestProject_Create(t *testing.T) {
	os.Remove(dbName)

	defer os.Remove(dbName)

	db, err := gorm.Open("sqlite3", dbName)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	defer db.Close()

	db.LogMode(true)

	// Migrate the schema
	db.AutoMigrate(&Project{})
	db.Exec("PRAGMA foreign_keys = ON;")

	t.Run("test new", func(t *testing.T) {
		p := Project{
			Name:        "Test Project 111 ",
			Description: "Test Description Asdf ",
		}

		err := db.Create(&p).Error

		assert.NoError(t, err)
		assert.NotZero(t, p.ID)
		assert.Equal(t, "Test Project 111", p.Name)
		assert.Equal(t, "Test Description Asdf", p.Description)
		assert.Equal(t, StatusOK, p.Status)
		assert.NotZero(t, p.CreatedAt)
		assert.NotZero(t, p.UpdatedAt)

		p2 := new(Project)
		err = db.First(p2, p.ID).Error

		assert.NoError(t, err)
		assert.Equal(t, "Test Project 111", p2.Name)
		assert.Equal(t, "Test Description Asdf", p2.Description)
		assert.Equal(t, StatusOK, p2.Status)
		assert.NotZero(t, p2.CreatedAt)
		assert.NotZero(t, p2.UpdatedAt)
	})

	t.Run("test new with empty name", func(t *testing.T) {
		p := Project{
			Description: "Test Description Asdf 2",
		}

		err := db.Create(&p).Error

		assert.NoError(t, err)
		assert.NotZero(t, p.ID)
		assert.NotEmpty(t, p.Name)
		assert.Equal(t, "Test Description Asdf 2", p.Description)
		assert.Equal(t, StatusOK, p.Status)
		assert.NotZero(t, p.CreatedAt)
		assert.NotZero(t, p.UpdatedAt)

		p2 := new(Project)
		err = db.First(p2, p.ID).Error

		assert.NoError(t, err)
		assert.Equal(t, p.Name, p2.Name)
		assert.Equal(t, "Test Description Asdf 2", p2.Description)
		assert.Equal(t, StatusOK, p2.Status)
		assert.NotZero(t, p2.CreatedAt)
		assert.NotZero(t, p2.UpdatedAt)
	})

	t.Run("test new with ID", func(t *testing.T) {
		p := Project{
			Name:        " FooProject ",
			Description: " Bar Desc ",
			Status:      StatusFail,
		}
		p.ID = 123

		err := db.Create(&p).Error

		assert.NoError(t, err)
		assert.Equal(t, uint(123), p.ID)
		assert.Equal(t, "FooProject", p.Name)
		assert.Equal(t, "Bar Desc", p.Description)
		assert.Equal(t, StatusFail, p.Status)
		assert.NotZero(t, p.CreatedAt)
		assert.NotZero(t, p.UpdatedAt)

		p2 := new(Project)
		err = db.First(p2, p.ID).Error

		assert.NoError(t, err)
		assert.Equal(t, uint(123), p2.ID)
		assert.Equal(t, "FooProject", p2.Name)
		assert.Equal(t, "Bar Desc", p2.Description)
		assert.Equal(t, StatusFail, p2.Status)
		assert.NotZero(t, p2.CreatedAt)
		assert.NotZero(t, p2.UpdatedAt)
	})

	t.Run("should fail with same ID", func(t *testing.T) {
		p := Project{
			Name:        "ACME",
			Description: "Lorem Ipsum",
		}
		p.ID = 123

		err := db.Create(&p).Error

		assert.Error(t, err)
	})
}

func TestProject_Save(t *testing.T) {
	os.Remove(dbName)

	defer os.Remove(dbName)

	db, err := gorm.Open("sqlite3", dbName)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	defer db.Close()

	db.LogMode(true)

	// Migrate the schema
	db.AutoMigrate(&Project{})

	var pid uint

	t.Run("test new", func(t *testing.T) {
		p := Project{
			Name:        "Test Project 111 ",
			Description: "Test Description Asdf ",
		}

		err := db.Create(&p).Error

		assert.NoError(t, err)
		assert.NotZero(t, p.ID)
		assert.Equal(t, "Test Project 111", p.Name)
		assert.Equal(t, "Test Description Asdf", p.Description)
		assert.Equal(t, StatusOK, p.Status)
		assert.NotZero(t, p.CreatedAt)
		assert.NotZero(t, p.UpdatedAt)

		p2 := new(Project)
		err = db.First(p2, p.ID).Error

		assert.NoError(t, err)
		assert.Equal(t, "Test Project 111", p2.Name)
		assert.Equal(t, "Test Description Asdf", p2.Description)
		assert.Equal(t, StatusOK, p2.Status)
		assert.NotZero(t, p2.CreatedAt)
		assert.NotZero(t, p2.UpdatedAt)

		pid = p2.ID
	})

	t.Run("test update existing", func(t *testing.T) {
		p := new(Project)
		err = db.First(p, pid).Error

		assert.NoError(t, err)

		p.Name = " New Name "
		p.Description = "Baz"

		err := db.Save(&p).Error

		assert.NoError(t, err)

		assert.NotZero(t, p.ID)
		assert.Equal(t, "New Name", p.Name)
		assert.Equal(t, "Baz", p.Description)
		assert.NotZero(t, p.CreatedAt)
		assert.NotZero(t, p.UpdatedAt)

		p2 := new(Project)
		err = db.First(p2, p.ID).Error

		assert.NoError(t, err)
		assert.Equal(t, uint(1), p2.ID)
		assert.Equal(t, "New Name", p2.Name)
		assert.Equal(t, "Baz", p2.Description)
		assert.NotZero(t, p2.CreatedAt)
		assert.NotZero(t, p2.UpdatedAt)
	})

	t.Run("test update existing no name", func(t *testing.T) {
		p := Project{
			Description: "Foo Test Bar",
		}
		p.ID = pid

		err := db.Save(&p).Error

		assert.Error(t, err)

		p2 := new(Project)
		err = db.First(p2, p.ID).Error

		assert.NoError(t, err)
		assert.Equal(t, uint(1), p2.ID)
		assert.Equal(t, "New Name", p2.Name)
		assert.Equal(t, "Baz", p2.Description)
		assert.NotZero(t, p2.CreatedAt)
		assert.NotZero(t, p2.UpdatedAt)
	})
}
