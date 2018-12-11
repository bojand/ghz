package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/bojand/ghz"
	"github.com/bojand/ghz/printer"
	"github.com/jinzhu/configor"
)

var (
	// set by goreleaser with -ldflags="-X main.version=..."
	version = "dev"

	proto    = flag.String("proto", "", `The .proto file.`)
	protoset = flag.String("protoset", "", `The .protoset file.`)
	call     = flag.String("call", "", `A fully-qualified symbol name.`)
	cert     = flag.String("cert", "", "Client certificate file. If Omitted insecure is used.")
	cname    = flag.String("cname", "", "Server Cert CName Override - useful for self signed certs.")
	cPath    = flag.String("config", "", "Path to the config JSON file.")
	insecure = flag.Bool("insecure", false, "Specify for non TLS connection")

	c = flag.Uint("c", 50, "Number of requests to run concurrently.")
	n = flag.Uint("n", 200, "Number of requests to run. Default is 200.")
	q = flag.Uint("q", 0, "Rate limit, in queries per second (QPS). Default is no rate limit.")
	t = flag.Uint("t", 20, "Timeout for each request in seconds.")
	z = flag.Duration("z", 0, "Duration of application to send requests.")
	x = flag.Duration("x", 0, "Maximum duration of application to send requests.")

	data     = flag.String("d", "", "The call data as stringified JSON. If the value is '@' then the request contents are read from stdin.")
	dataPath = flag.String("D", "", "Path for call data JSON file.")
	binData  = flag.Bool("b", false, "The call data as serialized binary message read from stdin.")
	binPath  = flag.String("B", "", "The call data as serialized binary message read from a file.")
	md       = flag.String("m", "", "Request metadata as stringified JSON.")
	mdPath   = flag.String("M", "", "Path for call metadata JSON file.")

	paths = flag.String("i", "", "Comma separated list of proto import paths")

	output = flag.String("o", "", "Output path")
	format = flag.String("O", "", "Output format")

	ct = flag.Uint("T", 10, "Connection timeout in seconds for the initial connection dial.")
	kt = flag.Uint("L", 0, "Keepalive time in seconds.")

	name = flag.String("name", "", "Name of the test.")

	cpus = flag.Uint("cpus", uint(runtime.GOMAXPROCS(-1)), "")

	v = flag.Bool("v", false, "Print the version.")
)

