package database

import (
	"os"
	"strconv"
	"testing"

	"github.com/bojand/ghz/web/model"
	"github.com/stretchr/testify/assert"
)

func TestDatabase_CreateProject(t *testing.T) {
	os.Remove(dbName)

	defer os.Remove(dbName)

	db, err := New("sqlite3", dbName, false)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	defer db.Close()

	t.Run("test new", func(t *testing.T) {
		p := model.Project{
			Name:        "Test Proj 111 ",
			Description: "Test Description Asdf ",
		}

		err := db.CreateProject(&p)

		assert.NoError(t, err)
		assert.NotZero(t, p.ID)
		assert.Equal(t, "Test Proj 111", p.Name)
		assert.Equal(t, "Test Description Asdf", p.Description)
		assert.Equal(t, model.StatusOK, p.Status)
		assert.NotZero(t, p.CreatedAt)
		assert.NotZero(t, p.UpdatedAt)

		p2 := new(model.Project)
		err = db.DB.First(p2, p.ID).Error

		assert.NoError(t, err)
		assert.Equal(t, "Test Proj 111", p2.Name)
		assert.Equal(t, "Test Description Asdf", p2.Description)
		assert.Equal(t, model.StatusOK, p2.Status)
		assert.NotZero(t, p2.CreatedAt)
		assert.NotZero(t, p2.UpdatedAt)
	})

	t.Run("test new with empty name", func(t *testing.T) {
		p := model.Project{
			Description: "Test Description Asdf 2",
		}

		err := db.CreateProject(&p)

		assert.NoError(t, err)
		assert.NotZero(t, p.ID)
		assert.NotEmpty(t, p.Name)
		assert.Equal(t, "Test Description Asdf 2", p.Description)
		assert.Equal(t, model.StatusOK, p.Status)
		assert.NotZero(t, p.CreatedAt)
		assert.NotZero(t, p.UpdatedAt)

		p2 := new(model.Project)
		err = db.DB.First(p2, p.ID).Error

		assert.NoError(t, err)
		assert.Equal(t, p.Name, p2.Name)
		assert.Equal(t, "Test Description Asdf 2", p2.Description)
		assert.Equal(t, model.StatusOK, p2.Status)
		assert.NotZero(t, p2.CreatedAt)
		assert.NotZero(t, p2.UpdatedAt)
	})

	t.Run("test new with ID", func(t *testing.T) {
		p := model.Project{
			Name:        " Foo Project ",
			Description: " Bar Desc ",
			Status:      model.StatusFail,
		}
		p.ID = 123

		err := db.CreateProject(&p)

		assert.NoError(t, err)
		assert.Equal(t, uint(123), p.ID)
		assert.Equal(t, "Foo Project", p.Name)
		assert.Equal(t, "Bar Desc", p.Description)
		assert.Equal(t, model.StatusFail, p.Status)
		assert.NotZero(t, p.CreatedAt)
		assert.NotZero(t, p.UpdatedAt)

		p2 := new(model.Project)
		err = db.DB.First(p2, p.ID).Error

		assert.NoError(t, err)
		assert.Equal(t, uint(123), p2.ID)
		assert.Equal(t, "Foo Project", p2.Name)
		assert.Equal(t, "Bar Desc", p2.Description)
		assert.Equal(t, model.StatusFail, p2.Status)
		assert.NotZero(t, p2.CreatedAt)
		assert.NotZero(t, p2.UpdatedAt)
	})

	t.Run("should fail with same ID", func(t *testing.T) {
		p := model.Project{
			Name:        "ACME",
			Description: "Lorem Ipsum",
		}
		p.ID = 123

		err := db.CreateProject(&p)

		assert.Error(t, err)
	})
}

