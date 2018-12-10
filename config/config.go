package config

import "time"

// Config for the run.
type Config struct {
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
