package config

import (
	"encoding/json"
	"io/ioutil"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const expected = `{"proto":"asdf","call":"","cert":"","n":0,"c":0,"qps":0,"timeout":0,"dataPath":"","metadataPath":"","format":"oval","output":"","host":"","cpus":0,"z":"4h30m0s"}`

func TestConfig_MarshalJSON(t *testing.T) {
	z, _ := time.ParseDuration("4h30m")
	c := Config{Proto: "asdf", Z: z, Format: "oval"}
	cJSON, err := json.Marshal(&c)
	assert.NoError(t, err)
	assert.Equal(t, expected, string(cJSON))
}

func TestConfig_UnmarshalJSON(t *testing.T) {
	c := Config{}
	err := json.Unmarshal([]byte(expected), &c)
	z, _ := time.ParseDuration("4h30m")
	ec := Config{Proto: "asdf", Z: z, Format: "oval"}

	assert.NoError(t, err)
	assert.Equal(t, ec, c)
	assert.Equal(t, ec.Z.String(), c.Z.String())
}

func TestConfig_Default(t *testing.T) {
	c := &Config{}
	c.Default()

	assert.Equal(t, c.N, 200)
	assert.Equal(t, c.C, 50)
	assert.Equal(t, c.CPUs, runtime.GOMAXPROCS(-1))
}

func TestConfig_ReadConfig(t *testing.T) {
	c, err := ReadConfig("../testdata/grpcannon.json")

	data := make(map[string]interface{})
	data["name"] = "mydata"

	ec := Config{
		Proto:        "my.proto",
		Call:         "mycall",
		Data:         &data,
		Cert:         "",
		N:            200,
		C:            50,
		QPS:          0,
		Z:            0,
		DataPath:     "",
		MetadataPath: "",
		Format:       "",
		Output:       "",
		Host:         "",
		CPUs:         runtime.GOMAXPROCS(-1),
		ImportPaths:  []string{"/path/to/protos"}}

	assert.NoError(t, err)

	assert.Equal(t, ec, *c)
	assert.Equal(t, ec.Data, c.Data)
}

func TestConfig_Validate(t *testing.T) {
	t.Run("missing proto", func(t *testing.T) {
		c := &Config{}
		err := c.Validate()
		assert.Equal(t, "proto: is required", err.Error())
	})

	t.Run("invalid proto", func(t *testing.T) {
		c := &Config{Proto: "asdf"}
		err := c.Validate()
		assert.Equal(t, "proto: must have .proto extension", err.Error())
	})

	t.Run("missing call", func(t *testing.T) {
		c := &Config{Proto: "asdf.proto"}
		err := c.Validate()
		assert.Equal(t, "call: is required", err.Error())
	})

	t.Run("N < 0", func(t *testing.T) {
		c := &Config{Proto: "asdf.proto", Call: "call", Cert: "cert", N: -1}
		err := c.Validate()
		assert.Equal(t, "n: must be at least 0", err.Error())
	})

	t.Run("C < 0", func(t *testing.T) {
		c := &Config{Proto: "asdf.proto", Call: "call", Cert: "cert", C: -1}
		err := c.Validate()
		assert.Equal(t, "c: must be at least 0", err.Error())
	})

	t.Run("QPS < 0", func(t *testing.T) {
		c := &Config{Proto: "asdf.proto", Call: "call", Cert: "cert", QPS: -1}
		err := c.Validate()
		assert.Equal(t, "q: must be at least 0", err.Error())
	})

	t.Run("T < 0", func(t *testing.T) {
		c := &Config{Proto: "asdf.proto", Call: "call", Cert: "cert", Timeout: -1}
		err := c.Validate()
		assert.Equal(t, "t: must be at least 0", err.Error())
	})

	t.Run("CPUs < 0", func(t *testing.T) {
		c := &Config{Proto: "asdf.proto", Call: "call", Cert: "cert", CPUs: -1}
		err := c.Validate()
		assert.Equal(t, "cpus: must be at least 0", err.Error())
	})

	t.Run("missing data", func(t *testing.T) {
		c := &Config{Proto: "asdf.proto", Call: "call", Cert: "cert"}
		err := c.Validate()
		assert.Equal(t, "data: is required", err.Error())
	})

	t.Run("DataPath", func(t *testing.T) {
		c := &Config{Proto: "asdf.proto", Call: "call", Cert: "cert", DataPath: "asdf"}
		err := c.Validate()
		assert.NoError(t, err)
	})
}

func TestConfig_InitData(t *testing.T) {
	t.Run("when empty", func(t *testing.T) {
		c := &Config{}
		err := c.InitData()
		assert.Equal(t, "No data specified", err.Error())
	})

	t.Run("with map data", func(t *testing.T) {
		data := make(map[string]interface{})
		data["name"] = "mydata"
		c := &Config{Data: &data}
		err := c.InitData()
		assert.NoError(t, err)
		assert.Equal(t, c.Data, &data)
	})

	t.Run("with file specified", func(t *testing.T) {
		data := make(map[string]interface{})
		dat, err := ioutil.ReadFile("../testdata/data.json")
		assert.NoError(t, err)
		err = json.Unmarshal([]byte(dat), &data)
		assert.NoError(t, err)

		c := &Config{DataPath: "../testdata/data.json"}
		err = c.InitData()
		assert.NoError(t, err)
		assert.Equal(t, c.Data, &data)
	})
}
