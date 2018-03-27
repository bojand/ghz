package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/protoparse"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/jhump/protoreflect/dynamic/grpcdynamic"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/stats"
	"google.golang.org/grpc/status"
)

type rpcStatsTagKey string

const rpcStatsTagID = rpcStatsTagKey("id")

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
// TODO add connetion timeout

var (
	proto  = flag.String("proto", "", `The .proto file.`)
	call   = flag.String("call", "", `A fully-qualified symbol name.`)
	cacert = flag.String("cacert", "", "Root certificate file.")
	cert   = flag.String("cert", "", "Client certificate file. If Omitted insecure is used.")
	key    = flag.String("key", "", "Private key file.")

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
			Cert:     *cert,
			N:        *n,
			C:        *c,
			QPS:      *q,
			Z:        *z,
			Timeout:  *t,
			DataPath: *dataPath,
			// Metadata:     *md,
			MetadataPath: *mdPath,
			Format:       *format,
			Host:         host,
			CPUs:         *cpus}

		err := config.SetData(*data)
		if err != nil {
			errAndExit(err.Error())
		}

		err = config.SetMetadata(*md)
		if err != nil {
			errAndExit(err.Error())
		}

		err = config.InitData()
		if err != nil {
			errAndExit(err.Error())
		}

		err = config.InitMetadata()
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
	// config.ProtoFile = file

	config.ImportPaths = append(config.ImportPaths, filepath.Dir(config.Proto), ".")

	fmt.Printf("host: %s\nproto: %s\ncall: %s\nimports:%s\ndata:%+v\n", host, config.Proto, config.Call, config.ImportPaths, config.Data)

	// resp, err := doCall(config)
	// if err != nil {
	// 	errAndExit(err.Error())
	// }

	// fmt.Printf("Response: %s Duration: %+v\n", resp.GetResponseString(), resp.Duration)

	err = doReq(config)
	if err != nil {
		errAndExit(err.Error())
	}
	fmt.Printf("Done!")
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

func doReq(config *Config) error {
	mtd, err := getMethodDesc(config)
	if err != nil {
		return err
	}

	reqr, err := New(config, mtd)
	if err != nil {
		return err
	}

	return reqr.Run()
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
	// create credentials
	credOptions, err := CreateClientCredOption(config)
	if err != nil {
		return nil, err
	}

	// create client connection
	cc, err := grpc.Dial(config.Host, grpc.WithStatsHandler(&ClientHandler{}), credOptions)
	if err != nil {
		return nil, fmt.Errorf("Failed to create client to %s: %s", config.Host, err.Error())
	}
	defer cc.Close()

	// the client stub for the connection
	stub := grpcdynamic.NewStub(cc)

	// create call context
	ctx := context.Background()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	timeout := time.Duration(int64(config.Timeout) * int64(time.Second))
	ctx, _ = context.WithTimeout(ctx, timeout)

	// include the metadata
	if config.Metadata != nil && len(*config.Metadata) > 0 {
		reqMD := metadata.New(*config.Metadata)
		ctx = metadata.NewOutgoingContext(ctx, reqMD)
	}
	var respHeaders metadata.MD
	var respTrailers metadata.MD

	// payload
	input := dynamic.NewMessage(mtd.GetInputType())
	for k, v := range *config.Data {
		input.TrySetFieldByName(k, v)
	}

	start := time.Now()
	resp, err := stub.InvokeRpc(ctx, mtd, input, grpc.Trailer(&respTrailers), grpc.Header(&respHeaders))
	_, ok := status.FromError(err)
	if !ok || err != nil {
		return nil, err
	}

	end := time.Now()
	duration := end.Sub(start)

	DoParallelRequests(&stub, mtd)

	return &CallResult{resp, mtd, &duration}, nil
}

// DoParallelRequests testing concurrent gRPC requests
func DoParallelRequests(stub *grpcdynamic.Stub, mtd *desc.MethodDescriptor) {
	log.Println("==================")
	var wg sync.WaitGroup
	wg.Add(5)

	for i := 0; i < 5; i++ {
		go func(counter int) {
			defer wg.Done()

			log.Printf("start %d\n", counter)

			input := dynamic.NewMessage(mtd.GetInputType())
			input.TrySetFieldByName("name", fmt.Sprintf("Msg %d", counter))

			ctx := context.Background()

			// ctx = context.WithValue(ctx, rpcStatsTagID, counter)

			start := time.Now()

			resp, err := stub.InvokeRpc(ctx, mtd, input)
			_, ok := status.FromError(err)

			end := time.Now()
			duration := end.Sub(start)

			if resp != nil && err == nil && ok {
				log.Printf("response for %d: %+v\n", counter, resp)
			}

			log.Printf("duration for %d: %+v\n", counter, duration)

		}(i)
	}
	wg.Wait()
	log.Println("==================")
}

