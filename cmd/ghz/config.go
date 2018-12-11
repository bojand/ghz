package main

import (
	"encoding/json"
	"errors"
	"time"
)

type duration struct {
	time.Duration
}

func (d *duration) UnmarshalText(text []byte) error {
	var err error
	d.Duration, err = time.ParseDuration(string(text))
	return err
}

// Config for the run.
type config struct {
	Proto         string             `json:"proto" toml:"proto" yaml:"proto"`
	Protoset      string             `json:"protoset" toml:"protoset" yaml:"protoset"`
	Call          string             `json:"call" toml:"call" yaml:"call" required:"true"`
	Cert          string             `json:"cert" toml:"cert" yaml:"cert"`
	CName         string             `json:"cname" toml:"cname" yaml:"cname"`
	N             uint               `json:"n" toml:"n" yaml:"n" default:"200"`
	C             uint               `json:"c" toml:"c" yaml:"c" default:"50"`
	QPS           uint               `json:"q" toml:"q" yaml:"q"`
	Z             duration           `json:"z" toml:"z" yaml:"z"`
	X             duration           `json:"x" toml:"x" yaml:"x"`
	Timeout       uint               `json:"t" toml:"t" yaml:"t" default:"20"`
	Data          interface{}        `json:"d,omitempty" toml:"d,omitempty" yaml:"d,omitempty"`
	DataPath      string             `json:"D" toml:"D" yaml:"D"`
	BinData       []byte             `json:"-" toml:"-" yaml:"-"`
	BinDataPath   string             `json:"B" toml:"B" yaml:"B"`
	Metadata      *map[string]string `json:"m,omitempty" toml:"m,omitempty" yaml:"m,omitempty"`
	MetadataPath  string             `json:"M" toml:"M" yaml:"M"`
	Output        string             `json:"o" toml:"o" yaml:"o"`
	Format        string             `json:"O" toml:"O" yaml:"O"`
	Host          string             `json:"host" toml:"host" yaml:"host"`
	DialTimeout   uint               `json:"T" toml:"T" yaml:"T" default:"10"`
	KeepaliveTime uint               `json:"L" toml:"L" yaml:"L"`
	CPUs          uint               `json:"cpus" toml:"cpus" yaml:"cpus"`
	ImportPaths   []string           `json:"i,omitempty" toml:"i,omitempty" yaml:"i,omitempty"`
	Insecure      bool               `json:"insecure,omitempty" toml:"insecure,omitempty" yaml:"insecure,omitempty"`
	Name          string             `json:"name,omitempty" toml:"name,omitempty" yaml:"name,omitempty"`
}

// UnmarshalJSON is our custom implementation to handle the Duration field Z
// and validate data
func (c *config) UnmarshalJSON(data []byte) error {
	type Alias config
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

	zd, _ := time.ParseDuration(aux.Z)
	c.Z = duration{zd}
	return nil
}

// MarshalJSON is our custom implementation to handle the Duration field Z
func (c config) MarshalJSON() ([]byte, error) {
	type Alias config
	return json.Marshal(&struct {
		*Alias
		Z string `json:"z"`
		X string `json:"x"`
	}{
		Alias: (*Alias)(&c),
		Z:     c.Z.String(),
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
