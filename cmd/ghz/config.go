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
	Proto         string             `json:"proto" toml:"proto"`
	Protoset      string             `json:"protoset"`
	Call          string             `json:"call" required:"true"`
	Cert          string             `json:"cert"`
	CName         string             `json:"cName"`
	N             uint               `json:"n" toml:"n" default:"200"`
	C             uint               `json:"c" toml:"c" default:"50"`
	QPS           uint               `json:"q" toml:"q"`
	Z             duration           `json:"z" toml:"z"`
	X             duration           `json:"x" toml:"x"`
	Timeout       uint               `json:"t" toml:"t" default:"20"`
	Data          interface{}        `json:"d,omitempty" toml:"d,omitempty"`
	DataPath      string             `json:"D" toml:"D"`
	BinData       []byte             `json:"-" toml:"-"`
	BinDataPath   string             `json:"B" toml:"B"`
	Metadata      *map[string]string `json:"m,omitempty" toml:"m,omitempty"`
	MetadataPath  string             `json:"M" toml:"M"`
	Output        string             `json:"o" toml:"o"`
	Format        string             `json:"O" toml:"O"`
	Host          string             `json:"host" toml:"host"`
	DialTimeout   uint               `json:"T" toml:"T" default:"10"`
	KeepaliveTime uint               `json:"L" toml:"L"`
	CPUs          uint               `json:"cpus" toml:"cpus"`
	ImportPaths   []string           `json:"i,omitempty" toml:"i,omitempty"`
	Insecure      bool               `json:"insecure,omitempty" toml:"insecure,omitempty"`
	Name          string             `json:"name,omitempty" toml:"name,omitempty"`
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
