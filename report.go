package ghz

import (
	"encoding/json"
	"time"
)

// Options represents the request options
type Options struct {
	Call          string             `json:"call,omitempty"`
	Proto         string             `json:"proto,omitempty"`
	Host          string             `json:"host,omitempty"`
	Cert          string             `json:"cert,omitempty"`
	CName         string             `json:"cname,omitempty"`
	N             int                `json:"n,omitempty"`
	C             int                `json:"c,omitempty"`
	QPS           int                `json:"qps,omitempty"`
	Z             time.Duration      `json:"z,omitempty"`
	Timeout       time.Duration      `json:"timeout,omitempty"`
	DialTimeout   time.Duration      `json:"dialTimeout,omitempty"`
	KeepaliveTime time.Duration      `json:"keepAlice,omitempty"`
	Data          interface{}        `json:"data,omitempty"`
	Binary        bool               `json:"binary"`
	Metadata      *map[string]string `json:"metadata,omitempty"`
	Insecure      bool               `json:"insecure,omitempty"`
	CPUs          int                `json:"CPUs"`
	Name          string             `json:"name,omitempty"`
}

// Report holds the data for the full test
type Report struct {
	Name string `json:"name,omitempty"`

	Options *Options  `json:"options,omitempty"`
	Date    time.Time `json:"date"`
}

func newReport(call, proto, host string, c *RunConfig) *Report {

	r := &Report{
		Name: c.name,
	}

	r.Options = &Options{
		Call:          call,
		Proto:         proto,
		Host:          host,
		Cert:          c.cert,
		CName:         c.cname,
		N:             c.n,
		C:             c.c,
		QPS:           c.qps,
		Z:             c.z,
		Timeout:       c.timeout,
		DialTimeout:   c.dialTimeout,
		KeepaliveTime: c.keepaliveTime,
		Binary:        c.binary,
		Insecure:      c.insecure,
		CPUs:          c.cpus,
		Name:          c.name,
	}

	_ = json.Unmarshal(c.data, &r.Options.Data)

	_ = json.Unmarshal(c.metadata, &r.Options.Metadata)

	return r
}
