package main

// TODO
// * Add support for data from stdin
// * Docs and more tests
// * goreleaser
// * Add more metrics such as duration of different parts and size

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/bojand/grpcannon"
	"github.com/bojand/grpcannon/printer"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/protoparse"
)

var (
	proto = flag.String("proto", "", `The .proto file.`)
	call  = flag.String("call", "", `A fully-qualified symbol name.`)
	cert  = flag.String("cert", "", "Client certificate file. If Omitted insecure is used.")

	c = flag.Int("c", 50, "Number of requests to run concurrently.")
	n = flag.Int("n", 200, "Number of requests to run. Default is 200.")
	q = flag.Int("q", 0, "Rate limit, in queries per second (QPS). Default is no rate limit.")
	t = flag.Int("t", 20, "Timeout for each request in seconds.")
	z = flag.Duration("z", 0, "")

	data     = flag.String("d", "", "The call data as stringified JSON.")
	dataPath = flag.String("D", "", "Path for call data JSON file.")
	md       = flag.String("m", "", "Request metadata as stringified JSON.")
	mdPath   = flag.String("M", "", "Path for call metadata JSON file.")

	paths = flag.String("i", "", "Comma separated list of proto import paths")

	output = flag.String("o", "", "Output path")
	format = flag.String("O", "", "Output format")

	ct = flag.Int("T", 10, "Connection timeout in seconds for the initial connection dial.")
	kt = flag.Int("L", 0, "Keepalive time in seconds.")

	cpus = flag.Int("cpus", runtime.GOMAXPROCS(-1), "")

	localConfigName = "grpcannon.json"
)

var usage = `Usage: grpcannon [options...] <host>
Options:
  -proto	The Protocol Buffer file
  -call		A fully-qualified method name in 'service/method' or 'service.method' format.
  -cert		File containing client certificate (public key), to present to the server.

  -c  Number of requests to run concurrently. Total number of requests cannot
	  be smaller than the concurrency level. Default is 50.
  -n  Number of requests to run. Default is 200.
  -q  Rate limit, in queries per second (QPS). Default is no rate limit.
  -t  Timeout for each request in seconds. Default is 20, use 0 for infinite.
  -z  Duration of application to send requests. When duration is reached,
      application stops and exits. If duration is specified, n is ignored.
      Examples: -z 10s -z 3m.
  
  -d  The call data as stringified JSON.
  -D  Path for call data JSON file. For example, /home/user/file.json or ./file.json.
  -m  Request metadata as stringified JSON.
  -M  Path for call data JSON file. For example, /home/user/metadata.json or ./metadata.json.
 
  -o  Output path. If none provided stdout is used.
  -O  Output type. If none provided, a summary is printed.
      "csv" is the only supported alternative. Dumps the response
	  metrics in comma-separated values format.
	  
  -i  Comma separated list of proto import paths. The current working directory and the directory
	  of the protocol buffer file are automatically added to the import list.
	  
  -T  Connection timeout in seconds for the initial connection dial. Default is 10.
  -L  Keepalive time in seconds. Only used if present and above 0.

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

	var cfg *Config

	if _, err := os.Stat(localConfigName); err == nil {
		cfg, err = ReadConfig(localConfigName)
		if err != nil {
			errAndExit(err.Error())
		}
	} else {

		iPaths := []string{}
		pathsTrimmed := strings.TrimSpace(*paths)
		if pathsTrimmed != "" {
			iPaths = strings.Split(pathsTrimmed, ",")
		}

		cfg, err = NewConfig(*proto, *call, *cert, *n, *c, *q, *z, *t,
			*data, *dataPath, *md, *mdPath, *output, *format, host, *ct, *kt, *cpus, iPaths)
		if err != nil {
			errAndExit(err.Error())
		}
	}

	file, err := os.Open(cfg.Proto)
	if err != nil {
		errAndExit(err.Error())
	}
	defer file.Close()

	runtime.GOMAXPROCS(cfg.CPUs)

	report, err := runTest(cfg)
	if err != nil {
		errAndExit(err.Error())
	}

	p := printer.ReportPrinter{
		Report: report,
		Out:    os.Stdout}

	p.Print(cfg.Format)
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

func runTest(config *Config) (*grpcannon.Report, error) {
	mtd, err := getMethodDesc(config)
	if err != nil {
		return nil, err
	}

	opts := &grpcannon.Options{
		Cert:          config.Cert,
		N:             config.N,
		C:             config.C,
		QPS:           config.QPS,
		Z:             config.Z,
		Timeout:       config.Timeout,
		DialTimtout:   config.DialTimeout,
		KeepaliveTime: config.KeepaliveTime,
		Host:          config.Host,
		Data:          config.Data,
		Metadata:      config.Metadata,
	}

	reqr, err := grpcannon.New(opts, mtd)
	if err != nil {
		return nil, err
	}

	cancel := make(chan os.Signal, 1)
	signal.Notify(cancel, os.Interrupt)
	go func() {
		<-cancel
		reqr.Stop()
	}()

	if config.Z > 0 {
		go func() {
			time.Sleep(config.Z)
			reqr.Stop()
			fmt.Printf("Stopped due to test timeout after %+v\n", config.Z)
		}()
	}

	return reqr.Run()
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
