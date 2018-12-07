package ghz

import (
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRunConfig_newRunConfig(t *testing.T) {
	t.Run("without any options should have defaults", func(t *testing.T) {
		c, err := newConfig()

		assert.NoError(t, err)

		assert.Equal(t, false, c.insecure)
		assert.Equal(t, uint(200), c.n)
		assert.Equal(t, uint(50), c.c)
		assert.Equal(t, uint(0), c.qps)
		assert.Equal(t, false, c.binary)
		assert.Equal(t, time.Duration(0), c.z)
		assert.Equal(t, time.Duration(0), c.keepaliveTime)
		assert.Equal(t, time.Duration(20*time.Second), c.timeout)
		assert.Equal(t, time.Duration(10*time.Second), c.dialTimeout)
		assert.Equal(t, runtime.GOMAXPROCS(-1), c.cpus)
		assert.NotEmpty(t, c.name)
		assert.Empty(t, c.data)
		assert.Empty(t, c.metadata)
	})

	t.Run("with options", func(t *testing.T) {
		c, err := newConfig(
			WithCertificate("certfile", "somecname"),
			WithInsecure,
			WithN(100),
			WithC(20),
			WithQPS(5),
			WithZ(time.Duration(5*time.Minute)),
			WithKeepalive(time.Duration(60*time.Second)),
			WithTimeout(time.Duration(10*time.Second)),
			WithDialTimeout(time.Duration(30*time.Second)),
			WithName("asdf"),
			WithCPUs(4),
			WithDataFromJSON(`{"name":"bob"}`),
			WithMetadataFromJSON(`{"request-id":"123"}`),
			WithProtoFile("testdata/data.proto", []string{}),
		)

		assert.NoError(t, err)

		assert.Equal(t, true, c.insecure)
		assert.Equal(t, "certfile", c.cert)
		assert.Equal(t, "somecname", c.cname)
		assert.Equal(t, uint(100), c.n)
		assert.Equal(t, uint(20), c.c)
		assert.Equal(t, uint(5), c.qps)
		assert.Equal(t, false, c.binary)
		assert.Equal(t, time.Duration(5*time.Minute), c.z)
		assert.Equal(t, time.Duration(60*time.Second), c.keepaliveTime)
		assert.Equal(t, time.Duration(10*time.Second), c.timeout)
		assert.Equal(t, time.Duration(30*time.Second), c.dialTimeout)
		assert.Equal(t, 4, c.cpus)
		assert.Equal(t, "asdf", c.name)
		assert.Equal(t, `{"name":"bob"}`, string(c.data))
		assert.Equal(t, `{"request-id":"123"}`, string(c.metadata))
		assert.Equal(t, "testdata/data.proto", string(c.proto))
		assert.Equal(t, "", string(c.protoset))
		assert.Equal(t, []string{"testdata", "."}, c.importPaths)
	})

	t.Run("with binary data and protoset", func(t *testing.T) {
		c, err := newConfig(
			WithCertificate("certfile", "somecname"),
			WithInsecure,
			WithN(100),
			WithC(20),
			WithQPS(5),
			WithZ(time.Duration(5*time.Minute)),
			WithKeepalive(time.Duration(60*time.Second)),
			WithTimeout(time.Duration(10*time.Second)),
			WithDialTimeout(time.Duration(30*time.Second)),
			WithName("asdf"),
			WithCPUs(4),
			WithBinaryData([]byte("asdf1234foobar")),
			WithMetadataFromJSON(`{"request-id":"123"}`),
			WithProtoset("testdata/bundle.protoset"),
		)

		assert.NoError(t, err)

		assert.Equal(t, true, c.insecure)
		assert.Equal(t, "certfile", c.cert)
		assert.Equal(t, "somecname", c.cname)
		assert.Equal(t, uint(100), c.n)
		assert.Equal(t, uint(20), c.c)
		assert.Equal(t, uint(5), c.qps)
		assert.Equal(t, true, c.binary)
		assert.Equal(t, time.Duration(5*time.Minute), c.z)
		assert.Equal(t, time.Duration(60*time.Second), c.keepaliveTime)
		assert.Equal(t, time.Duration(10*time.Second), c.timeout)
		assert.Equal(t, time.Duration(30*time.Second), c.dialTimeout)
		assert.Equal(t, 4, c.cpus)
		assert.Equal(t, "asdf", c.name)
		assert.Equal(t, []byte("asdf1234foobar"), c.data)
		assert.Equal(t, `{"request-id":"123"}`, string(c.metadata))
		assert.Equal(t, "", string(c.proto))
		assert.Equal(t, "testdata/bundle.protoset", string(c.protoset))
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

		c, err := newConfig(
			WithCertificate("certfile", "somecname"),
			WithInsecure,
			WithN(100),
			WithC(20),
			WithQPS(5),
			WithZ(time.Duration(5*time.Minute)),
			WithKeepalive(time.Duration(60*time.Second)),
			WithTimeout(time.Duration(10*time.Second)),
			WithDialTimeout(time.Duration(30*time.Second)),
			WithName("asdf"),
			WithCPUs(4),
			WithData(d),
			WithMetadata(&md),
		)

		assert.NoError(t, err)

		assert.Equal(t, true, c.insecure)
		assert.Equal(t, "certfile", c.cert)
		assert.Equal(t, "somecname", c.cname)
		assert.Equal(t, uint(100), c.n)
		assert.Equal(t, uint(20), c.c)
		assert.Equal(t, uint(5), c.qps)
		assert.Equal(t, false, c.binary)
		assert.Equal(t, time.Duration(5*time.Minute), c.z)
		assert.Equal(t, time.Duration(60*time.Second), c.keepaliveTime)
		assert.Equal(t, time.Duration(10*time.Second), c.timeout)
		assert.Equal(t, time.Duration(30*time.Second), c.dialTimeout)
		assert.Equal(t, 4, c.cpus)
		assert.Equal(t, "asdf", c.name)
		assert.Equal(t, `{"name":"bob","age":11,"fruits":["apple","peach","pear"]}`, string(c.data))
		assert.Equal(t, `{"request-id":"123","token":"foobar"}`, string(c.metadata))
	})
}
