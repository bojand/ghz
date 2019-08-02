package database

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	db, err := New("sqlite3", "test/testdb.db", false)

	assert.NotNil(t, db)
	assert.Nil(t, err)

	db.Close()

	os.Remove("test/testdb.db")
}
func TestInvalidDialect(t *testing.T) {
	_, err := New("asdf", "invalidtestdb.db", false)
	assert.NotNil(t, err)
}

func TestCreateSqliteFolder(t *testing.T) {
	// ensure path not exists
	os.RemoveAll("test/somepath")

	db, err := New("sqlite3", "test/somepath/testdb.db", false)
	assert.Nil(t, err)
	assert.DirExists(t, "test/somepath")
	db.Close()

	assert.Nil(t, os.RemoveAll("test"))
}

func TestWithAlreadyExistingSqliteFolder(t *testing.T) {
	// ensure path not exists
	err := os.RemoveAll("test/somepath")
	assert.NoError(t, err)

	err = os.MkdirAll("test/somepath", 0777)
	assert.NoError(t, err)

	db, err := New("sqlite3", "test/somepath/testdb.db", false)
	assert.Nil(t, err)
	assert.NoError(t, err)
	assert.DirExists(t, "test/somepath")
	db.Close()

	assert.Nil(t, os.RemoveAll("test"))
}
