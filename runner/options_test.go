package runner

import (
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRunConfig_newRunConfig(t *testing.T) {
	t.Run("fail with empty call", func(t *testing.T) {
		c, err := newConfig("  ", "localhost:50050")

		assert.Error(t, err)
		assert.Nil(t, c)
	})

	t.Run("fail with empty host ", func(t *testing.T) {
		c, err := newConfig("  call ", "   ")

		assert.Error(t, err)
		assert.Nil(t, c)
	})

	t.Run("fail with invalid extension", func(t *testing.T) {
		c, err := newConfig("call", "localhost:50050",
			WithProtoFile("testdata/data.bin", []string{}),
		)

		assert.Error(t, err)
		assert.Nil(t, c)
	})

	t.Run("without any options should have defaults", func(t *testing.T) {
		c, err := newConfig("  call  ", "  localhost:50050  ",
			WithProtoFile("testdata/data.proto", []string{}),
		)

		assert.NoError(t, err)

		assert.Equal(t, "call", c.call)
		assert.Equal(t, "localhost:50050", c.host)
		assert.Equal(t, false, c.insecure)
		assert.Equal(t, 200, c.n)
		assert.Equal(t, 50, c.c)
		assert.Equal(t, 0, c.qps)
		assert.Equal(t, false, c.binary)
		assert.Equal(t, time.Duration(0), c.z)
		assert.Equal(t, time.Duration(0), c.keepaliveTime)
		assert.Equal(t, time.Duration(20*time.Second), c.timeout)
		assert.Equal(t, time.Duration(10*time.Second), c.dialTimeout)
		assert.Equal(t, runtime.GOMAXPROCS(-1), c.cpus)
		assert.Empty(t, c.name)
		assert.Empty(t, c.data)
		assert.False(t, c.binary)
		assert.Empty(t, c.metadata)
		assert.Equal(t, "testdata/data.proto", string(c.proto))
		assert.Equal(t, "", string(c.protoset))
		assert.Equal(t, []string{"testdata", "."}, c.importPaths)
		assert.Equal(t, c.enableCompression, false)
	})

	t.Run("with options", func(t *testing.T) {
		c, err := newConfig(
			"call", "localhost:50050",
			WithInsecure(true),
			WithTotalRequests(100),
			WithConcurrency(20),
			WithQPS(5),
			WithRunDuration(time.Duration(5*time.Minute)),
			WithKeepalive(time.Duration(60*time.Second)),
			WithTimeout(time.Duration(10*time.Second)),
			WithDialTimeout(time.Duration(30*time.Second)),
			WithName("asdf"),
			WithCPUs(4),
			WithDataFromJSON(`{"name":"bob"}`),
			WithMetadataFromJSON(`{"request-id":"123"}`),
			WithProtoFile("testdata/data.proto", []string{"/home/protos"}),
		)

		assert.NoError(t, err)

		assert.Equal(t, "call", c.call)
		assert.Equal(t, "localhost:50050", c.host)
		assert.Equal(t, true, c.insecure)
		assert.Equal(t, 100, c.n)
		assert.Equal(t, 20, c.c)
		assert.Equal(t, 5, c.qps)
		assert.Equal(t, false, c.binary)
		assert.Equal(t, time.Duration(5*time.Minute), c.z)
		assert.Equal(t, time.Duration(60*time.Second), c.keepaliveTime)
		assert.Equal(t, time.Duration(10*time.Second), c.timeout)
		assert.Equal(t, time.Duration(30*time.Second), c.dialTimeout)
		assert.Equal(t, 4, c.cpus)
		assert.False(t, c.binary)
		assert.Equal(t, "asdf", c.name)
		assert.Equal(t, `{"name":"bob"}`, string(c.data))
		assert.Equal(t, `{"request-id":"123"}`, string(c.metadata))
		assert.Equal(t, "testdata/data.proto", string(c.proto))
		assert.Equal(t, "", string(c.protoset))
		assert.Equal(t, []string{"testdata", ".", "/home/protos"}, c.importPaths)
		assert.Equal(t, c.enableCompression, false)
	})

	t.Run("with binary data, protoset and metadata file", func(t *testing.T) {
		c, err := newConfig(
			"call", "localhost:50050",
			WithCertificate("../testdata/localhost.crt", "../testdata/localhost.key"),
			WithServerNameOverride("cname"),
			WithAuthority("someauth"),
			WithTotalRequests(100),
			WithConcurrency(20),
			WithQPS(5),
			WithRunDuration(time.Duration(5*time.Minute)),
			WithKeepalive(time.Duration(60*time.Second)),
			WithTimeout(time.Duration(10*time.Second)),
			WithDialTimeout(time.Duration(30*time.Second)),
			WithName("asdf"),
			WithCPUs(4),
			WithBinaryData([]byte("asdf1234foobar")),
			WithMetadataFromFile("../testdata/metadata.json"),
			WithProtoset("testdata/bundle.protoset"),
		)

		assert.NoError(t, err)

		assert.Equal(t, "call", c.call)
		assert.Equal(t, "localhost:50050", c.host)
		assert.Equal(t, false, c.insecure)
		assert.Equal(t, "../testdata/localhost.crt", c.cert)
		assert.Equal(t, "../testdata/localhost.key", c.key)
		assert.Equal(t, "cname", c.cname)
		assert.Equal(t, "someauth", c.authority)
		assert.Equal(t, 100, c.n)
		assert.Equal(t, 20, c.c)
		assert.Equal(t, 5, c.qps)
		assert.Equal(t, true, c.binary)
		assert.Equal(t, time.Duration(5*time.Minute), c.z)
		assert.Equal(t, time.Duration(60*time.Second), c.keepaliveTime)
		assert.Equal(t, time.Duration(10*time.Second), c.timeout)
		assert.Equal(t, time.Duration(30*time.Second), c.dialTimeout)
		assert.Equal(t, 4, c.cpus)
		assert.Equal(t, "asdf", c.name)
		assert.Equal(t, []byte("asdf1234foobar"), c.data)
		assert.Equal(t, `{"request-id": "{{.RequestNumber}}"}`, string(c.metadata))
		assert.Equal(t, "", string(c.proto))
		assert.Equal(t, "testdata/bundle.protoset", string(c.protoset))
		assert.NotNil(t, c.creds)
		assert.Equal(t, c.enableCompression, false)
	})

	t.Run("with data interface and metadata map", func(t *testing.T) {
		type dataStruct struct {
			Name   string   `json:"name"`
			Age    int      `json:"age"`
			Fruits []string `json:"fruits"`
		}

		d := &dataStruct{
			Name:   "bob",
			Age:    11,
			Fruits: []string{"apple", "peach", "pear"}}

		md := make(map[string]string)
		md["token"] = "foobar"
		md["request-id"] = "123"

		tags := make(map[string]string)
		tags["env"] = "staging"
		tags["created by"] = "joe developer"

		rmd := make(map[string]string)
		rmd["auth"] = "bizbaz"

		c, err := newConfig(
			"call", "localhost:50050",
			WithProtoFile("testdata/data.proto", []string{}),
			WithCertificate("../testdata/localhost.crt", "../testdata/localhost.key"),
			WithInsecure(true),
			WithTotalRequests(100),
			WithConcurrency(20),
			WithQPS(5),
			WithRunDuration(time.Duration(5*time.Minute)),
			WithKeepalive(time.Duration(60*time.Second)),
			WithTimeout(time.Duration(10*time.Second)),
			WithDialTimeout(time.Duration(30*time.Second)),
			WithName("asdf"),
			WithCPUs(4),
			WithData(d),
			WithMetadata(&md),
			WithTags(&tags),
			WithReflectionMetadata(&rmd),
		)

		assert.NoError(t, err)

		assert.Equal(t, "call", c.call)
		assert.Equal(t, "localhost:50050", c.host)
		assert.Equal(t, true, c.insecure)
		assert.Equal(t, "../testdata/localhost.crt", c.cert)
		assert.Equal(t, "../testdata/localhost.key", c.key)
		assert.Equal(t, 100, c.n)
		assert.Equal(t, 20, c.c)
		assert.Equal(t, 5, c.qps)
		assert.Equal(t, false, c.binary)
		assert.Equal(t, time.Duration(5*time.Minute), c.z)
		assert.Equal(t, time.Duration(60*time.Second), c.keepaliveTime)
		assert.Equal(t, time.Duration(10*time.Second), c.timeout)
		assert.Equal(t, time.Duration(30*time.Second), c.dialTimeout)
		assert.Equal(t, 4, c.cpus)
		assert.Equal(t, "asdf", c.name)
		assert.Equal(t, `{"name":"bob","age":11,"fruits":["apple","peach","pear"]}`, string(c.data))
		assert.Equal(t, `{"request-id":"123","token":"foobar"}`, string(c.metadata))
		assert.Equal(t, `{"created by":"joe developer","env":"staging"}`, string(c.tags))
		assert.Equal(t, "testdata/data.proto", string(c.proto))
		assert.Equal(t, "", string(c.protoset))
		assert.Equal(t, []string{"testdata", "."}, c.importPaths)
		assert.NotNil(t, c.creds)
		assert.Equal(t, map[string]string{"auth": "bizbaz"}, *c.rmd)
		assert.Equal(t, c.enableCompression, false)
	})

	t.Run("with binary data from file", func(t *testing.T) {
		c, err := newConfig("call", "localhost:50050",
			WithProtoFile("testdata/data.proto", []string{}),
			WithBinaryDataFromFile("../testdata/hello_request_data.bin"),
		)

		assert.NoError(t, err)

		assert.Equal(t, "call", c.call)
		assert.Equal(t, "localhost:50050", c.host)
		assert.Equal(t, false, c.insecure)
		assert.Equal(t, 200, c.n)
		assert.Equal(t, 50, c.c)
		assert.Equal(t, 0, c.qps)
		assert.Equal(t, time.Duration(0), c.z)
		assert.Equal(t, time.Duration(0), c.keepaliveTime)
		assert.Equal(t, time.Duration(20*time.Second), c.timeout)
		assert.Equal(t, time.Duration(10*time.Second), c.dialTimeout)
		assert.Equal(t, runtime.GOMAXPROCS(-1), c.cpus)
		assert.Empty(t, c.name)
		assert.NotEmpty(t, c.data)
		assert.True(t, c.binary)
		assert.Empty(t, c.metadata)
		assert.Equal(t, "testdata/data.proto", string(c.proto))
		assert.Equal(t, "", string(c.protoset))
		assert.Equal(t, []string{"testdata", "."}, c.importPaths)
		assert.Equal(t, c.enableCompression, false)
	})

	t.Run("with data from file", func(t *testing.T) {
		c, err := newConfig("call", "localhost:50050",
			WithProtoFile("testdata/data.proto", []string{}),
			WithDataFromFile("../testdata/data.json"),
		)

		assert.NoError(t, err)

		assert.Equal(t, "call", c.call)
		assert.Equal(t, "localhost:50050", c.host)
		assert.Equal(t, false, c.insecure)
		assert.Equal(t, 200, c.n)
		assert.Equal(t, 50, c.c)
		assert.Equal(t, 0, c.qps)
		assert.Equal(t, false, c.binary)
		assert.Equal(t, time.Duration(0), c.z)
		assert.Equal(t, time.Duration(0), c.keepaliveTime)
		assert.Equal(t, time.Duration(20*time.Second), c.timeout)
		assert.Equal(t, time.Duration(10*time.Second), c.dialTimeout)
		assert.Equal(t, runtime.GOMAXPROCS(-1), c.cpus)
		assert.Empty(t, c.name)
		assert.NotEmpty(t, c.data)
		assert.False(t, c.binary)
		assert.Empty(t, c.metadata)
		assert.Equal(t, "testdata/data.proto", string(c.proto))
		assert.Equal(t, "", string(c.protoset))
		assert.Equal(t, []string{"testdata", "."}, c.importPaths)
		assert.Equal(t, c.enableCompression, false)
	})

	t.Run("with data from reader", func(t *testing.T) {

		file, _ := os.Open("../testdata/data.json")
		defer file.Close()

		c, err := newConfig("call", "localhost:50050",
			WithProtoFile("testdata/data.proto", []string{}),
			WithDataFromReader(file),
		)

		assert.NoError(t, err)

		assert.Equal(t, "call", c.call)
		assert.Equal(t, "localhost:50050", c.host)
		assert.Equal(t, false, c.insecure)
		assert.Equal(t, 200, c.n)
		assert.Equal(t, 50, c.c)
		assert.Equal(t, 0, c.qps)
		assert.Equal(t, 1, c.nConns)
		assert.Equal(t, false, c.binary)
		assert.Equal(t, time.Duration(0), c.z)
		assert.Equal(t, time.Duration(0), c.keepaliveTime)
		assert.Equal(t, time.Duration(20*time.Second), c.timeout)
		assert.Equal(t, time.Duration(10*time.Second), c.dialTimeout)
		assert.Equal(t, runtime.GOMAXPROCS(-1), c.cpus)
		assert.Empty(t, c.name)
		assert.NotEmpty(t, c.data)
		assert.False(t, c.binary)
		assert.Empty(t, c.metadata)
		assert.Equal(t, "testdata/data.proto", string(c.proto))
		assert.Equal(t, "", string(c.protoset))
		assert.Equal(t, []string{"testdata", "."}, c.importPaths)
		assert.Equal(t, c.enableCompression, false)
	})

	t.Run("with connections", func(t *testing.T) {

		file, _ := os.Open("../testdata/data.json")
		defer file.Close()

		c, err := newConfig("call", "localhost:50050",
			WithProtoFile("testdata/data.proto", []string{}),
			WithDataFromReader(file),
			WithConnections(5),
		)

		assert.NoError(t, err)

		assert.Equal(t, "call", c.call)
		assert.Equal(t, "localhost:50050", c.host)
		assert.Equal(t, false, c.insecure)
		assert.Equal(t, 200, c.n)
		assert.Equal(t, 50, c.c)
		assert.Equal(t, 0, c.qps)
		assert.Equal(t, 5, c.nConns)
		assert.Equal(t, false, c.binary)
		assert.Equal(t, time.Duration(0), c.z)
		assert.Equal(t, time.Duration(0), c.keepaliveTime)
		assert.Equal(t, time.Duration(20*time.Second), c.timeout)
		assert.Equal(t, time.Duration(10*time.Second), c.dialTimeout)
		assert.Equal(t, runtime.GOMAXPROCS(-1), c.cpus)
		assert.Empty(t, c.name)
		assert.NotEmpty(t, c.data)
		assert.False(t, c.binary)
		assert.Empty(t, c.metadata)
		assert.Equal(t, "testdata/data.proto", string(c.proto))
		assert.Equal(t, "", string(c.protoset))
		assert.Equal(t, []string{"testdata", "."}, c.importPaths)
		assert.Equal(t, c.enableCompression, false)
	})

	t.Run("with invalid connections > concurrency", func(t *testing.T) {

		file, _ := os.Open("../testdata/data.json")
		defer file.Close()

		c, err := newConfig("call", "localhost:50050",
			WithProtoFile("testdata/data.proto", []string{}),
			WithDataFromReader(file),
			WithConcurrency(5),
			WithConnections(6),
		)

		assert.Error(t, err)
		assert.Nil(t, c)
	})
}
