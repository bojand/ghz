package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_Read(t *testing.T) {
	var tests = []struct {
		name     string
		in       string
		expected *Config
	}{
		{"config1.toml",
			"../test/config1.toml",
			&Config{
				Server:   Server{Port: 80},
				Database: Database{Type: "sqlite3", Connection: "data/ghz.db"},
				Log:      Log{Level: "info"}}},
		{"config2.toml",
			"../test/config2.toml",
			&Config{
				Server:   Server{Port: 4321},
				Database: Database{Type: "postgres", Connection: "host=dbhost user=dbuser dbname=ghz sslmode=disable password=dbpwd"},
				Log:      Log{Level: "warn", Path: "/tmp/ghz.log"}}},
		{"config3.toml",
			"../test/config3.toml",
			&Config{
				Server:   Server{Port: 3000},
				Database: Database{Type: "postgres", Connection: "host=localhost port=5432 dbname=ghz sslmode=disable"},
				Log:      Log{Level: "debug", Path: ""}}},
		{"config2.json",
			"../test/config2.json",
			&Config{
				Server:   Server{Port: 4321},
				Database: Database{Type: "postgres", Connection: "host=dbhost user=dbuser dbname=ghz sslmode=disable password=dbpwd"},
				Log:      Log{Level: "warn", Path: "/tmp/ghz.log"}}},
		{"config2.yml",
			"../test/config2.yml",
			&Config{
				Server:   Server{Port: 4321},
				Database: Database{Type: "postgres", Connection: "host=dbhost user=dbuser dbname=ghz sslmode=disable password=dbpwd"},
				Log:      Log{Level: "warn", Path: "/tmp/ghz.log"}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := Read(tt.in)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, actual)
		})
	}
}
