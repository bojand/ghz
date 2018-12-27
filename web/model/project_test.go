package model

import (
	"os"
	"testing"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/stretchr/testify/assert"
)

const dbName = "../test/project_test.db"

func TestProject_Create(t *testing.T) {
	defer os.Remove(dbName)

	db, err := gorm.Open("sqlite3", dbName)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	defer db.Close()

	db.LogMode(true)

	// Migrate the schema
	db.AutoMigrate(&Project{})

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
		assert.NotNil(t, p.CreatedAt)
		assert.NotNil(t, p.UpdatedAt)
		assert.Nil(t, p.DeletedAt)

		p2 := new(Project)
		err = db.First(p2, p.ID).Error

		assert.NoError(t, err)
		assert.Equal(t, "Test Project 111", p2.Name)
		assert.Equal(t, "Test Description Asdf", p2.Description)
		assert.Equal(t, StatusOK, p2.Status)
		assert.NotNil(t, p2.CreatedAt)
		assert.NotNil(t, p2.UpdatedAt)
		assert.Nil(t, p2.DeletedAt)
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
		assert.NotNil(t, p.CreatedAt)
		assert.NotNil(t, p.UpdatedAt)
		assert.Nil(t, p.DeletedAt)

		p2 := new(Project)
		err = db.First(p2, p.ID).Error

		assert.NoError(t, err)
		assert.Equal(t, p.Name, p2.Name)
		assert.Equal(t, "Test Description Asdf 2", p2.Description)
		assert.Equal(t, StatusOK, p2.Status)
		assert.NotNil(t, p2.CreatedAt)
		assert.NotNil(t, p2.UpdatedAt)
		assert.Nil(t, p2.DeletedAt)
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
		assert.NotNil(t, p.CreatedAt)
		assert.NotNil(t, p.UpdatedAt)
		assert.Nil(t, p.DeletedAt)

		p2 := new(Project)
		err = db.First(p2, p.ID).Error

		assert.NoError(t, err)
		assert.Equal(t, uint(123), p2.ID)
		assert.Equal(t, "FooProject", p2.Name)
		assert.Equal(t, "Bar Desc", p2.Description)
		assert.Equal(t, StatusFail, p2.Status)
		assert.NotNil(t, p.CreatedAt)
		assert.NotNil(t, p.UpdatedAt)
		assert.Nil(t, p.DeletedAt)
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
	defer os.Remove(dbName)

	db, err := gorm.Open("sqlite3", dbName)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	defer db.Close()

	db.LogMode(true)

	// Migrate the schema
	db.AutoMigrate(&Project{})

	t.Run("test update existing", func(t *testing.T) {
		p := Project{
			Name:        " New Name ",
			Description: "Baz",
		}
		p.ID = uint(1)

		err := db.Save(&p).Error

		assert.NoError(t, err)

		assert.NotZero(t, p.ID)
		assert.Equal(t, "New Name", p.Name)
		assert.Equal(t, "Baz", p.Description)
		assert.NotNil(t, p.CreatedAt)
		assert.NotNil(t, p.UpdatedAt)
		assert.Nil(t, p.DeletedAt)
	})

	t.Run("test update existing just name", func(t *testing.T) {
		p := Project{
			Name: " New Name 2",
		}
		p.ID = uint(1)

		err := db.Save(&p).Error

		assert.NoError(t, err)

		assert.NotZero(t, p.ID)
		assert.Equal(t, "New Name 2", p.Name)
		assert.Equal(t, "", p.Description)
		assert.NotNil(t, p.CreatedAt)
		assert.NotNil(t, p.UpdatedAt)
		assert.Nil(t, p.DeletedAt)
	})

	t.Run("test update existing no name", func(t *testing.T) {
		p := Project{
			Description: "Foo Test Bar",
		}
		p.ID = uint(1)

		err := db.Save(&p).Error

		assert.Error(t, err)
	})
}
