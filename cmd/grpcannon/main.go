package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"time"

	"github.com/bojand/grpcannon"
	"github.com/bojand/grpcannon/config"
	"github.com/bojand/grpcannon/printer"
	"github.com/bojand/grpcannon/protodesc"
)

var (
	// set by goreleaser with -ldflags="-X main.version=..."
	version = "dev"

	proto = flag.String("proto", "", `The .proto file.`)
	call  = flag.String("call", "", `A fully-qualified symbol name.`)
	cert  = flag.String("cert", "", "Client certificate file. If Omitted insecure is used.")
	cname = flag.String("cname", "", "Server Cert CName Override - useful for self signed certs")

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

	v = flag.Bool("v", false, "Print the version.")

	localConfigName = "grpcannon.json"
)

var usage = `Usage: grpcannon [options...] <host>
Options:
  -proto	The protocol buffer file.
  -call		A fully-qualified method name in 'service/method' or 'service.method' format.
  -cert		The file containing the CA root cert file.
  -cname	an override of the expect Server Cname presented by the server.

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
  -M  Path for call metadata JSON file. For example, /home/user/metadata.json or ./metadata.json.
 
  -o  Output path. If none provided stdout is used.
  -O  Output type. If none provided, a summary is printed.
      "csv" is the only supported alternative. Dumps the response
	  metrics in comma-separated values format.
	  
  -i  Comma separated list of proto import paths. The current working directory and the directory
	  of the protocol buffer file are automatically added to the import list.
	  
  -T  Connection timeout in seconds for the initial connection dial. Default is 10.
  -L  Keepalive time in seconds. Only used if present and above 0.

  -cpus		Number of used cpu cores. (default for current machine is %d cores)

  -v  Print the version.
`

func main() {
	flag.Usage = func() {
		fmt.Fprint(os.Stderr, fmt.Sprintf(usage, runtime.NumCPU()))
	}

	flag.Parse()

	if *v {
		fmt.Println(version)
		os.Exit(0)
	}

	if flag.NArg() < 1 {
		usageAndExit("")
	}

	host := flag.Args()[0]

	var cfg *config.Config

	if _, err := os.Stat(localConfigName); err == nil {
		cfg, err = config.ReadConfig(localConfigName)
		if err != nil {
			errAndExit(err.Error())
		}
	} else {
		iPaths := []string{}
		pathsTrimmed := strings.TrimSpace(*paths)
		if pathsTrimmed != "" {
			iPaths = strings.Split(pathsTrimmed, ",")
		}

		cfg, err = config.New(*proto, *call, *cert, *cname, *n, *c, *q, *z, *t,
			*data, *dataPath, *md, *mdPath, *output, *format, host, *ct, *kt, *cpus, iPaths)
		if err != nil {
			errAndExit(err.Error())
		}
	}

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

func runTest(config *config.Config) (*grpcannon.Report, error) {
	mtd, err := protodesc.GetMethodDesc(config.Call, config.Proto, config.ImportPaths)
	if err != nil {
		return nil, err
	}

	opts := &grpcannon.Options{
		Host:          config.Host,
		Cert:          config.Cert,
		CName:		   config.CName,
		N:             config.N,
		C:             config.C,
		QPS:           config.QPS,
		Z:             config.Z,
		Timeout:       config.Timeout,
		DialTimtout:   config.DialTimeout,
		KeepaliveTime: config.KeepaliveTime,
		Data:          config.Data,
		Metadata:      config.Metadata,
	}

	reqr, err := grpcannon.New(mtd, opts)
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