// // CreateClientCredOption creates the credential dial options based on config
// func CreateClientCredOption(config *Config) (grpc.DialOption, error) {
// 	credOptions := grpc.WithInsecure()
// 	if strings.TrimSpace(config.Cert) != "" {
// 		creds, err := credentials.NewClientTLSFromFile(config.Cert, "")
// 		if err != nil {
// 			return nil, err
// 		}
// 		credOptions = grpc.WithTransportCredentials(creds)
// 	}

// 	return credOptions, nil
// }

// ClientHandler is for gRPC stats
type ClientHandler struct{}

// HandleConn handle the connection
func (c *ClientHandler) HandleConn(ctx context.Context, cs stats.ConnStats) {
	// no-op
}

// TagConn exists to satisfy gRPC stats.Handler.
func (c *ClientHandler) TagConn(ctx context.Context, cti *stats.ConnTagInfo) context.Context {
	// no-op
	return ctx
}

// HandleRPC implements per-RPC tracing and stats instrumentation.
func (c *ClientHandler) HandleRPC(ctx context.Context, rs stats.RPCStats) {

	// switch st := rs.(type) {
	switch rs.(type) {
	case *stats.Begin:
		// log.Printf("Begin: %+v\n %+v\n", ctx, rs)

	// case *stats.OutHeader:
	// 	log.Println("OutHeader")
	// case *stats.InHeader:
	// 	log.Println("InHeader")
	// case *stats.InTrailer:
	// 	log.Println("InTrailer")
	// case *stats.OutTrailer:
	// 	log.Println("OutTrailer")
	// case *stats.OutPayload:
	// 	log.Println("OutPayload")
	// case *stats.InPayload:
	// log.Println("InPayload")
	case *stats.End:
		// log.Printf("End: %+v %+v\n", ctx, rs)
		idValue, ok := ctx.Value(rpcStatsTagID).(int)
		if ok {
			// log.Printf("[End] ID value: %+v\n", idValue)
			startID := fmt.Sprintf("start_%v", idValue)
			startValue, ok := ctx.Value(rpcStatsTagKey(startID)).(time.Time)
			if ok {
				end := time.Now()
				duration := end.Sub(startValue)
				log.Printf("[End] Duration for %+v: %+v\n", startID, duration)
			}
		}
		// default:
		// log.Println("unexpected stats: %T", st)
	}
}

// HandleRPC2 implements per-RPC tracing and stats instrumentation.
func (c *ClientHandler) HandleRPC2(ctx context.Context, rs stats.RPCStats) {

	switch st := rs.(type) {
	case *stats.Begin:
		log.Println("Begin")
	case *stats.OutHeader:
		log.Println("OutHeader")
	case *stats.InHeader:
		log.Println("InHeader")
	case *stats.InTrailer:
		log.Println("InTrailer")
	case *stats.OutTrailer:
		log.Println("OutTrailer")
	case *stats.OutPayload:
		log.Println("OutPayload")
	case *stats.InPayload:
		log.Println("InPayload")
	case *stats.End:
		log.Println("End")
	default:
		log.Println("unexpected stats: %T", st)
	}
}

// TagRPC implements per-RPC context management.
func (c *ClientHandler) TagRPC(ctx context.Context, info *stats.RPCTagInfo) context.Context {
	if info == nil {
		return ctx
	}

	// idValue, ok := ctx.Value(rpcStatsTagID).(int)
	// if ok {
	idValue := rand.Intn(100000)
	ctx = context.WithValue(ctx, rpcStatsTagID, idValue)
	startID := fmt.Sprintf("start_%v", idValue)
	rpcStatsTagStart := rpcStatsTagKey(startID)
	start := time.Now()
	ctx = context.WithValue(ctx, rpcStatsTagStart, start)
	// }

	return ctx
}
