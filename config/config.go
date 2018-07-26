package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// Config for the run.
type Config struct {
	Proto         string             `json:"proto"`
	Protoset      string             `json:"protoset"`
	Call          string             `json:"call"`
	Cert          string             `json:"cert"`
	CName         string             `json:"cName"`
	N             int                `json:"n"`
	C             int                `json:"c"`
	QPS           int                `json:"q"`
	Z             time.Duration      `json:"z"`
	X             time.Duration      `json:"x"`
	Timeout       int                `json:"t"`
	Data          interface{}        `json:"d,omitempty"`
	DataPath      string             `json:"D"`
	Metadata      *map[string]string `json:"m,omitempty"`
	MetadataPath  string             `json:"M"`
	Output        string             `json:"o"`
	Format        string             `json:"O"`
	Host          string             `json:"host"`
	DialTimeout   int                `json:"T"`
	KeepaliveTime int                `json:"L"`
	CPUs          int                `json:"cpus"`
	ImportPaths   []string           `json:"i,omitempty"`
}

// New creates a new config
func New(proto, protoset, call, cert, cName string, n, c, qps int, z time.Duration, x time.Duration,
	timeout int, data, dataPath, metadata, mdPath, output, format, host string,
	dialTimout, keepaliveTime, cpus int, importPaths []string) (*Config, error) {

	cfg := &Config{
		Proto:         proto,
		Protoset:      protoset,
		Call:          call,
		Cert:          cert,
		CName:         cName,
		N:             n,
		C:             c,
		QPS:           qps,
		Z:             z,
		X:             x,
		Timeout:       timeout,
		DataPath:      dataPath,
		MetadataPath:  mdPath,
		Output:        output,
		Format:        format,
		Host:          host,
		ImportPaths:   importPaths,
		DialTimeout:   dialTimout,
		KeepaliveTime: keepaliveTime,
		CPUs:          cpus}

	if data == "@" {
		b, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			return nil, err
		}
		data = string(b)
	}

	err := cfg.setData(data)
	if err != nil {
		return nil, err
	}

	err = cfg.setMetadata(metadata)
	if err != nil {
		return nil, err
	}

	err = cfg.init()
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

// Default sets the defaults values for some of the properties
// that need to have valid values
func (c *Config) Default() {
	if c.N == 0 {
		c.N = 200
	}

	if c.C == 0 {
		c.C = 50
	}

	if c.CPUs == 0 {
		c.CPUs = runtime.GOMAXPROCS(-1)
	}

	if c.Timeout == 0 {
		c.Timeout = 20
	}

	if c.DialTimeout == 0 {
		c.DialTimeout = 10
	}

	c.ImportPaths = append(c.ImportPaths, ".")
	dir := filepath.Dir(c.Proto)
	if dir != "." {
		c.ImportPaths = append(c.ImportPaths, dir)
	}
}

// Validate the config
func (c *Config) Validate() error {
	if strings.TrimSpace(c.Proto) == "" && strings.TrimSpace(c.Protoset) == "" {
		return errors.New("Proto or Protoset required")
	}

	if strings.TrimSpace(c.Proto) != "" {
		if filepath.Ext(c.Proto) != ".proto" {
			return errors.Errorf(fmt.Sprintf("proto: must have .proto extension"))
		}
	} else {
		if filepath.Ext(c.Protoset) != ".protoset" {
			return errors.Errorf(fmt.Sprintf("protoset: must have .protoset extension"))
		}
	}

	if err := requiredString(c.Call); err != nil {
		return errors.Wrap(err, "call")
	}

	if err := minValue(c.N, 0); err != nil {
		return errors.Wrap(err, "n")
	}

	if err := minValue(c.C, 0); err != nil {
		return errors.Wrap(err, "c")
	}

	if err := minValue(c.QPS, 0); err != nil {
		return errors.Wrap(err, "q")
	}

	if err := minValue(c.Timeout, 0); err != nil {
		return errors.Wrap(err, "t")
	}

	if err := minValue(c.DialTimeout, 0); err != nil {
		return errors.Wrap(err, "connectionTimeout")
	}

	if err := minValue(c.KeepaliveTime, 0); err != nil {
		return errors.Wrap(err, "keepaliveTime")
	}

	if err := minValue(c.CPUs, 0); err != nil {
		return errors.Wrap(err, "cpus")
	}

	if strings.TrimSpace(c.DataPath) == "" {
		if c.Data == nil {
			return errors.New("data: is required")
		}
	}

	return nil
}

// UnmarshalJSON is our custom implementation to handle the Duration field Z
// and validate data
func (c *Config) UnmarshalJSON(data []byte) error {
	type Alias Config
	aux := &struct {
		Z string `json:"z"`
		X string `json:"x"`
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

	c.Z, _ = time.ParseDuration(aux.Z)
	return nil
}

// MarshalJSON is our custom implementation to handle the Duration field Z
func (c Config) MarshalJSON() ([]byte, error) {
	type Alias Config
	return json.Marshal(&struct {
		*Alias
		Z string `json:"z"`
		X string `json:"x"`
	}{
		Alias: (*Alias)(&c),
		Z:     c.Z.String(),
	})
}

// ReadConfig reads the JSON config from path, applies the defaults and validates
func ReadConfig(path string) (*Config, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return parseConfig(b)
}

// InitData returns the payload data
func (c *Config) initData() error {
	if c.Data != nil {
		return nil
	} else if strings.TrimSpace(c.DataPath) != "" {
		d, err := ioutil.ReadFile(c.DataPath)
		if err != nil {
			return err
		}

		return json.Unmarshal(d, &c.Data)
	}

	return errors.New("No data specified")
}

// SetData sets data based on input JSON string
func (c *Config) setData(in string) error {
	if strings.TrimSpace(in) != "" {
		return json.Unmarshal([]byte(in), &c.Data)
	}
	return nil
}

// SetMetadata sets the metadata based on input JSON string
func (c *Config) setMetadata(in string) error {
	if strings.TrimSpace(in) != "" {
		return json.Unmarshal([]byte(in), &c.Metadata)
	}
	return nil
}

// InitMetadata returns the payload data
func (c *Config) initMetadata() error {
	if c.Metadata != nil && len(*c.Metadata) > 0 {
		return nil
	} else if strings.TrimSpace(c.MetadataPath) != "" {
		d, err := ioutil.ReadFile(c.MetadataPath)
		if err != nil {
			return err
		}

		return json.Unmarshal(d, &c.Metadata)
	}

	return nil
}

func (c *Config) initDurations() {
	if c.X > 0 {
		c.Z = c.X
	} else if c.Z > 0 {
		c.N = math.MaxInt32
	}
}

func (c *Config) init() error {
	err := c.initData()
	if err != nil {
		return err
	}

	err = c.initMetadata()
	if err != nil {
		return err
	}

	c.initDurations()

	c.Default()

	err = c.Validate()
	if err != nil {
		return err
	}

	return nil
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

// RequiredString checks if the required string is empty
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

func parseConfigString(s string) (*Config, error) {
	return parseConfig([]byte(s))
}

func parseConfig(b []byte) (*Config, error) {
	c := &Config{}

	if err := json.Unmarshal(b, c); err != nil {
		return nil, errors.Wrap(err, "parsing json")
	}

	err := c.init()
	if err != nil {
		return nil, err
	}

	return c, nil
}
