# ghz
--
    import "."

Package ghz can be used to perform benchmarking and load testing against gRPC
services.

    report, err := ghz.Run(
    	"helloworld.Greeter.SayHello",
    	"localhost:50051",
    	WithProtoFile("greeter.proto", []string{}),
    	WithDataFromFile("data.json"),
    	WithInsecure(true),
    )

## Usage

```go
const (
	// ReasonNormalEnd indicates a normal end to the run
	ReasonNormalEnd = StopReason("normal")

	// ReasonCancel indicates end due to cancellation
	ReasonCancel = StopReason("cancel")

	// ReasonTimeout indicates run ended due to Z parameter timeout
	ReasonTimeout = StopReason("timeout")
)
```

#### type Bucket

```go
type Bucket struct {
	// The Mark for histogram bucket in seconds
	Mark float64 `json:"mark"`

	// The count in the bucket
	Count int `json:"count"`

	// The frequency of results in the bucket as a decimal percentage
	Frequency float64 `json:"frequency"`
}
```

Bucket holds histogram data

#### type LatencyDistribution

```go
type LatencyDistribution struct {
	Percentage int           `json:"percentage"`
	Latency    time.Duration `json:"latency"`
}
```

LatencyDistribution holds latency distribution data

#### type Option

```go
type Option func(*RunConfig) error
```

Option controls some aspect of run

#### func  WithBinaryData

```go
func WithBinaryData(data []byte) Option
```
WithBinaryData specifies the binary data

    msg := &helloworld.HelloRequest{}
    msg.Name = "bob"
    binData, _ := proto.Marshal(msg)
    WithBinaryData(binData)

#### func  WithBinaryDataFromFile

```go
func WithBinaryDataFromFile(path string) Option
```
WithBinaryDataFromFile specifies the binary data

    WithBinaryDataFromFile("request_data.bin")

#### func  WithCPUs

```go
func WithCPUs(c uint) Option
```
WithCPUs specifies the number of CPU's to be used

    WithCPUs(4)

#### func  WithCertificate

```go
func WithCertificate(cert string, cname string) Option
```
WithCertificate specifies the certificate options for the run

    WithCertificate("certfile.crt", "")

#### func  WithConcurrency

```go
func WithConcurrency(c uint) Option
```
WithConcurrency specifies the C (number of concurrent requests) option

    WithConcurrency(20)

#### func  WithData

```go
func WithData(data interface{}) Option
```
WithData specifies data as generic data that can be serailized to JSON

#### func  WithDataFromFile

```go
func WithDataFromFile(path string) Option
```
WithDataFromFile loads JSON data from file

    WithDataFromFile("data.json")

#### func  WithDataFromJSON

```go
func WithDataFromJSON(data string) Option
```
WithDataFromJSON loads JSON data from string

    WithDataFromJSON(`{"name":"bob"}`)

#### func  WithDataFromReader

```go
func WithDataFromReader(r io.Reader) Option
```
WithDataFromReader loads JSON data from reader

    file, _ := os.Open("data.json")
    WithDataFromReader(file)

#### func  WithDialTimeout

```go
func WithDialTimeout(dt time.Duration) Option
```
WithDialTimeout specifies the inital connection dial timeout

    WithDialTimeout(time.Duration(20*time.Second))

#### func  WithInsecure

```go
func WithInsecure(insec bool) Option
```
WithInsecure specifies that this run should be done using insecure mode

    WithInsecure(true)

#### func  WithKeepalive

```go
func WithKeepalive(k time.Duration) Option
```
WithKeepalive specifies the keepalive timeout

    WithKeepalive(time.Duration(1*time.Minute))

#### func  WithMetadata

```go
func WithMetadata(md *map[string]string) Option
```
WithMetadata specifies the metadata to be used as a map

    md := make(map[string]string)
    md["token"] = "foobar"
    md["request-id"] = "123"
    WithMetadata(&md)

#### func  WithMetadataFromFile

```go
func WithMetadataFromFile(path string) Option
```
WithMetadataFromFile loads JSON metadata from file

    WithMetadataFromJSON("metadata.json")

#### func  WithMetadataFromJSON

```go
func WithMetadataFromJSON(md string) Option
```
WithMetadataFromJSON specifies the metadata to be read from JSON string

    WithMetadataFromJSON(`{"request-id":"123"}`)

#### func  WithName

```go
func WithName(name string) Option
```
WithName sets the name of the test run

    WithName("greeter service test")

#### func  WithProtoFile

```go
func WithProtoFile(proto string, importPaths []string) Option
```
WithProtoFile specified proto file path and optionally import paths We will
automatically add the proto file path's directory and the current directory

    WithProtoFile("greeter.proto", []string{"/home/protos"})

#### func  WithProtoset

```go
func WithProtoset(protoset string) Option
```
WithProtoset specified protoset file path

    WithProtoset("bundle.protoset")

#### func  WithQPS

```go
func WithQPS(qps uint) Option
```
WithQPS specifies the QPS (queries per second) limit option

    WithQPS(10)

