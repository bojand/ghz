package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// Config for the run.
type Config struct {
	Proto    string        `json:"proto"`
	Call     string        `json:"call"`
	CACert   string        `json:"cacert"`
	Cert     string        `json:"cert"`
	Key      string        `json:"key"`
	Insecure bool          `json:"insecure"`
	N        int           `json:"n"`
	C        int           `json:"c"`
	QPS      int           `json:"q"`
	Z        time.Duration `json:"z"`
	Timeout  int           `json:"t"`
	Data     string        `json:"d"`
	DataPath string        `json:"D"`
	Metadata string        `json:"m"`
	MDPath   string        `json:"M"`
	Format   string        `json:"o"`
	Output   string        `json:"O"`
	CPUs     int           `json:"cpus"`
}

// ReadConfig reads the configuration from `path`.
func ReadConfig(path string) (*Config, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return ParseConfig(b)
}

// ParseConfigString returns config from JSON string.
func ParseConfigString(s string) (*Config, error) {
	return ParseConfig([]byte(s))
}

// ParseConfig returns config from JSON bytes.
func ParseConfig(b []byte) (*Config, error) {
	c := &Config{}

	if err := json.Unmarshal(b, c); err != nil {
		return nil, errors.Wrap(err, "parsing json")
	}

	c.Default()

	if err := c.Validate(); err != nil {
		return nil, errors.Wrap(err, "validating")
	}

	return c, nil
}

// Default implementation.
func (c *Config) Default() {
	if c.N == 0 {
		c.N = 200
	}

	if c.C == 0 {
		c.C = 50
	}
}

// Validate implementation.
func (c *Config) Validate() error {
	if err := requiredString(c.Proto); err != nil {
		return errors.Wrap(err, ".proto")
	}

	if err := requiredString(c.Call); err != nil {
		return errors.Wrap(err, ".call")
	}

	if c.Insecure == false {
		if strings.TrimSpace(c.Cert) != "" {
			if err := requiredString(c.Key); err != nil {
				return errors.Wrap(err, ".key")
			}
		} else if err := requiredString(c.CACert); err != nil {
			return errors.Wrap(err, ".cacert")
		}
	}

	if err := minValue(c.N, 0); err != nil {
		return errors.Wrap(err, ".n")
	}

	if err := minValue(c.C, 0); err != nil {
		return errors.Wrap(err, ".c")
	}

	if err := minValue(c.QPS, 0); err != nil {
		return errors.Wrap(err, ".q")
	}

	if err := minValue(c.Timeout, 0); err != nil {
		return errors.Wrap(err, ".t")
	}

	if err := minValue(c.CPUs, 0); err != nil {
		return errors.Wrap(err, ".cpus")
	}

	if strings.TrimSpace(c.DataPath) != "" {
		if err := requiredString(c.Data); err != nil {
			return errors.Wrap(err, ".data")
		}
	}

	return nil
}

// UnmarshalJSON is our custom implementation to handle the Duration field Z
func (c *Config) UnmarshalJSON(data []byte) error {
	type Alias Config
	aux := &struct {
		Z string `json:"z"`
		*Alias
	}{
		Alias: (*Alias)(c),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	c.Z, _ = time.ParseDuration(aux.Z)
	return nil
}

// MarshalJSON is our custom implementation to handle the Duration field Z
func (c *Config) MarshalJSON() ([]byte, error) {
	type Alias Config
	aux := &struct {
		Z string `json:"z"`
		*Alias
	}{
		Alias: (*Alias)(c),
	}

	aux.Z = c.Z.String()

	return json.Marshal(&aux)
}

func requiredString(s string) error {
	if strings.TrimSpace(s) == "" {
		return errors.New("is required")
	}

	return nil
}

func minValue(v int, min int) error {
	if v < min {
		return errors.Errorf(fmt.Sprintf("must be at least %d", min))
	}

	return nil
}
