package main

import (
	"encoding/json"
	"errors"
	"strings"
	"time"
)

// Duration is our duration with TOML support
type Duration time.Duration

// UnmarshalText is our custom unmarshaller with TOML support
func (d *Duration) UnmarshalText(text []byte) error {
	dur, err := time.ParseDuration(string(text))
	if err != nil {
		return err
	}

	*d = Duration(dur)
	return nil
}

// MarshalText implements encoding.TextMarshaler
func (d Duration) MarshalText() ([]byte, error) {
	return []byte(time.Duration(d).String()), nil
}

func (d Duration) String() string {
	return time.Duration(d).String()
}

// config for the run.
type config struct {
	Proto             string             `json:"proto" toml:"proto" yaml:"proto"`
	Protoset          string             `json:"protoset" toml:"protoset" yaml:"protoset"`
	Call              string             `json:"call" toml:"call" yaml:"call" required:"true"`
	RootCert          string             `json:"cacert" toml:"cacert" yaml:"cacert"`
	Cert              string             `json:"cert" toml:"cert" yaml:"cert"`
	Key               string             `json:"key" toml:"key" yaml:"key"`
	SkipTLSVerify     bool               `json:"skipTLS" toml:"skipTLS" yaml:"skipTLS"`
	CName             string             `json:"cname" toml:"cname" yaml:"cname"`
	Authority         string             `json:"authority" toml:"authority" yaml:"authority"`
	Insecure          bool               `json:"insecure,omitempty" toml:"insecure,omitempty" yaml:"insecure,omitempty"`
	N                 uint               `json:"total" toml:"total" yaml:"total" default:"200"`
	C                 uint               `json:"concurrency" toml:"concurrency" yaml:"concurrency" default:"50"`
	Connections       uint               `json:"connections" toml:"connections" yaml:"connections" default:"1"`
	QPS               uint               `json:"qps" toml:"qps" yaml:"qps"`
	Z                 Duration           `json:"duration" toml:"duration" yaml:"duration"`
	ZStop             string             `json:"duration-stop" toml:"duration-stop" yaml:"duration-stop" default:"close"`
	X                 Duration           `json:"max-duration" toml:"max-duration" yaml:"max-duration"`
	Timeout           Duration           `json:"timeout" toml:"timeout" yaml:"timeout" default:"20s"`
	Data              interface{}        `json:"data,omitempty" toml:"data,omitempty" yaml:"data,omitempty"`
	DataPath          string             `json:"data-file" toml:"data-file" yaml:"data-file"`
	BinData           []byte             `json:"-" toml:"-" yaml:"-"`
	BinDataPath       string             `json:"binary-file" toml:"binary-file" yaml:"binary-file"`
	Metadata          *map[string]string `json:"metadata,omitempty" toml:"metadata,omitempty" yaml:"metadata,omitempty"`
	MetadataPath      string             `json:"metadata-file" toml:"metadata-file" yaml:"metadata-file"`
	SI                Duration           `json:"stream-interval" toml:"stream-interval" yaml:"stream-interval"`
	Output            string             `json:"output" toml:"output" yaml:"output"`
	Format            string             `json:"format" toml:"format" yaml:"format" default:"summary"`
	DialTimeout       Duration           `json:"connect-timeout" toml:"connect-timeout" yaml:"connect-timeout" default:"10s"`
	KeepaliveTime     Duration           `json:"keepalive" toml:"keepalive" yaml:"keepalive"`
	CPUs              uint               `json:"cpus" toml:"cpus" yaml:"cpus"`
	ImportPaths       []string           `json:"import-paths,omitempty" toml:"import-paths,omitempty" yaml:"import-paths,omitempty"`
	Name              string             `json:"name,omitempty" toml:"name,omitempty" yaml:"name,omitempty"`
	Tags              *map[string]string `json:"tags,omitempty" toml:"tags,omitempty" yaml:"tags,omitempty"`
	ReflectMetadata   *map[string]string `json:"reflect-metadata,omitempty" toml:"reflect-metadata,omitempty" yaml:"reflect-metadata,omitempty"`
	Debug             string             `json:"debug,omitempty" toml:"debug,omitempty" yaml:"debug,omitempty"`
	Host              string             `json:"host" toml:"host" yaml:"host"`
	EnableCompression bool               `json:"enable-compression,omitempty" toml:"enable-compression,omitempty" yaml:"enable-compression,omitempty"`
}

// UnmarshalJSON is our custom implementation to handle the Duration fields
// and validate data
func (c *config) UnmarshalJSON(data []byte) error {
	type Alias config
	aux := &struct {
		Z       string `json:"duration"`
		X       string `json:"max-duration"`
		SI      string `json:"stream-interval"`
		Timeout string `json:"timeout"`
		*Alias
	}{
		Alias: (*Alias)(c),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	if aux.Data != nil {
		err := checkData(aux.Data)
		if err != nil {
			return err
		}
	}

	if aux.Z != "" {
		zd, err := time.ParseDuration(aux.Z)
		if err != nil {
			return nil
		}

		c.Z = Duration(zd)
	}

	aux.ZStop = strings.ToLower(aux.ZStop)
	if aux.ZStop != "close" && aux.ZStop != "ignore" && aux.ZStop != "wait" {
		aux.ZStop = "close"
	}

	if aux.X != "" {
		xd, err := time.ParseDuration(aux.X)
		if err != nil {
			return nil
		}

		c.X = Duration(xd)
	}

	if aux.SI != "" {
		sid, err := time.ParseDuration(aux.SI)
		if err != nil {
			return nil
		}

		c.SI = Duration(sid)
	}

	if aux.Timeout != "" {
		td, err := time.ParseDuration(aux.Timeout)
		if err != nil {
			return nil
		}

		c.Timeout = Duration(td)
	}

	return nil
}

// MarshalJSON is our custom implementation to handle the Duration fields
func (c config) MarshalJSON() ([]byte, error) {
	type Alias config
	return json.Marshal(&struct {
		*Alias
		Z       string `json:"duration"`
		X       string `json:"max-duration"`
		SI      string `json:"stream-interval"`
		Timeout string `json:"timeout"`
	}{
		Alias:   (*Alias)(&c),
		Z:       c.Z.String(),
		X:       c.X.String(),
		SI:      c.SI.String(),
		Timeout: c.Timeout.String(),
	})
}

func checkData(data interface{}) error {
	_, isObjData := data.(map[string]interface{})
	if !isObjData {
		arrData, isArrData := data.([]interface{})
		if !isArrData {
			return errors.New("Unsupported type for Data")
		}
		if len(arrData) == 0 {
			return errors.New("Data array must not be empty")
		}
		for _, elem := range arrData {
			_, isObjData = elem.(map[string]interface{})
			if !isObjData {
				return errors.New("Data array contains unsupported type")
			}
		}

	}

	return nil
}
