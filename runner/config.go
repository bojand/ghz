package runner

import (
	"time"
)

// // Duration is our duration with TOML support
// type Duration time.Duration

// // UnmarshalText is our custom unmarshaller with TOML support
// func (d *Duration) UnmarshalText(text []byte) error {
// 	dur, err := time.ParseDuration(string(text))
// 	if err != nil {
// 		return err
// 	}

// 	*d = Duration(dur)
// 	return nil
// }

// // MarshalText implements encoding.TextMarshaler
// func (d Duration) MarshalText() ([]byte, error) {
// 	return []byte(time.Duration(d).String()), nil
// }

// func (d Duration) String() string {
// 	return time.Duration(d).String()
// }

// // UnmarshalJSON is our custom unmarshaller with JSON support
// func (d *Duration) UnmarshalJSON(text []byte) error {
// 	first := text[0]
// 	last := text[len(text)-1]
// 	if first == '"' && last == '"' {
// 		text = text[1 : len(text)-1]
// 	}
// 	dur, err := time.ParseDuration(string(text))
// 	if err != nil {
// 		return err
// 	}

// 	*d = Duration(dur)
// 	return nil
// }

// // MarshalJSON implements encoding JSONMarshaler
// func (d Duration) MarshalJSON() ([]byte, error) {
// 	return []byte(`"` + time.Duration(d).String() + `"`), nil
// }

// Config for the run.
type Config struct {
	Proto       string   `json:"proto" toml:"proto" yaml:"proto" mapstructure:"proto"`
	Protoset    string   `json:"protoset" toml:"protoset" yaml:"protoset" mapstructure:"protoset"`
	Call        string   `json:"call" toml:"call" yaml:"call" mapstructure:"call"`
	ImportPaths []string `json:"import-paths,omitempty" toml:"import-paths,omitempty" yaml:"import-paths,omitempty" mapstructure:"import-paths,omitempty"`

	RootCert      string `json:"cacert" toml:"cacert" yaml:"cacert" mapstructure:"cacert"`
	Cert          string `json:"cert" toml:"cert" yaml:"cert" mapstructure:"cert"`
	Key           string `json:"key" toml:"key" yaml:"key" mapstructure:"key"`
	CName         string `json:"cname" toml:"cname" yaml:"cname" mapstructure:"cname"`
	SkipTLSVerify bool   `json:"skip-verify" toml:"skip-verify" yaml:"skip-verify" mapstructure:"skip-verify"`
	Insecure      bool   `json:"insecure,omitempty" toml:"insecure,omitempty" yaml:"insecure,omitempty" mapstructure:"call,omitempty"`
	Authority     string `json:"authority" toml:"authority" yaml:"authority" mapstructure:"authority"`

	Async bool `json:"async,omitempty" toml:"async,omitempty" yaml:"async,omitempty" mapstructure:"async,omitempty"`
	RPS   uint `json:"rps" toml:"rps" yaml:"rps" mapstructure:"rps,omitempty"`

	LoadSchedule     string        `json:"load-schedule" toml:"load-schedule" yaml:"load-schedule" mapstructure:"load-schedule" default:"const"`
	LoadStart        uint          `json:"load-start" toml:"load-start" yaml:"load-start" mapstructure:"load-start"`
	LoadEnd          uint          `json:"load-end" toml:"load-end" yaml:"load-end" mapstructure:"load-end"`
	LoadStep         int           `json:"load-step" toml:"load-step" yaml:"load-step" mapstructure:"load-step"`
	LoadStepDuration time.Duration `json:"load-step-duration" toml:"load-step-duration" yaml:"load-step-duration" mapstructure:"load-duration"`
	LoadMaxDuration  time.Duration `json:"load-max-duration" toml:"load-max-duration" yaml:"load-max-duration" mapstructure:"load-max-duration"`

	C             uint          `json:"concurrency" toml:"concurrency" yaml:"concurrency" mapstructure:"call" efault:"50"`
	CSchedule     string        `json:"concurrency-schedule" toml:"concurrency-schedule" yaml:"concurrency-schedule" mapstructure:"concurrency-schedule" default:"const"`
	CStart        uint          `json:"concurrency-start" toml:"concurrency-start" yaml:"concurrency-start"  mapstructure:"concurrency-start" default:"1"`
	CEnd          uint          `json:"concurrency-end" toml:"concurrency-end" yaml:"concurrency-end" mapstructure:"concurrency-end" default:"0" `
	CStep         int           `json:"concurrency-step" toml:"concurrency-step" yaml:"concurrency-step" mapstructure:"concurrency-step" default:"0"`
	CStepDuration time.Duration `json:"concurrency-step-duration" toml:"concurrency-step-duration" yaml:"concurrency-step-duration" mapstructure:"concurrency-step-duration" default:"0"`
	CMaxDuration  time.Duration `json:"concurrency-max-duration" toml:"concurrency-max-duration" yaml:"concurrency-max-duration" mapstructure:"concurrency-max-duration" default:"0"`

	Total         uint          `json:"total" toml:"total" yaml:"total" mapstructure:"total" default:"200"`
	Timeout       time.Duration `json:"timeout" toml:"timeout" yaml:"timeout" mapstructure:"timeout" default:"20s"`
	TotalDuration time.Duration `json:"duration" toml:"duration" yaml:"duration" mapstructure:"duration"`
	MaxDuration   time.Duration `json:"max-duration" toml:"max-duration" yaml:"max-duration" mapstructure:"max-duration"`
	ZStop         string        `json:"duration-stop" toml:"duration-stop" yaml:"duration-stop" mapstructure:"duration-stop" default:"close"`

	Data                  interface{}       `json:"data,omitempty" toml:"data,omitempty" yaml:"data,omitempty" mapstructure:"data,omitempty"`
	DataPath              string            `json:"data-file" toml:"data-file" yaml:"data-file" mapstructure:"data-file"`
	BinData               []byte            `json:"-" toml:"-" yaml:"-" mapstructure:"-"`
	BinDataPath           string            `json:"binary-file" toml:"binary-file" yaml:"binary-file" mapstructure:"binary-file"`
	Metadata              map[string]string `json:"metadata,omitempty" toml:"metadata,omitempty" yaml:"metadata,omitempty" mapstructure:"metadata,omitempty"`
	MetadataPath          string            `json:"metadata-file" toml:"metadata-file" yaml:"metadata-file" mapstructure:"metadata-file"`
	SI                    time.Duration     `json:"stream-interval" toml:"stream-interval" yaml:"stream-interval" mapstructure:"stream-interfacl"`
	StreamCallDuration    time.Duration     `json:"stream-call-duration" toml:"stream-call-duration" yaml:"stream-call-duration" mapstructure:"stream-call-duration"`
	StreamCallCount       uint              `json:"stream-call-count" toml:"stream-call-count" yaml:"stream-call-count" mapstructure:"stream-call-count"`
	StreamDynamicMessages bool              `json:"stream-dynamic-messages" toml:"stream-dynamic-messages" yaml:"stream-dynamic-messages" mapstructure:"stream-dynamic-messages"`
	ReflectMetadata       map[string]string `json:"reflect-metadata,omitempty" toml:"reflect-metadata,omitempty" yaml:"reflect-metadata,omitempty" mapstructure:"reflect-metadata,omitempty"`

	Output      string `json:"output" toml:"output" yaml:"output" mapstructure:"output"`
	Format      string `json:"format" toml:"format" yaml:"format" mapstructure:"format" default:"summary"`
	SkipFirst   uint   `json:"skip-first" toml:"skip-first" yaml:"skip-first" mapstructure:"skip-first"`
	CountErrors bool   `json:"count-errors" toml:"count-errors" yaml:"count-errors" mapstructure:"count-errors"`

	Connections       uint          `json:"connections" toml:"connections" yaml:"connections" mapstructure:"connections" default:"1"`
	DialTimeout       time.Duration `json:"connect-timeout" toml:"connect-timeout" yaml:"connect-timeout" mapstructure:"connect-timeout" default:"10s"`
	KeepaliveTime     time.Duration `json:"keepalive" toml:"keepalive" yaml:"keepalive" mapstructure:"keepalive"`
	EnableCompression bool          `json:"enable-compression,omitempty" toml:"enable-compression,omitempty" yaml:"enable-compression,omitempty"  mapstructure:"enable-compression,omitempty"`
	LBStrategy        string        `json:"lb-strategy" toml:"lb-strategy" yaml:"lb-strategy" mapstructure:"lb-strategy"`

	Name  string            `json:"name,omitempty" toml:"name,omitempty" yaml:"name,omitempty" mapstructure:"name,omitempty"`
	Tags  map[string]string `json:"tags,omitempty" toml:"tags,omitempty" yaml:"tags,omitempty" mapstructure:"tags,omitempty"`
	CPUs  uint              `json:"cpus" toml:"cpus" yaml:"cpus" mapstructure:"cpus"`
	Debug string            `json:"debug,omitempty" toml:"debug,omitempty" yaml:"debug,omitempty" mapstructure:"debug,omitempty"`
	Host  string            `json:"host" toml:"host" yaml:"host" mapstructure:"host"`
}

