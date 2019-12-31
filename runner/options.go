package runner

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/pkg/errors"
	"google.golang.org/grpc/credentials"
)

// RunConfig represents the request Configs
type RunConfig struct {
	// call settings
	call        string
	host        string
	proto       string
	importPaths []string
	protoset    string

	// security settings
	creds      credentials.TransportCredentials
	cacert     string
	cert       string
	key        string
	cname      string
	skipVerify bool
	insecure   bool
	authority  string

	// test
	n   int
	c   int
	qps int

	// number of connections
	nConns int

	// timeouts
	z             time.Duration
	timeout       time.Duration
	dialTimeout   time.Duration
	keepaliveTime time.Duration

	zstop string

	streamInterval time.Duration

	// data
	data     []byte
	binary   bool
	metadata []byte
	rmd      *map[string]string

	// debug
	hasLog bool
	log    Logger

	// misc
	name string
	cpus int
	tags []byte
}

// Option controls some aspect of run
type Option func(*RunConfig) error

// WithCertificate specifies the certificate options for the run
//	WithCertificate("client.crt", "client.key")
func WithCertificate(cert, key string) Option {
	return func(o *RunConfig) error {
		cert = strings.TrimSpace(cert)
		key = strings.TrimSpace(key)

		o.cert = cert
		o.key = key

		return nil
	}
}

// WithServerNameOverride specifies the certificate options for the run
func WithServerNameOverride(cname string) Option {
	return func(o *RunConfig) error {
		o.cname = cname

		return nil
	}
}

// WithAuthority specifies the value to be used as the :authority pseudo-header.
// This only works with WithInsecure option.
func WithAuthority(authority string) Option {
	return func(o *RunConfig) error {
		o.authority = authority

		return nil
	}
}

// WithRootCertificate specifies the root certificate options for the run
//	WithRootCertificate("ca.crt")
func WithRootCertificate(cert string) Option {
	return func(o *RunConfig) error {
		cert = strings.TrimSpace(cert)

		o.cacert = cert

		return nil
	}
}

// WithInsecure specifies that this run should be done using insecure mode
//	WithInsecure(true)
func WithInsecure(insec bool) Option {
	return func(o *RunConfig) error {
		o.insecure = insec

		return nil
	}
}

// WithSkipTLSVerify skip client side TLS verification of server certificate
func WithSkipTLSVerify(skip bool) Option {
	return func(o *RunConfig) error {
		o.skipVerify = skip

		return nil
	}
}

// WithTotalRequests specifies the N (number of total requests) setting
//	WithTotalRequests(1000)
func WithTotalRequests(n uint) Option {
	return func(o *RunConfig) error {
		o.n = int(n)

		return nil
	}
}

// WithConcurrency specifies the C (number of concurrent requests) option
//	WithConcurrency(20)
func WithConcurrency(c uint) Option {
	return func(o *RunConfig) error {
		o.c = int(c)

		return nil
	}
}

// WithQPS specifies the QPS (queries per second) limit option
//	WithQPS(10)
func WithQPS(qps uint) Option {
	return func(o *RunConfig) error {
		o.qps = int(qps)

		return nil
	}
}

// WithRunDuration specifies the Z (total test duration) option
//	WithRunDuration(time.Duration(2*time.Minute))
func WithRunDuration(z time.Duration) Option {
	return func(o *RunConfig) error {
		o.z = z

		return nil
	}
}

// WithDurationStopAction specifies how run duration (Z) timeout is handled
// Possible options are "close", "ignore", and "wait"
//	WithDurationStopAction("ignore")
func WithDurationStopAction(action string) Option {
	return func(o *RunConfig) error {
		action = strings.ToLower(action)

		if action == "close" || action == "wait" || action == "ignore" {
			o.zstop = action
		}

		return nil
	}
}

// WithTimeout specifies the timeout for each request
//	WithTimeout(time.Duration(20*time.Second))
func WithTimeout(timeout time.Duration) Option {
	return func(o *RunConfig) error {
		o.timeout = timeout

		return nil
	}
}

// WithDialTimeout specifies the initial connection dial timeout
//	WithDialTimeout(time.Duration(20*time.Second))
func WithDialTimeout(dt time.Duration) Option {
	return func(o *RunConfig) error {
		o.dialTimeout = dt

		return nil
	}
}

