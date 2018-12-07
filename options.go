package ghz

import (
	"encoding/json"
	"runtime"
	"time"

	"github.com/bojand/hri"
)

// RunConfig represents the request Configs
type RunConfig struct {
	cert          string
	cname         string
	insecure      bool
	n             int
	c             int
	qps           int
	z             time.Duration
	timeout       time.Duration
	dialTimeout   time.Duration
	keepaliveTime time.Duration
	data          []byte
	binary        bool
	metadata      []byte
	name          string
	cpus          int
}

// Option controls some aspect of run
type Option func(*RunConfig) error

// WithCertificate specifies the certificate options for the run
func WithCertificate(cert string, cname string) Option {
	return func(o *RunConfig) error {
		o.cert = cert
		o.cname = cname

		return nil
	}
}

// WithInsecure specifies that this run should be done using insecure mode
func WithInsecure(o *RunConfig) error {
	o.insecure = true
	return nil
}

// WithN specifies the N (number of total requests) Config
func WithN(n int) Option {
	return func(o *RunConfig) error {
		o.n = n

		return nil
	}
}

// WithC specifies the N (number of concurrent requests) Config
func WithC(c int) Option {
	return func(o *RunConfig) error {
		o.c = c

		return nil
	}
}

// WithQPS specifies the QPS (queries per second) limit Config
func WithQPS(qps int) Option {
	return func(o *RunConfig) error {
		o.qps = qps

		return nil
	}
}

// WithZ specifies the Z (total test duration) Config
func WithZ(z time.Duration) Option {
	return func(o *RunConfig) error {
		o.z = z

		return nil
	}
}

// WithTimeout specifies the timeout for each request
func WithTimeout(timeout time.Duration) Option {
	return func(o *RunConfig) error {
		o.timeout = timeout

		return nil
	}
}

// WithDialTimeout specifies the inital connection dial timeout
func WithDialTimeout(dt time.Duration) Option {
	return func(o *RunConfig) error {
		o.dialTimeout = dt

		return nil
	}
}

// WithKeepalive specifies the keepalive timeout
func WithKeepalive(k time.Duration) Option {
	return func(o *RunConfig) error {
		o.keepaliveTime = k

		return nil
	}
}

// WithBinaryData specifies the binary data
func WithBinaryData(data []byte) Option {
	return func(o *RunConfig) error {
		o.data = data
		o.binary = true

		return nil
	}
}

// WithDataFromJSON loads JSON data from string
func WithDataFromJSON(data string) Option {
	return func(o *RunConfig) error {
		o.data = []byte(data)
		o.binary = false

		return nil
	}
}

// WithData specifies data as generic data that can be serailized to JSON
func WithData(data interface{}) Option {
	return func(o *RunConfig) error {
		dataJSON, err := json.Marshal(data)

		if err != nil {
			return err
		}

		o.data = dataJSON
		o.binary = false

		return nil
	}
}

// WithMetadataFromJSON specifies the metadata to be read from JSON string
func WithMetadataFromJSON(md string) Option {
	return func(o *RunConfig) error {
		o.metadata = []byte(md)

		return nil
	}
}

// WithMetadata specifies the metadata to be used as a map
func WithMetadata(md *map[string]string) Option {
	return func(o *RunConfig) error {
		mdJSON, err := json.Marshal(md)
		if err != nil {
			return err
		}

		o.metadata = mdJSON

		return nil
	}
}

// WithName sets the name of the test run
func WithName(name string) Option {
	return func(o *RunConfig) error {
		o.name = name

		return nil
	}
}

// WithCPUs specifies the number of CPU's to be used
func WithCPUs(c int) Option {
	return func(o *RunConfig) error {
		o.cpus = c

		return nil
	}
}

func newConfig(options ...Option) (*RunConfig, error) {
	c := &RunConfig{
		n:           200,
		c:           50,
		timeout:     time.Duration(20 * time.Second),
		dialTimeout: time.Duration(10 * time.Second),
		cpus:        runtime.GOMAXPROCS(-1),
		name:        hri.Random(),
	}

	for _, option := range options {
		err := option(c)

		if err != nil {
			return nil, err
		}
	}

	return c, nil
}
