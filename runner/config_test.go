package runner

import (
	"strconv"
	"testing"

	"github.com/jinzhu/configor"
	"github.com/stretchr/testify/assert"
)

func TestConfig_Load(t *testing.T) {
	var tests = []struct {
		name     string
		expected *Config
		ok       bool
	}{
		{
			"invalid data",
			&Config{},
			false,
		},
		{
			"invalid duration",
			&Config{},
			false,
		},
	}

	for i, tt := range tests {
		t.Run("toml "+tt.name, func(t *testing.T) {
			var actual Config
			cfgPath := "../testdata/config/config" + strconv.Itoa(i) + ".toml"
			err := configor.Load(&actual, cfgPath)
			if tt.ok {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, actual)
			} else {
				assert.Error(t, err)
			}
		})

		t.Run("json "+tt.name, func(t *testing.T) {
			var actual Config
			cfgPath := "../testdata/config/config" + strconv.Itoa(i) + ".json"
			err := configor.Load(&actual, cfgPath)
			if tt.ok {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, actual)
			} else {
				assert.Error(t, err)
			}
		})
	}
}