// WithKeepalive specifies the keepalive timeout
//	WithKeepalive(time.Duration(1*time.Minute))
func WithKeepalive(k time.Duration) Option {
	return func(o *RunConfig) error {
		o.keepaliveTime = k

		return nil
	}
}

// WithBinaryData specifies the binary data
//	msg := &helloworld.HelloRequest{}
//	msg.Name = "bob"
//	binData, _ := proto.Marshal(msg)
//	WithBinaryData(binData)
func WithBinaryData(data []byte) Option {
	return func(o *RunConfig) error {
		o.data = data
		o.binary = true

		return nil
	}
}

// WithBinaryDataFromFile specifies the binary data
//	WithBinaryDataFromFile("request_data.bin")
func WithBinaryDataFromFile(path string) Option {
	return func(o *RunConfig) error {
		data, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		o.data = data
		o.binary = true

		return nil
	}
}

// WithDataFromJSON loads JSON data from string
//	WithDataFromJSON(`{"name":"bob"}`)
func WithDataFromJSON(data string) Option {
	return func(o *RunConfig) error {
		o.data = []byte(data)
		o.binary = false

		return nil
	}
}

// WithData specifies data as generic data that can be serailized to JSON
func WithData(data interface{}) Option {
	return func(o *RunConfig) error {
		dataJSON, err := json.Marshal(data)

		if err != nil {
			return err
		}

		o.data = dataJSON
		o.binary = false

		return nil
	}
}

// WithDataFromReader loads JSON data from reader
// 	file, _ := os.Open("data.json")
// 	WithDataFromReader(file)
func WithDataFromReader(r io.Reader) Option {
	return func(o *RunConfig) error {
		data, err := ioutil.ReadAll(r)
		if err != nil {
			return err
		}

		o.data = data
		o.binary = false

		return nil
	}
}

// WithDataFromFile loads JSON data from file
//	WithDataFromFile("data.json")
func WithDataFromFile(path string) Option {
	return func(o *RunConfig) error {
		data, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		o.data = data
		o.binary = false

		return nil
	}
}

// WithMetadataFromJSON specifies the metadata to be read from JSON string
//	WithMetadataFromJSON(`{"request-id":"123"}`)
func WithMetadataFromJSON(md string) Option {
	return func(o *RunConfig) error {
		o.metadata = []byte(md)

		return nil
	}
}

// WithMetadata specifies the metadata to be used as a map
// 	md := make(map[string]string)
// 	md["token"] = "foobar"
// 	md["request-id"] = "123"
// 	WithMetadata(&md)
func WithMetadata(md *map[string]string) Option {
	return func(o *RunConfig) error {
		mdJSON, err := json.Marshal(md)
		if err != nil {
			return err
		}

		o.metadata = mdJSON

		return nil
	}
}

// WithMetadataFromFile loads JSON metadata from file
//	WithMetadataFromJSON("metadata.json")
func WithMetadataFromFile(path string) Option {
	return func(o *RunConfig) error {
		mdJSON, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		o.metadata = mdJSON

		return nil
	}
}

// WithName sets the name of the test run
//	WithName("greeter service test")
func WithName(name string) Option {
	return func(o *RunConfig) error {
		name = strings.TrimSpace(name)
		if name != "" {
			o.name = name
		}

		return nil
	}
}

// WithTags specifies the user defined tags as a map
// 	tags := make(map[string]string)
// 	tags["env"] = "staging"
// 	tags["created by"] = "joe developer"
// 	WithTags(&tags)
func WithTags(tags *map[string]string) Option {
	return func(o *RunConfig) error {
		tagsJSON, err := json.Marshal(tags)
		if err != nil {
			return err
		}

		o.tags = tagsJSON

		return nil
	}
}

// WithCPUs specifies the number of CPU's to be used
//	WithCPUs(4)
func WithCPUs(c uint) Option {
	return func(o *RunConfig) error {
		if c > 0 {
			o.cpus = int(c)
		}

		return nil
	}
}

