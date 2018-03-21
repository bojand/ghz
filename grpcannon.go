package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/protoparse"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/jhump/protoreflect/dynamic/grpcdynamic"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// CallResult holds RPC call result
type CallResult struct {
	Response   interface{}
	MethodInfo *desc.MethodDescriptor
	Duration   *time.Duration
}

// GetResponseString return string of response if not server streaming
func (r *CallResult) GetResponseString() string {
	if r.Response != nil && !r.MethodInfo.IsServerStreaming() {
		if m, ok := r.Response.(fmt.Stringer); ok {
			return m.String()
		}
		return ""
	}
	return ""
}

// TODO add import paths option
// TODO add keepalive options

var (
	proto    = flag.String("proto", "", `The .proto file.`)
	call     = flag.String("call", "", `A fully-qualified symbol name.`)
	cacert   = flag.String("cacert", "", "Root certificate file.")
	cert     = flag.String("cert", "", "Client certificate file.")
	key      = flag.String("key", "", "Private key file.")
	insecure = flag.Bool("insecure", false, "Use insecure mode.")

	data     = flag.String("d", "", "The call data as stringified JSON.")
	dataPath = flag.String("D", "", "Path for call data JSON file.")
	md       = flag.String("m", "", "Request metadata as stringified JSON.")
	mdPath   = flag.String("M", "", "Path for call metadata JSON file.")

	format = flag.String("o", "", "Output format")

	c = flag.Int("c", 50, "Number of requests to run concurrently.")
	n = flag.Int("n", 200, "Number of requests to run. Default is 200.")
	q = flag.Int("q", 0, "Rate limit, in queries per second (QPS). Default is no rate limit.")
	t = flag.Int("t", 20, "Timeout for each request in seconds.")
	z = flag.Duration("z", 0, "")

	cpus = flag.Int("cpus", runtime.GOMAXPROCS(-1), "")

	localConfigName = "grpcannon.json"
)

var usage = `Usage: grpcannon [options...] <host>
Options:
  -proto	The Protocol Buffer file
  -call		A fully-qualified method name in 'service/method' or 'service.method' format.
  -cacert	File containing trusted root certificates for verifying the server.
  -cert		File containing client certificate (public key), to present to the server. 
			Must also provide -key option.
  -key 		File containing client private key, to present to the server. Must also provide -cert option.
  -insecure Use insecure mode. Ignores any of the cert options above.

  -n  Number of requests to run. Default is 200.
  -c  Number of requests to run concurrently. Total number of requests cannot
      be smaller than the concurrency level. Default is 50.
  -q  Rate limit, in queries per second (QPS). Default is no rate limit.
  -z  Duration of application to send requests. When duration is reached,
      application stops and exits. If duration is specified, n is ignored.
      Examples: -z 10s -z 3m.
  -t  Timeout for each request in seconds. Default is 20, use 0 for infinite.
  
  -d  The call data as stringified JSON.
  -D  Path for call data JSON file. For example, /home/user/file.json or ./file.json.
  -m  Request metadata as stringified JSON.
  -M  Path for call data JSON file. For example, /home/user/metadata.json or ./metadata.json.
 
  -o  Output type. If none provided, a summary is printed.
      "csv" is the only supported alternative. Dumps the response
      metrics in comma-separated values format.

  -cpus		Number of used cpu cores. (default for current machine is %d cores)
`

func main() {
	flag.Usage = func() {
		fmt.Fprint(os.Stderr, fmt.Sprintf(usage, runtime.NumCPU()))
	}

	flag.Parse()
	if flag.NArg() < 1 {
		usageAndExit("")
	}

	host := flag.Args()[0]

	var config *Config

	if _, err := os.Stat(localConfigName); err == nil {
		config, err = ReadConfig(localConfigName)
		if err != nil {
			errAndExit(err.Error())
		}
	} else {
		// TODO Fix up with .New()

		config = &Config{
			Proto:    *proto,
			Call:     *call,
			CACert:   *cacert,
			Cert:     *cert,
			Key:      *key,
			Insecure: *insecure,
			N:        *n,
			C:        *c,
			QPS:      *q,
			Z:        *z,
			Timeout:  *t,
			DataPath: *dataPath,
			Metadata: *md,
			MDPath:   *mdPath,
			Format:   *format,
			Host:     host,
			CPUs:     *cpus}

		err := config.SetData(*data)
		if err != nil {
			errAndExit(err.Error())
		}

		err = config.InitData()
		if err != nil {
			errAndExit(err.Error())
		}

		err = config.Validate()
		if err != nil {
			errAndExit(err.Error())
		}
	}

	file, err := os.Open(config.Proto)
	if err != nil {
		errAndExit(err.Error())
	}
	defer file.Close()
	config.ProtoFile = file

	config.ImportPaths = []string{filepath.Dir(config.Proto), "."}

	fmt.Printf("host: %s\nproto: %s\ncall: %s\nimports:%s\ndata:%+v\n", host, config.Proto, config.Call, config.ImportPaths, config.Data)

	resp, err := doCall(config)
	if err != nil {
		errAndExit(err.Error())
	}

	fmt.Printf("Response: %s Duration: %+v\n", resp.GetResponseString(), resp.Duration)
}