var usage = `Usage: ghz [options...] host
Options:

-proto	The protocol buffer file.
-protoset	The compiled protoset file. Alternative to proto. -proto takes precedence.
-call		A fully-qualified method name in 'service/method' or 'service.method' format.
-cert		The file containing the CA root cert file.
-cname	An override of the expect Server Cname presented by the server.
-config	Path to the config JSON file
-insecure     Specify for non TLS connection

-c  Number of requests to run concurrently. Total number of requests cannot
be smaller than the concurrency level. Default is 50.
-n  Number of requests to run. Default is 200.
-q  Rate limit, in queries per second (QPS). Default is no rate limit.
-t  Timeout for each request in seconds. Default is 20, use 0 for infinite.
-z  Duration of application to send requests. When duration is reached,
application stops and exits. If duration is specified, n is ignored.
Examples: -z 10s -z 3m.
-x  Maximum duration of application to send requests with n setting respected.
If duration is reached before n requests are completed, application stops and exits.
Examples: -x 10s -x 3m.

-d  The call data as stringified JSON.
If the value is '@' then the request contents are read from stdin.
-D  Path for call data JSON file. For example, /home/user/file.json or ./file.json.
-b  The call data comes as serialized binary message read from stdin.
-B  Path for the call data as serialized binary message.
-m  Request metadata as stringified JSON.
-M  Path for call metadata JSON file. For example, /home/user/metadata.json or ./metadata.json.

-o  Output path. If none provided stdout is used.
-O  Output type. If none provided, a summary is printed.
"csv" outputs the response metrics in comma-separated values format.
"json" outputs the metrics report in JSON format.
"pretty" outputs the metrics report in pretty JSON format.
"html" outputs the metrics report as HTML.
"influx-summary" outputs the metrics summary as influxdb line protocol.
"influx-details" outputs the metrics details as influxdb line protocol.

-i  Comma separated list of proto import paths. The current working directory and the directory
of the protocol buffer file are automatically added to the import list.

-T  Connection timeout in seconds for the initial connection dial. Default is 10.
-L  Keepalive time in seconds. Only used if present and above 0.

-name  Name of the test.

-cpus  Number of used cpu cores. (default for current machine is %d cores)

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

	cfgPath := strings.TrimSpace(*cPath)

	var cfg *config

	if cfgPath != "" {
		var conf config
		err := configor.Load(&conf, cfgPath)
		if err != nil {
			errAndExit(err.Error())
		}

		cfg = &conf
	} else {
		if flag.NArg() < 1 {
			usageAndExit("")
		}

		var err error
		cfg, err = createConfigFromArgs()
		if err != nil {
			errAndExit(err.Error())
		}
	}

	// init / fix up durations
	if cfg.X.Duration > 0 {
		cfg.Z.Duration = cfg.X.Duration
	} else if cfg.Z.Duration > 0 {
		cfg.N = math.MaxInt32
	}

	// set up all the options
	options := make([]ghz.Option, 0, 15)

	options = append(options,
		ghz.WithProtoFile(cfg.Proto, cfg.ImportPaths),
		ghz.WithProtoset(cfg.Protoset),
		ghz.WithCertificate(cfg.Cert, cfg.CName),
		ghz.WithInsecure(cfg.Insecure),
		ghz.WithConcurrency(cfg.C),
		ghz.WithTotalRequests(cfg.N),
		ghz.WithQPS(cfg.QPS),
		ghz.WithTimeout(time.Duration(cfg.Timeout)*time.Second),
		ghz.WithRunDuration(cfg.Z.Duration),
		ghz.WithDialTimeout(time.Duration(cfg.DialTimeout)*time.Second),
		ghz.WithKeepalive(time.Duration(cfg.KeepaliveTime)*time.Second),
		ghz.WithName(cfg.Name),
		ghz.WithCPUs(cfg.CPUs),
		ghz.WithMetadata(cfg.Metadata),
	)

	if strings.TrimSpace(cfg.MetadataPath) != "" {
		options = append(options, ghz.WithMetadataFromFile(strings.TrimSpace(cfg.MetadataPath)))
	}

	// data
	if dataStr, ok := cfg.Data.(string); ok && dataStr == "@" {
		options = append(options, ghz.WithDataFromReader(os.Stdin))
	} else if strings.TrimSpace(cfg.DataPath) != "" {
		options = append(options, ghz.WithDataFromFile(strings.TrimSpace(cfg.DataPath)))
	} else {
		options = append(options, ghz.WithData(cfg.Data))
	}

	// or binary data
	if len(cfg.BinData) > 0 {
		options = append(options, ghz.WithBinaryData(cfg.BinData))
	}
	if len(cfg.BinDataPath) > 0 {
		options = append(options, ghz.WithBinaryDataFromFile(cfg.BinDataPath))
	}

	report, err := ghz.Run(cfg.Call, cfg.Host, options...)
	if err != nil {
		errAndExit(err.Error())
	}

	output := os.Stdout
	outputPath := strings.TrimSpace(cfg.Output)
	if outputPath != "" {
		f, err := os.Create(outputPath)
		if err != nil {
			errAndExit(err.Error())
		}
		defer f.Close()
		output = f
	}

	p := printer.ReportPrinter{
		Report: report,
		Out:    output}

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

func createConfigFromArgs() (*config, error) {
	host := flag.Args()[0]

	iPaths := []string{}
	pathsTrimmed := strings.TrimSpace(*paths)
	if pathsTrimmed != "" {
		iPaths = strings.Split(pathsTrimmed, ",")
	}

	var binaryData []byte
	if *binData {
		b, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			return nil, err
		}

		binaryData = b
	}

	var metadata *map[string]string
	*md = strings.TrimSpace(*md)
	if *md != "" {
		if err := json.Unmarshal([]byte(*md), metadata); err != nil {
			return nil, err
		}
	}

	var dataObj interface{}
	if *data != "@" && strings.TrimSpace(*data) != "" {
		if err := json.Unmarshal([]byte(*data), dataObj); err != nil {
			return nil, err
		}
	}

	cfg := &config{
		Host:          host,
		Proto:         *proto,
		Protoset:      *protoset,
		Call:          *call,
		Cert:          *cert,
		CName:         *cname,
		N:             *n,
		C:             *c,
		QPS:           *q,
		Z:             duration{*z},
		X:             duration{*x},
		Timeout:       *t,
		Data:          dataObj,
		DataPath:      *dataPath,
		BinData:       binaryData,
		BinDataPath:   *binPath,
		Metadata:      metadata,
		MetadataPath:  *mdPath,
		Output:        *output,
		Format:        *format,
		ImportPaths:   iPaths,
		DialTimeout:   *ct,
		KeepaliveTime: *kt,
		CPUs:          *cpus,
		Insecure:      *insecure,
		Name:          *name,
	}

	return cfg, nil
}