// WithProtoFile specified proto file path and optionally import paths
// We will automatically add the proto file path's directory and the current directory
//	WithProtoFile("greeter.proto", []string{"/home/protos"})
func WithProtoFile(proto string, importPaths []string) Option {
	return func(o *RunConfig) error {
		proto = strings.TrimSpace(proto)
		if proto != "" {
			if filepath.Ext(proto) != ".proto" {
				return errors.Errorf(fmt.Sprintf("proto: must have .proto extension"))
			}

			o.proto = proto

			dir := filepath.Dir(proto)
			if dir != "." {
				o.importPaths = append(o.importPaths, dir)
			}

			o.importPaths = append(o.importPaths, ".")

			if len(importPaths) > 0 {
				o.importPaths = append(o.importPaths, importPaths...)
			}
		}

		return nil
	}
}

// WithProtoset specified protoset file path
//	WithProtoset("bundle.protoset")
func WithProtoset(protoset string) Option {
	return func(o *RunConfig) error {
		protoset = strings.TrimSpace(protoset)
		o.protoset = protoset

		return nil
	}
}

// WithStreamInterval sets the stream interval
func WithStreamInterval(d time.Duration) Option {
	return func(o *RunConfig) error {
		o.streamInterval = d

		return nil
	}
}

// WithReflectionMetadata specifies the metadata to be used as a map
// 	md := make(map[string]string)
// 	md["token"] = "foobar"
// 	md["request-id"] = "123"
// 	WithReflectionMetadata(&md)
func WithReflectionMetadata(md *map[string]string) Option {
	return func(o *RunConfig) error {
		o.rmd = md

		return nil
	}
}

// WithConnections specifies the number of gRPC connections to use
//	WithConnections(5)
func WithConnections(c uint) Option {
	return func(o *RunConfig) error {
		if c > 0 {
			o.nConns = int(c)
		}

		return nil
	}
}

// WithLogger specifies the logging option
func WithLogger(log Logger) Option {
	return func(o *RunConfig) error {
		o.log = log
		o.hasLog = true

		return nil
	}
}

func newConfig(call, host string, options ...Option) (*RunConfig, error) {
	call = strings.TrimSpace(call)
	host = strings.TrimSpace(host)

	// init with defaults
	c := &RunConfig{
		call:        call,
		host:        host,
		n:           200,
		c:           50,
		nConns:      1,
		timeout:     time.Duration(20 * time.Second),
		dialTimeout: time.Duration(10 * time.Second),
		cpus:        runtime.GOMAXPROCS(-1),
		zstop:       "close",
	}

	// apply options
	for _, option := range options {
		err := option(c)

		if err != nil {
			return nil, err
		}
	}

	// checks
	if c.nConns > c.c {
		return nil, errors.New("Number of connections cannot be greater than concurrency")
	}

	if c.call == "" {
		return nil, errors.New("Call required")
	}

	if c.host == "" {
		return nil, errors.New("Host required")
	}

	creds, err := createClientTransportCredentials(
		c.skipVerify,
		c.cacert,
		c.cert,
		c.key,
		c.cname,
	)

	if err != nil {
		return nil, err
	}

	c.creds = creds

	return c, nil
}

func createClientTransportCredentials(skipVerify bool, cacertFile, clientCertFile, clientKeyFile, cname string) (credentials.TransportCredentials, error) {
	var tlsConf tls.Config

	if clientCertFile != "" {
		// Load the client certificates from disk
		certificate, err := tls.LoadX509KeyPair(clientCertFile, clientKeyFile)
		if err != nil {
			return nil, fmt.Errorf("could not load client key pair: %v", err)
		}
		tlsConf.Certificates = []tls.Certificate{certificate}
	}

	if skipVerify {
		tlsConf.InsecureSkipVerify = true
	} else if cacertFile != "" {
		// Create a certificate pool from the certificate authority
		certPool := x509.NewCertPool()
		ca, err := ioutil.ReadFile(cacertFile)
		if err != nil {
			return nil, fmt.Errorf("could not read ca certificate: %v", err)
		}

		// Append the certificates from the CA
		if ok := certPool.AppendCertsFromPEM(ca); !ok {
			return nil, errors.New("failed to append ca certs")
		}

		tlsConf.RootCAs = certPool
	}

	if cname != "" {
		tlsConf.ServerName = cname
	}

	return credentials.NewTLS(&tlsConf), nil
}
