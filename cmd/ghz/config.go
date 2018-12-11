package main

import (
	"encoding/json"
	"errors"
	"time"
)

// Config for the run.
type config struct {
	Proto         string             `json:"proto"`
	Protoset      string             `json:"protoset"`
	Call          string             `json:"call" required:"true"`
	Cert          string             `json:"cert"`
	CName         string             `json:"cName"`
	N             uint               `json:"n" default:"200"`
	C             uint               `json:"c" default:"50"`
	QPS           uint               `json:"q"`
	Z             time.Duration      `json:"z"`
	X             time.Duration      `json:"x"`
	Timeout       uint               `json:"t" default:"20"`
	Data          interface{}        `json:"d,omitempty"`
	DataPath      string             `json:"D"`
	BinData       []byte             `json:"-"`
	BinDataPath   string             `json:"B"`
	Metadata      *map[string]string `json:"m,omitempty"`
	MetadataPath  string             `json:"M"`
	Output        string             `json:"o"`
	Format        string             `json:"O"`
	Host          string             `json:"host"`
	DialTimeout   uint               `json:"T" default:"10"`
	KeepaliveTime uint               `json:"L"`
	CPUs          uint               `json:"cpus"`
	ImportPaths   []string           `json:"i,omitempty"`
	Insecure      bool               `json:"insecure,omitempty"`
	Name          string             `json:"name,omitempty"`
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

	c.Z, _ = time.ParseDuration(aux.Z)
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