#### func  WithRunDuration

```go
func WithRunDuration(z time.Duration) Option
```
WithRunDuration specifies the Z (total test duration) option

    WithRunDuration(time.Duration(2*time.Minute))

#### func  WithTimeout

```go
func WithTimeout(timeout time.Duration) Option
```
WithTimeout specifies the timeout for each request

    WithTimeout(time.Duration(20*time.Second))

#### func  WithTotalRequests

```go
func WithTotalRequests(n uint) Option
```
WithTotalRequests specifies the N (number of total requests) setting

    WithTotalRequests(1000)

#### type Options

```go
type Options struct {
	Call          string             `json:"call,omitempty"`
	Proto         string             `json:"proto,omitempty"`
	Protoset      string             `json:"protoset,omitempty"`
	Host          string             `json:"host,omitempty"`
	Cert          string             `json:"cert,omitempty"`
	CName         string             `json:"cname,omitempty"`
	N             uint               `json:"n,omitempty"`
	C             uint               `json:"c,omitempty"`
	QPS           uint               `json:"qps,omitempty"`
	Z             time.Duration      `json:"z,omitempty"`
	Timeout       time.Duration      `json:"timeout,omitempty"`
	DialTimeout   time.Duration      `json:"dialTimeout,omitempty"`
	KeepaliveTime time.Duration      `json:"keepAlice,omitempty"`
	Data          interface{}        `json:"data,omitempty"`
	Binary        bool               `json:"binary"`
	Metadata      *map[string]string `json:"metadata,omitempty"`
	Insecure      bool               `json:"insecure"`
	CPUs          int                `json:"CPUs"`
	Name          string             `json:"name,omitempty"`
}
```

Options represents the request options

#### type Report

```go
type Report struct {
	Name      string     `json:"name,omitempty"`
	EndReason StopReason `json:"endReason,omitempty"`

	Options *Options  `json:"options,omitempty"`
	Date    time.Time `json:"date"`

	Count   uint64        `json:"count"`
	Total   time.Duration `json:"total"`
	Average time.Duration `json:"average"`
	Fastest time.Duration `json:"fastest"`
	Slowest time.Duration `json:"slowest"`
	Rps     float64       `json:"rps"`

	ErrorDist      map[string]int `json:"errorDistribution"`
	StatusCodeDist map[string]int `json:"statusCodeDistribution"`

	LatencyDistribution []LatencyDistribution `json:"latencyDistribution"`
	Histogram           []Bucket              `json:"histogram"`
	Details             []ResultDetail        `json:"details"`
}
```

Report holds the data for the full test

#### func  Run

```go
func Run(call, host string, options ...Option) (*Report, error)
```
Run executes the test

    report, err := ghz.Run(
    	"helloworld.Greeter.SayHello",
    	"localhost:50051",
    	WithProtoFile("greeter.proto", []string{}),
    	WithDataFromFile("data.json"),
    	WithInsecure(true),
    )

#### func (Report) MarshalJSON

```go
func (r Report) MarshalJSON() ([]byte, error)
```
MarshalJSON is custom marshal for report to properly format the date

#### type Reporter

```go
type Reporter struct {
}
```

Reporter gethers all the results

#### func (*Reporter) Finalize

```go
func (r *Reporter) Finalize(stopReason StopReason, total time.Duration) *Report
```
Finalize all the gathered data into a final report

#### func (*Reporter) Run

```go
func (r *Reporter) Run()
```
Run runs the reporter

#### type Requester

```go
type Requester struct {
}
```

Requester is used for doing the requests

#### func (*Requester) Finish

```go
func (b *Requester) Finish() *Report
```
Finish finishes the test run

#### func (*Requester) Run

```go
func (b *Requester) Run() (*Report, error)
```
Run makes all the requests and returns a report of results It blocks until all
work is done.

#### func (*Requester) Stop

```go
func (b *Requester) Stop(reason StopReason)
```
Stop stops the test

#### type ResultDetail

```go
type ResultDetail struct {
	Timestamp time.Time     `json:"timestamp"`
	Latency   time.Duration `json:"latency"`
	Error     string        `json:"error"`
	Status    string        `json:"status"`
}
```

ResultDetail data for each result

#### type RunConfig

```go
type RunConfig struct {
}
```

RunConfig represents the request Configs

#### type StopReason

```go
type StopReason string
```

StopReason is a reason why the run ended

#### func  ReasonFromString

```go
func ReasonFromString(str string) StopReason
```
ReasonFromString creates a Status from a string

#### func (StopReason) MarshalJSON

```go
func (s StopReason) MarshalJSON() ([]byte, error)
```
MarshalJSON formats a Threshold value into a JSON string

#### func (StopReason) String

```go
func (s StopReason) String() string
```
String() is the string representation of threshold

#### func (*StopReason) UnmarshalJSON

```go
func (s *StopReason) UnmarshalJSON(b []byte) error
```
UnmarshalJSON prases a Threshold value from JSON string