func errAndExit(msg string) {
	fmt.Fprintf(os.Stderr, msg)
	fmt.Fprintf(os.Stderr, "\n")
	os.Exit(1)
}

func usageAndExit(msg string) {
	if msg != "" {
		fmt.Fprintf(os.Stderr, msg)
		fmt.Fprintf(os.Stderr, "\n\n")
	}
	flag.Usage()
	fmt.Fprintf(os.Stderr, "\n")
	os.Exit(1)
}

func parseSymbol(svcAndMethod string) (string, string) {
	pos := strings.LastIndex(svcAndMethod, "/")
	if pos < 0 {
		pos = strings.LastIndex(svcAndMethod, ".")
		if pos < 0 {
			return "", ""
		}
	}
	return svcAndMethod[:pos], svcAndMethod[pos+1:]
}

func doCall(config *Config) (*CallResult, error) {
	mtd, err := getMethodDesc(config)
	if err != nil {
		return nil, err
	}
	if !mtd.IsClientStreaming() && !mtd.IsServerStreaming() {
		return invokeUnary(config, mtd)
	}
	return nil, errors.New("Unsupported call")
}

func getMethodDesc(config *Config) (*desc.MethodDescriptor, error) {
	p := &protoparse.Parser{ImportPaths: config.ImportPaths}

	fileName := filepath.Base(config.Proto)
	fds, err := p.ParseFiles(fileName)
	if err != nil {
		return nil, err
	}

	fileDesc := fds[0]

	svc, mth := parseSymbol(config.Call)
	if svc == "" || mth == "" {
		return nil, fmt.Errorf("given method name %q is not in expected format: 'service/method' or 'service.method'", config.Call)
	}

	dsc := fileDesc.FindSymbol(svc)
	if dsc == nil {
		return nil, fmt.Errorf("target server does not expose service %q", svc)
	}

	sd, ok := dsc.(*desc.ServiceDescriptor)
	if !ok {
		return nil, fmt.Errorf("target server does not expose service %q", svc)
	}

	mtd := sd.FindMethodByName(mth)
	if mtd == nil {
		return nil, fmt.Errorf("service %q does not include a method named %q", svc, mth)
	}

	return mtd, nil
}

func invokeUnary(config *Config, mtd *desc.MethodDescriptor) (*CallResult, error) {
	cc, err := grpc.Dial(config.Host, grpc.WithInsecure())
	if err != nil {
		return nil, fmt.Errorf("Failed to create client to %s: %s", config.Host, err.Error())
	}
	defer cc.Close()

	input := dynamic.NewMessage(mtd.GetInputType())
	// err = input.UnmarshalJSON([]byte(config.Data))
	// if err != nil {
	// 	return nil, errors.Wrap(err, "Invalid data JSON: ")
	// }

	for k, v := range *config.Data {
		input.TrySetFieldByName(k, v)
	}

	stub := grpcdynamic.NewStub(cc)

	start := time.Now()
	// var respHeaders metadata.MD
	// var respTrailers metadata.MD
	// resp, err := stub.InvokeRpc(context.Background(), mtd, input, grpc.Trailer(&respTrailers), grpc.Header(&respHeaders))
	resp, err := stub.InvokeRpc(context.Background(), mtd, input)
	_, ok := status.FromError(err)
	if !ok || err != nil {
		return nil, err
	}

	end := time.Now()
	duration := end.Sub(start)

	// log.Printf("%+v\n", respHeaders)
	// log.Printf("%+v\n", respTrailers)

	return &CallResult{resp, mtd, &duration}, nil
}
