package runner

import (
	"errors"
	"path"
	"strings"
	"time"

	"github.com/jinzhu/configor"
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

// UnmarshalJSON is our custom unmarshaller with JSON support
func (d *Duration) UnmarshalJSON(text []byte) error {
	// strValue := string(text)
	first := text[0]
	last := text[len(text)-1]
	if first == '"' && last == '"' {
		text = text[1 : len(text)-1]
	}
	dur, err := time.ParseDuration(string(text))
	if err != nil {
		return err
	}

	*d = Duration(dur)
	return nil
}

// MarshalJSON implements encoding JSONMarshaler
func (d Duration) MarshalJSON() ([]byte, error) {
	return []byte(time.Duration(d).String()), nil
}

// Config for the run.
type Config struct {
	Proto             string            `json:"proto" toml:"proto" yaml:"proto"`
	Protoset          string            `json:"protoset" toml:"protoset" yaml:"protoset"`
	Call              string            `json:"call" toml:"call" yaml:"call"`
	RootCert          string            `json:"cacert" toml:"cacert" yaml:"cacert"`
	Cert              string            `json:"cert" toml:"cert" yaml:"cert"`
	Key               string            `json:"key" toml:"key" yaml:"key"`
	SkipTLSVerify     bool              `json:"skipTLS" toml:"skipTLS" yaml:"skipTLS"`
	CName             string            `json:"cname" toml:"cname" yaml:"cname"`
	Authority         string            `json:"authority" toml:"authority" yaml:"authority"`
	Insecure          bool              `json:"insecure,omitempty" toml:"insecure,omitempty" yaml:"insecure,omitempty"`
	N                 uint              `json:"total" toml:"total" yaml:"total" default:"200"`
	C                 uint              `json:"concurrency" toml:"concurrency" yaml:"concurrency" default:"50"`
	Connections       uint              `json:"connections" toml:"connections" yaml:"connections" default:"1"`
	QPS               uint              `json:"qps" toml:"qps" yaml:"qps"`
	Z                 Duration          `json:"duration" toml:"duration" yaml:"duration"`
	ZStop             string            `json:"duration-stop" toml:"duration-stop" yaml:"duration-stop" default:"close"`
	X                 Duration          `json:"max-duration" toml:"max-duration" yaml:"max-duration"`
	Timeout           Duration          `json:"timeout" toml:"timeout" yaml:"timeout" default:"20s"`
	Data              interface{}       `json:"data,omitempty" toml:"data,omitempty" yaml:"data,omitempty"`
	DataPath          string            `json:"data-file" toml:"data-file" yaml:"data-file"`
	BinData           []byte            `json:"-" toml:"-" yaml:"-"`
	BinDataPath       string            `json:"binary-file" toml:"binary-file" yaml:"binary-file"`
	Metadata          map[string]string `json:"metadata,omitempty" toml:"metadata,omitempty" yaml:"metadata,omitempty"`
	MetadataPath      string            `json:"metadata-file" toml:"metadata-file" yaml:"metadata-file"`
	SI                Duration          `json:"stream-interval" toml:"stream-interval" yaml:"stream-interval"`
	Output            string            `json:"output" toml:"output" yaml:"output"`
	Format            string            `json:"format" toml:"format" yaml:"format" default:"summary"`
	DialTimeout       Duration          `json:"connect-timeout" toml:"connect-timeout" yaml:"connect-timeout" default:"10s"`
	KeepaliveTime     Duration          `json:"keepalive" toml:"keepalive" yaml:"keepalive"`
	CPUs              uint              `json:"cpus" toml:"cpus" yaml:"cpus"`
	ImportPaths       []string          `json:"import-paths,omitempty" toml:"import-paths,omitempty" yaml:"import-paths,omitempty"`
	Name              string            `json:"name,omitempty" toml:"name,omitempty" yaml:"name,omitempty"`
	Tags              map[string]string `json:"tags,omitempty" toml:"tags,omitempty" yaml:"tags,omitempty"`
	ReflectMetadata   map[string]string `json:"reflect-metadata,omitempty" toml:"reflect-metadata,omitempty" yaml:"reflect-metadata,omitempty"`
	Debug             string            `json:"debug,omitempty" toml:"debug,omitempty" yaml:"debug,omitempty"`
	Host              string            `json:"host" toml:"host" yaml:"host"`
	EnableCompression bool              `json:"enable-compression,omitempty" toml:"enable-compression,omitempty" yaml:"enable-compression,omitempty"`
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

// LoadConfig loads the config from a file
func LoadConfig(p string, c *Config) error {
	err := configor.Load(c, p)
	if err != nil {
		return err
	}

	if c.Data != nil {
		ext := path.Ext(p)
		if strings.EqualFold(ext, ".yaml") || strings.EqualFold(ext, ".yml") {
			objData, isObjData2 := c.Data.(map[interface{}]interface{})
			if isObjData2 {
				nd := make(map[string]interface{})
				for k, v := range objData {
					sk, isString := k.(string)
					if !isString {
						return errors.New("Data key must string")
					}
					if len(sk) > 0 {
						nd[sk] = v
					}
				}

				c.Data = nd
			}
		}

		err := checkData(c.Data)
		if err != nil {
			return err
		}
	}

	c.ZStop = strings.ToLower(c.ZStop)
	if c.ZStop != "close" && c.ZStop != "ignore" && c.ZStop != "wait" {
		c.ZStop = "close"
	}

	return nil
}