func TestDatabase_UpdateProject(t *testing.T) {
	os.Remove(dbName)

	defer os.Remove(dbName)

	db, err := New("sqlite3", dbName, false)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	defer db.Close()

	const pid = 4444

	t.Run("create new", func(t *testing.T) {
		p := model.Project{
			Name:        "testproject124",
			Description: "asdf",
		}
		p.ID = pid

		err := db.CreateProject(&p)

		assert.Nil(t, err)
	})

	t.Run("test update existing", func(t *testing.T) {
		p := new(model.Project)
		err = db.DB.First(p, pid).Error

		p.Name = " New Name "
		p.Description = "Baz"
		p.Status = model.StatusOK

		err := db.UpdateProject(p)

		assert.NoError(t, err)

		assert.NotZero(t, p.ID)
		assert.Equal(t, "New Name", p.Name)
		assert.Equal(t, "Baz", p.Description)
		assert.Equal(t, model.StatusOK, p.Status)
		assert.NotZero(t, p.CreatedAt)
		assert.NotZero(t, p.UpdatedAt)

		p2 := new(model.Project)
		err = db.DB.First(p2, p.ID).Error

		assert.NoError(t, err)
		assert.Equal(t, uint(pid), p2.ID)
		assert.Equal(t, "New Name", p2.Name)
		assert.Equal(t, "Baz", p2.Description)
		assert.Equal(t, model.StatusOK, p2.Status)
		assert.NotZero(t, p2.CreatedAt)
		assert.NotZero(t, p2.UpdatedAt)
	})
}

func TestDatabase_UpdateProjectStatus(t *testing.T) {
	os.Remove(dbName)

	defer os.Remove(dbName)

	db, err := New("sqlite3", dbName, false)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	defer db.Close()

	const pid = 4321

	t.Run("create new", func(t *testing.T) {
		p := model.Project{
			Name:        "Test Project 124",
			Description: "Asdf",
		}
		p.ID = pid

		err := db.CreateProject(&p)

		assert.Nil(t, err)
	})

	t.Run("test update status", func(t *testing.T) {
		err := db.UpdateProjectStatus(pid, model.StatusFail)

		assert.NoError(t, err)

		p2 := new(model.Project)
		err = db.DB.First(p2, pid).Error

		assert.NoError(t, err)
		assert.Equal(t, uint(pid), p2.ID)
		assert.Equal(t, "Test Project 124", p2.Name)
		assert.Equal(t, "Asdf", p2.Description)
		assert.Equal(t, string(model.StatusFail), string(p2.Status))
		assert.NotZero(t, p2.CreatedAt)
		assert.NotZero(t, p2.UpdatedAt)
	})
}

func TestDatabase_FindProjectByID(t *testing.T) {
	os.Remove(dbName)

	defer os.Remove(dbName)

	db, err := New("sqlite3", dbName, false)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	defer db.Close()

	const pid = 321

	t.Run("create new", func(t *testing.T) {
		p := model.Project{
			Name:        " Test Project 124  ",
			Description: " Some Project Description Asdf ",
		}
		p.ID = pid

		err := db.CreateProject(&p)

		assert.Nil(t, err)
	})

	t.Run("test existing", func(t *testing.T) {
		p, err := db.FindProjectByID(321)

		assert.NoError(t, err)
		assert.Equal(t, uint(pid), p.ID)
		assert.Equal(t, "Test Project 124", p.Name)
		assert.Equal(t, "Some Project Description Asdf", p.Description)
		assert.Equal(t, model.StatusOK, p.Status)
		assert.NotZero(t, p.CreatedAt)
		assert.NotZero(t, p.UpdatedAt)
	})

	t.Run("test not found", func(t *testing.T) {
		p, err := db.FindProjectByID(22232)

		assert.Error(t, err)
		assert.Nil(t, p)
	})
}

