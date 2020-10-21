package runner

import (
	"strconv"
	"testing"
	"time"

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
		{
			"invalid max-duration",
			&Config{},
			false,
		},
		{
			"invalid stream-interval",
			&Config{},
			false,
		},
		{
			"invalid timeout",
			&Config{},
			false,
		},
		{
			"valid",
			&Config{
				Insecure:    true,
				ImportPaths: []string{"/home/user/pb/grpcbin"},
				Proto:       "grpcbin.proto",
				Call:        "grpcbin.GRPCBin.DummyUnary",
				Host:        "127.0.0.1:9000",
				Z:           Duration(20 * time.Second),
				X:           Duration(60 * time.Second),
				SI:          Duration(25 * time.Second),
				Timeout:     Duration(30 * time.Second),
				N:           200,
				C:           50,
				Connections: 1,
				ZStop:       "close",
				Data: map[string]interface{}{
					"f_strings": []interface{}{"123", "456"},
				},
				Format:       "summary",
				DialTimeout:  Duration(10 * time.Second),
				LoadSchedule: "const",
				CSchedule:    "const",
				CStart:       1,
			},
			true,
		},
	}

	for i, tt := range tests {
		t.Run("toml "+tt.name, func(t *testing.T) {
			var actual Config
			cfgPath := "../testdata/config/config" + strconv.Itoa(i) + ".toml"
			err := LoadConfig(cfgPath, &actual)
			if tt.ok {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, &actual)
			} else {
				assert.Error(t, err)
			}
		})

		t.Run("json "+tt.name, func(t *testing.T) {
			var actual Config
			cfgPath := "../testdata/config/config" + strconv.Itoa(i) + ".toml"
			err := LoadConfig(cfgPath, &actual)
			if tt.ok {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, &actual)
			} else {
				assert.Error(t, err)
			}
		})

		t.Run("yaml "+tt.name, func(t *testing.T) {
			var actual Config
			cfgPath := "../testdata/config/config" + strconv.Itoa(i) + ".yaml"
			err := LoadConfig(cfgPath, &actual)
			if tt.ok {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, &actual)
			} else {
				assert.Error(t, err)
			}
		})
	}
}
