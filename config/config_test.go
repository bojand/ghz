package config

import (
	"encoding/json"
	"io/ioutil"
	"math"
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const expected = `{"proto":"asdf","protoset":"","call":"","cert":"","cName":"","n":0,"c":0,"q":0,"t":0,"D":"","M":"","o":"","O":"oval","host":"","T":0,"L":0,"cpus":0,"z":"4h30m0s","x":""}`

func TestConfig_MarshalJSON(t *testing.T) {
	z, _ := time.ParseDuration("4h30m")
	c := Config{Proto: "asdf", Z: z, Format: "oval"}
	cJSON, err := json.Marshal(&c)
	assert.NoError(t, err)
	assert.Equal(t, expected, string(cJSON))
}

func TestConfig_UnmarshalJSON(t *testing.T) {
	t.Run("duration", func(t *testing.T) {
		c := Config{}
		err := json.Unmarshal([]byte(expected), &c)
		z, _ := time.ParseDuration("4h30m")
		ec := Config{Proto: "asdf", Z: z, Format: "oval"}

		assert.NoError(t, err)
		assert.Equal(t, ec, c)
		assert.Equal(t, ec.Z.String(), c.Z.String())
	})

	t.Run("data not present", func(t *testing.T) {
		jsonStr := `{"proto":"protofile", "call":"someCall"}`
		c := Config{}
		err := json.Unmarshal([]byte(jsonStr), &c)
		ec := Config{Proto: "protofile", Call: "someCall"}

		assert.NoError(t, err)
		assert.Equal(t, ec, c)
		assert.Nil(t, c.Data)
	})

	t.Run("data empty string should fail", func(t *testing.T) {
		jsonStr := `{"proto":"pf", "call":"sc", "d":""}`
		c := Config{}
		err := json.Unmarshal([]byte(jsonStr), &c)

		assert.Error(t, err)
		assert.Equal(t, "Unsupported type for Data", err.Error())
	})

	t.Run("data string should fail", func(t *testing.T) {
		jsonStr := `{"proto":"pf", "call":"sc", "d":"bob"}`
		c := Config{}
		err := json.Unmarshal([]byte(jsonStr), &c)

		assert.Error(t, err)
		assert.Equal(t, "Unsupported type for Data", err.Error())
	})

	t.Run("data array of strings should fail", func(t *testing.T) {
		jsonStr := `{"proto":"pf", "call":"sc", "d":["bob", "kate"]}`
		c := Config{}
		err := json.Unmarshal([]byte(jsonStr), &c)

		assert.Error(t, err)
		assert.Equal(t, "Data array contains unsupported type", err.Error())
	})

	t.Run("data empty object", func(t *testing.T) {
		jsonStr := `{"proto":"pf", "call":"sc", "d":{}}`
		c := Config{}
		err := json.Unmarshal([]byte(jsonStr), &c)

		assert.NoError(t, err)
		assert.NotNil(t, c.Data)
		assert.Empty(t, c.Data)
	})

	t.Run("data empty array", func(t *testing.T) {
		jsonStr := `{"proto":"pf", "call":"sc", "d":[]}`
		c := Config{}
		err := json.Unmarshal([]byte(jsonStr), &c)

		assert.Error(t, err)
		assert.Equal(t, "Data array must not be empty", err.Error())
	})

	t.Run("data valid object", func(t *testing.T) {
		jsonStr := `{"proto":"pf", "call":"sc", "d":{"name":"bob"}}`
		c := Config{}
		err := json.Unmarshal([]byte(jsonStr), &c)

		assert.NoError(t, err)

		expected := make(map[string]interface{})
		expected["name"] = "bob"
		assert.Equal(t, expected, c.Data)
	})

	t.Run("data valid array of objects", func(t *testing.T) {
		jsonStr := `{"proto":"pf", "call":"sc", "d":[{"name":"bob"}, {"name":"kate"}]}`
		c := Config{}
		err := json.Unmarshal([]byte(jsonStr), &c)

		assert.NoError(t, err)
		assert.NotNil(t, c.Data)
		assert.NotEmpty(t, c.Data)
		assert.Len(t, c.Data, 2)

		expected1 := make(map[string]interface{})
		expected2 := make(map[string]interface{})
		expected1["name"] = "bob"
		expected2["name"] = "kate"

		actual, ok := c.Data.([]interface{})
		assert.True(t, ok)
		assert.NotNil(t, actual)

		assert.Equal(t, expected1, actual[0])
		assert.Equal(t, expected2, actual[1])
	})

	t.Run("data valid array of empty objects", func(t *testing.T) {
		jsonStr := `{"proto":"pf", "call":"sc", "d":[{},{},{}]}`
		c := Config{}
		err := json.Unmarshal([]byte(jsonStr), &c)

		assert.NoError(t, err)
		assert.NotNil(t, c.Data)
		assert.NotEmpty(t, c.Data)
		assert.Len(t, c.Data, 3)

		expected := make(map[string]interface{})

		actual, ok := c.Data.([]interface{})
		assert.True(t, ok)
		assert.NotNil(t, actual)

		assert.Equal(t, expected, actual[0])
		assert.Equal(t, expected, actual[1])
		assert.Equal(t, expected, actual[2])
	})

	t.Run("data valid array of mixed objects", func(t *testing.T) {
		jsonStr := `{"proto":"pf", "call":"sc", "d":[{},{"name":"bob"},{}]}`
		c := Config{}
		err := json.Unmarshal([]byte(jsonStr), &c)

		assert.NoError(t, err)
		assert.NotNil(t, c.Data)
		assert.NotEmpty(t, c.Data)
		assert.Len(t, c.Data, 3)

		expected := make(map[string]interface{})
		expected2 := make(map[string]interface{})
		expected2["name"] = "bob"

		actual, ok := c.Data.([]interface{})
		assert.True(t, ok)
		assert.NotNil(t, actual)

		assert.Equal(t, expected, actual[0])
		assert.Equal(t, expected2, actual[1])
		assert.Equal(t, expected, actual[2])
	})

	t.Run("data mixed with invalid should fail", func(t *testing.T) {
		jsonStr := `{"proto":"pf", "call":"sc", "d":[{"name":"bob"},"kate",{"name":"kate"}]}`
		c := Config{}
		err := json.Unmarshal([]byte(jsonStr), &c)

		assert.Error(t, err)
		assert.Equal(t, "Data array contains unsupported type", err.Error())
	})

	t.Run("invalid metadata array", func(t *testing.T) {
		jsonStr := `{"proto":"pf", "call":"sc", "d":[{"name":"bob"},{"name":"kate"}],"m":["requestId","1234"]}`
		c := Config{}
		err := json.Unmarshal([]byte(jsonStr), &c)

		assert.Error(t, err)
	})

	t.Run("invalid metadata shape", func(t *testing.T) {
		jsonStr := `{"proto":"pf", "call":"sc", "d":[{"name":"bob"},{"name":"kate"}],"m":{"requestId":{"n":"1234"}}}`
		c := Config{}
		err := json.Unmarshal([]byte(jsonStr), &c)

		assert.Error(t, err)
	})

	t.Run("invalid metadata type", func(t *testing.T) {
		jsonStr := `{"proto":"pf", "call":"sc", "d":[{"name":"bob"},{"name":"kate"}],"m":{"requestId":1234}}`
		c := Config{}
		err := json.Unmarshal([]byte(jsonStr), &c)

		assert.Error(t, err)
	})

	t.Run("metadata", func(t *testing.T) {
		jsonStr := `{"proto":"pf", "call":"sc", "d":[{"name":"bob"},{"name":"kate"}],"m":{"requestId":"1234"}}`
		c := Config{}
		err := json.Unmarshal([]byte(jsonStr), &c)

		assert.NoError(t, err)
		assert.NotNil(t, c.Data)
		assert.NotEmpty(t, c.Data)
		assert.Len(t, c.Data, 2)

		expected1 := make(map[string]interface{})
		expected2 := make(map[string]interface{})
		expected1["name"] = "bob"
		expected2["name"] = "kate"

		actual, ok := c.Data.([]interface{})
		assert.True(t, ok)
		assert.NotNil(t, actual)

		assert.Equal(t, expected1, actual[0])
		assert.Equal(t, expected2, actual[1])

		assert.Equal(t, "pf", c.Proto)
		assert.Equal(t, "sc", c.Call)

		expectedMD := make(map[string]string, 1)
		expectedMD["requestId"] = "1234"
		assert.Equal(t, expectedMD, *c.Metadata)
	})
}

func TestConfig_Default(t *testing.T) {
	c := &Config{}
	c.Default()

	assert.Equal(t, c.N, 200)
	assert.Equal(t, c.C, 50)
	assert.Equal(t, c.CPUs, runtime.GOMAXPROCS(-1))
}

func TestConfig_ReadConfig(t *testing.T) {
	os.Chdir("../testdata/")

	t.Run("grpcannon.json", func(t *testing.T) {
		c, err := ReadConfig("../testdata/grpcannon.json")

		assert.NoError(t, err)
		assert.NotNil(t, c)

		if c != nil {
			data := make(map[string]interface{})
			data["name"] = "mydata"

			ec := Config{
				Proto:         "my.proto",
				Call:          "mycall",
				Data:          data,
				Cert:          "mycert",
				CName:         "localhost",
				N:             200,
				C:             50,
				QPS:           0,
				Z:             0,
				Timeout:       20,
				DataPath:      "",
				MetadataPath:  "",
				Format:        "",
				Output:        "",
				Host:          "",
				DialTimeout:   10,
				KeepaliveTime: 0,
				CPUs:          runtime.GOMAXPROCS(-1),
				ImportPaths:   []string{"/path/to/protos", "."}}

			assert.Equal(t, ec, *c)
			assert.Equal(t, ec.Data, c.Data)
		}
	})

	t.Run("cfgpath.json", func(t *testing.T) {
		c, err := ReadConfig("../testdata/cfgpath.json")

		assert.NoError(t, err)
		assert.NotNil(t, c)

		if c != nil {
			data := make(map[string]interface{})
			data["name"] = "mydata"

			metaData := make(map[string]string)
			metaData["requestId"] = "12345"

			ec := Config{
				Proto:         "my.proto",
				Call:          "mycall",
				Data:          data,
				Cert:          "mycert",
				CName:         "localhost",
				N:             200,
				C:             50,
				QPS:           0,
				Z:             0,
				Timeout:       20,
				Metadata:      &metaData,
				DataPath:      "./data.json",
				MetadataPath:  "./metadata.json",
				Format:        "",
				Output:        "",
				Host:          "",
				DialTimeout:   10,
				KeepaliveTime: 0,
				CPUs:          runtime.GOMAXPROCS(-1),
				ImportPaths:   []string{"/path/to/protos", "."}}

			assert.Equal(t, ec, *c)
			assert.Equal(t, ec.Data, c.Data)
			assert.Equal(t, ec.Metadata, c.Metadata)

			actualMD := *c.Metadata
			val := actualMD["requestId"]
			assert.Equal(t, "12345", val)
		}
	})
}

func TestConfig_Validate(t *testing.T) {
	t.Run("missing proto", func(t *testing.T) {
		c := &Config{}
		err := c.Validate()
		assert.Equal(t, "Proto or Protoset required", err.Error())
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

func TestConfig_initData(t *testing.T) {
	t.Run("when empty", func(t *testing.T) {
		c := &Config{}
		err := c.initData()
		assert.Equal(t, "No data specified", err.Error())
	})

	t.Run("with map data", func(t *testing.T) {
		data := make(map[string]interface{})
		data["name"] = "mydata"
		c := &Config{Data: &data}
		err := c.initData()
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
		err = c.initData()
		assert.NoError(t, err)
		assert.Equal(t, c.Data, data)
	})
}

func TestConfig_initDurations(t *testing.T) {
	t.Run("with Z specified should set N to max int", func(t *testing.T) {
		dur, _ := time.ParseDuration("4h30m")
		c := &Config{N: 500, Z: dur}
		c.initDurations()
		assert.Equal(t, c.N, math.MaxInt32)
	})

	t.Run("with X specified should set Z to X and keep N", func(t *testing.T) {
		dur, _ := time.ParseDuration("4h30m")
		c := &Config{N: 500, X: dur}
		c.initDurations()
		assert.Equal(t, c.N, 500)
		assert.Equal(t, c.X, dur)
		assert.Equal(t, c.Z, dur)
	})

	t.Run("with X and Z specified should set Z to X and keep N", func(t *testing.T) {
		dur, _ := time.ParseDuration("4m")
		dur2, _ := time.ParseDuration("5m")
		c := &Config{N: 500, X: dur, Z: dur2}
		c.initDurations()
		assert.Equal(t, c.N, 500)
		assert.Equal(t, c.X, dur)
		assert.Equal(t, c.Z, dur)
	})
}