func TestDatabase_ListProjects(t *testing.T) {
	os.Remove(dbName)

	defer os.Remove(dbName)

	db, err := New("sqlite3", dbName, false)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	defer db.Close()

	t.Run("create new projects", func(t *testing.T) {
		i := 10
		for i < 20 {
			iStr := strconv.FormatInt(int64(i), 10)
			p := model.Project{
				Name:        "Test Proj " + iStr,
				Description: "Test Description " + iStr,
			}
			err := db.CreateProject(&p)

			assert.NoError(t, err)

			i = i + 1
		}
	})

	t.Run("find all asc", func(t *testing.T) {
		ps, err := db.ListProjects(20, 0, "id", "asc")

		assert.NoError(t, err)
		assert.Len(t, ps, 10)

		assert.Equal(t, uint(1), ps[0].ID)
		assert.Equal(t, uint(10), ps[9].ID)
	})

	t.Run("find all desc", func(t *testing.T) {
		ps, err := db.ListProjects(20, 0, "id", "desc")

		assert.NoError(t, err)
		assert.Len(t, ps, 10)

		assert.Equal(t, uint(10), ps[0].ID)
		assert.Equal(t, uint(1), ps[9].ID)
	})

	t.Run("default to id on invalid param", func(t *testing.T) {
		ps, err := db.ListProjects(20, 0, "status", "asc")

		assert.NoError(t, err)
		assert.Len(t, ps, 10)

		assert.Equal(t, uint(1), ps[0].ID)
		assert.Equal(t, uint(10), ps[9].ID)
	})

	t.Run("default to desc on unknown order", func(t *testing.T) {
		ps, err := db.ListProjects(20, 0, "id", "asdf")

		assert.NoError(t, err)
		assert.Len(t, ps, 10)

		assert.Equal(t, uint(10), ps[0].ID)
		assert.Equal(t, uint(1), ps[9].ID)
	})

	t.Run("list paged name desc", func(t *testing.T) {
		ps, err := db.ListProjects(3, 0, "name", "desc")

		assert.NoError(t, err)
		assert.Len(t, ps, 3)

		for i, pr := range ps {
			nStr := strconv.FormatInt(int64(19-i), 10)
			assert.Equal(t, "Test Proj "+nStr, pr.Name)
		}
	})

	t.Run("list paged name asc", func(t *testing.T) {
		ps, err := db.ListProjects(3, 0, "name", "asc")

		assert.NoError(t, err)
		assert.Len(t, ps, 3)

		for i, pr := range ps {
			nStr := strconv.FormatInt(int64(10+i), 10)
			assert.Equal(t, "Test Proj "+nStr, pr.Name)
		}
	})

	t.Run("list paged 2 name desc", func(t *testing.T) {
		ps, err := db.ListProjects(3, 1, "name", "desc")

		assert.NoError(t, err)
		assert.Len(t, ps, 3)

		for i, pr := range ps {
			nStr := strconv.FormatInt(int64(16-i), 10)
			assert.Equal(t, "Test Proj "+nStr, pr.Name)
		}
	})

	t.Run("list paged 2 name asc", func(t *testing.T) {
		ps, err := db.ListProjects(3, 1, "name", "asc")

		assert.NoError(t, err)
		assert.Len(t, ps, 3)

		for i, pr := range ps {
			nStr := strconv.FormatInt(int64(13+i), 10)
			assert.Equal(t, "Test Proj "+nStr, pr.Name)
		}
	})
}

func TestDatabase_CountProjects(t *testing.T) {
	os.Remove(dbName)

	defer os.Remove(dbName)

	db, err := New("sqlite3", dbName, false)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	defer db.Close()

	t.Run("no projects", func(t *testing.T) {
		count, err := db.CountProjects()

		assert.NoError(t, err)
		assert.Equal(t, uint(0), count)
	})

	t.Run("create new projects", func(t *testing.T) {
		i := 1
		for i < 12 {
			iStr := strconv.FormatInt(int64(i), 10)
			p := model.Project{
				Name:        "Test Proj " + iStr,
				Description: "Test Description " + iStr,
			}
			err := db.CreateProject(&p)

			assert.NoError(t, err)

			i = i + 1
		}
	})

	t.Run("with projects", func(t *testing.T) {
		count, err := db.CountProjects()

		assert.NoError(t, err)
		assert.Equal(t, uint(11), count)
	})
}

func TestDatabase_DeleteProject(t *testing.T) {
	os.Remove(dbName)

	defer os.Remove(dbName)

	db, err := New("sqlite3", dbName, false)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	defer db.Close()

	var pid uint

	t.Run("create new", func(t *testing.T) {
		p := model.Project{
			Name:        "testproject124",
			Description: "asdf",
		}

		err := db.CreateProject(&p)

		assert.NoError(t, err)

		pid = p.ID
	})

	t.Run("DeleteProject()", func(t *testing.T) {
		p := &model.Project{}
		p.ID = pid

		err := db.DeleteProject(p)

		assert.NoError(t, err)

		p2 := new(model.Project)
		err = db.DB.First(p2, pid).Error

		assert.Error(t, err)
	})
}