// func checkData(data interface{}) error {
// 	_, isObjData := data.(map[string]interface{})
// 	if !isObjData {
// 		arrData, isArrData := data.([]interface{})
// 		if !isArrData {
// 			return errors.New("Unsupported type for Data")
// 		}
// 		if len(arrData) == 0 {
// 			return errors.New("Data array must not be empty")
// 		}
// 		for _, elem := range arrData {
// 			_, isObjData = elem.(map[string]interface{})
// 			if !isObjData {
// 				return errors.New("Data array contains unsupported type")
// 			}
// 		}

// 	}

// 	return nil
// }

// // LoadConfig loads the config from a file
// func LoadConfig(p string, c *Config) error {
// 	err := configor.Load(c, p)
// 	if err != nil {
// 		return err
// 	}

// 	if c.Data != nil {
// 		ext := path.Ext(p)
// 		if strings.EqualFold(ext, ".yaml") || strings.EqualFold(ext, ".yml") {
// 			objData, isObjData2 := c.Data.(map[interface{}]interface{})
// 			if isObjData2 {
// 				nd := make(map[string]interface{})
// 				for k, v := range objData {
// 					sk, isString := k.(string)
// 					if !isString {
// 						return errors.New("Data key must string")
// 					}
// 					if len(sk) > 0 {
// 						nd[sk] = v
// 					}
// 				}

// 				c.Data = nd
// 			}
// 		}

// 		err := checkData(c.Data)
// 		if err != nil {
// 			return err
// 		}
// 	}

// 	c.ZStop = strings.ToLower(c.ZStop)
// 	if c.ZStop != "close" && c.ZStop != "ignore" && c.ZStop != "wait" {
// 		c.ZStop = "close"
// 	}

// 	return nil
// }
