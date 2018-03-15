package main

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const expected = `{"z":"4h30m0s","proto":"asdf","call":"","cacert":"","cert":"","key":"","insecure":false,"n":0,"c":0,"q":0,"t":0,"d":"","D":"","m":"","M":"","o":"oval","O":"","cpus":0}`

func TestConfig_MarshalJSON(t *testing.T) {
	z, _ := time.ParseDuration("4h30m")
	c := &Config{Proto: "asdf", Z: z, Format: "oval"}
	cJSON, err := json.Marshal(&c)
	assert.NoError(t, err)
	assert.Equal(t, expected, string(cJSON))
}

func TestConfig_UnmarshalJSON(t *testing.T) {
	c := &Config{}
	err := json.Unmarshal([]byte(expected), &c)
	z, _ := time.ParseDuration("4h30m")
	ec := &Config{Proto: "asdf", Z: z, Format: "oval"}

	assert.NoError(t, err)
	assert.Equal(t, ec, c)
	assert.Equal(t, ec.Z.String(), c.Z.String())
}
