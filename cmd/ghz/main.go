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

	"github.com/bojand/ghz/printer"
	"github.com/bojand/ghz/runner"
	"github.com/jinzhu/configor"
)

var (
	// set by goreleaser with -ldflags="-X main.version=..."
	version = "dev"

	cPath = flag.String("config", "", "Path to the JSON or TOML config file that specifies all the test run settings.")

	proto    = flag.String("proto", "", `The Protocol Buffer .proto file.`)
	protoset = flag.String("protoset", "", `The compiled protoset file. Alternative to proto. -proto takes precedence.`)
	call     = flag.String("call", "", `A fully-qualified method name in 'package/service/method' or 'package.service.method' format.`)
	paths    = flag.String("i", "", "Comma separated list of proto import paths. The current working directory and the directory of the protocol buffer file are automatically added to the import list.")

	cacert     = flag.String("cacert", "", "File containing trusted root certificates for verifying the server.")
	cert       = flag.String("cert", "", "File containing client certificate (public key), to present to the server. Must also provide -key option.")
	key        = flag.String("key", "", "File containing client private key, to present to the server. Must also provide -cert option.")
	cname      = flag.String("cname", "", "Server name override when validating TLS certificate - useful for self signed certs.")
	skipVerify = flag.Bool("skipTLS", false, "Skip TLS client verification of the server's certificate chain and host name.")
	insecure   = flag.Bool("insecure", false, "Use plaintext and insecure connection.")
	authority  = flag.String("authority", "", "Value to be used as the :authority pseudo-header. Only works if -insecure is used.")

	c = flag.Uint("c", 50, "Number of requests to run concurrently. Total number of requests cannot be smaller than the concurrency level. Default is 50.")
	n = flag.Uint("n", 200, "Number of requests to run. Default is 200.")
	q = flag.Uint("q", 0, "Rate limit, in queries per second (QPS). Default is no rate limit.")
	t = flag.Uint("t", 20, "Timeout for each request in seconds. Default is 20, use 0 for infinite.")
	z = flag.Duration("z", 0, "Duration of application to send requests. When duration is reached, application stops and exits. If duration is specified, n is ignored. Examples: -z 10s -z 3m.")
	x = flag.Duration("x", 0, "Maximum duration of application to send requests with n setting respected. If duration is reached before n requests are completed, application stops and exits. Examples: -x 10s -x 3m.")

	data     = flag.String("d", "", "The call data as stringified JSON. If the value is '@' then the request contents are read from stdin.")
	dataPath = flag.String("D", "", "File path for call data JSON file. Examples: /home/user/file.json or ./file.json.")
	binData  = flag.Bool("b", false, "The call data comes as serialized binary message read from stdin.")
	binPath  = flag.String("B", "", "File path for the call data as serialized binary message.")
	md       = flag.String("m", "", "Request metadata as stringified JSON.")
	mdPath   = flag.String("M", "", "File path for call metadata JSON file. Examples: /home/user/metadata.json or ./metadata.json.")
	si       = flag.Duration("si", 0, "Interval for stream requests between message sends.")

	output = flag.String("o", "", "Output path. If none provided stdout is used.")
	format = flag.String("O", "", "Output format. If none provided, a summary is printed.")

	ct = flag.Uint("T", 10, "Connection timeout in seconds for the initial connection dial. Default is 10.")
	kt = flag.Uint("L", 0, "Keepalive time in seconds. Only used if present and above 0.")

	name = flag.String("name", "", "User specified name for the test.")
	tags = flag.String("tags", "", "JSON representation of user-defined string tags.")

	cpus = flag.Uint("cpus", uint(runtime.GOMAXPROCS(-1)), "Number of used cpu cores.")

	v = flag.Bool("v", false, "Print the version.")
)

var usage = `Usage: ghz [options...] host
Options:

-config	Path to the JSON or TOML config file that specifies all the test run settings.

-proto		The Protocol Buffer .proto file.
-protoset	The compiled protoset file. Alternative to proto. -proto takes precedence.
-call		A fully-qualified method name in 'package/service/method' or 'package.service.method' format.
-i		Comma separated list of proto import paths. The current working directory and the directory
		of the protocol buffer file are automatically added to the import list.
	
-cacert		File containing trusted root certificates for verifying the server.
-cert		File containing client certificate (public key), to present to the server. Must also provide -key option.
-key 		File containing client private key, to present to the server. Must also provide -cert option.
-cname		Server name override when validating TLS certificate - useful for self signed certs.
-skipTLS	Skip TLS client verification of the server's certificate chain and host name.
-insecure	Use plaintext and insecure connection.
-authority	Value to be used as the :authority pseudo-header. Only works if -insecure is used.

-c  Number of requests to run concurrently. 
    Total number of requests cannot be smaller than the concurrency level. Default is 50.
-n  Number of requests to run. Default is 200.
-q  Rate limit, in queries per second (QPS). Default is no rate limit.
-t  Timeout for each request in seconds. Default is 20, use 0 for infinite.
-z  Duration of application to send requests. When duration is reached, application stops and exits.
    If duration is specified, n is ignored. Examples: -z 10s -z 3m.
-x  Maximum duration of application to send requests with n setting respected.
    If duration is reached before n requests are completed, application stops and exits.
    Examples: -x 10s -x 3m.

-d  The call data as stringified JSON.
    If the value is '@' then the request contents are read from stdin.
-D  Path for call data JSON file. Examples: /home/user/file.json or ./file.json.
-b  The call data comes as serialized binary message read from stdin.
-B  Path for the call data as serialized binary message.
-m  Request metadata as stringified JSON.
-M  Path for call metadata JSON file. Examples: /home/user/metadata.json or ./metadata.json.

-si Stream interval duration. Spread stream sends by given amount. 
    Only applies to client and bidi streaming calls. Example: 100ms

-o  Output path. If none provided stdout is used.
-O  Output type. If none provided, a summary is printed.
    "csv" outputs the response metrics in comma-separated values format.
    "json" outputs the metrics report in JSON format.
    "pretty" outputs the metrics report in pretty JSON format.
    "html" outputs the metrics report as HTML.
    "influx-summary" outputs the metrics summary as influxdb line protocol.
    "influx-details" outputs the metrics details as influxdb line protocol.

-T  Connection timeout in seconds for the initial connection dial. Default is 10.
-L  Keepalive time in seconds. Only used if present and above 0.

-name  User specified name for the test.
-tags  JSON representation of user-defined string tags.

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
	if cfg.X > 0 {
		cfg.Z = cfg.X
	} else if cfg.Z > 0 {
		cfg.N = math.MaxInt32
	}

	// set up all the options
	options := make([]runner.Option, 0, 15)

	options = append(options,
		runner.WithProtoFile(cfg.Proto, cfg.ImportPaths),
		runner.WithProtoset(cfg.Protoset),
		runner.WithRootCertificate(cfg.RootCert),
		runner.WithCertificate(cfg.Cert, cfg.Key),
		runner.WithServerNameOverride(cfg.CName),
		runner.WithSkipTLSVerify(cfg.SkipTLSVerify),
		runner.WithInsecure(cfg.Insecure),
		runner.WithAuthority(cfg.Authority),
		runner.WithConcurrency(cfg.C),
		runner.WithTotalRequests(cfg.N),
		runner.WithQPS(cfg.QPS),
		runner.WithTimeout(time.Duration(cfg.Timeout)*time.Second),
		runner.WithRunDuration(time.Duration(cfg.Z)),
		runner.WithDialTimeout(time.Duration(cfg.DialTimeout)*time.Second),
		runner.WithKeepalive(time.Duration(cfg.KeepaliveTime)*time.Second),
		runner.WithName(cfg.Name),
		runner.WithCPUs(cfg.CPUs),
		runner.WithMetadata(cfg.Metadata),
		runner.WithTags(cfg.Tags),
		runner.WithStreamInterval(time.Duration(cfg.SI)),
	)

	if strings.TrimSpace(cfg.MetadataPath) != "" {
		options = append(options, runner.WithMetadataFromFile(strings.TrimSpace(cfg.MetadataPath)))
	}

	// data
	if dataStr, ok := cfg.Data.(string); ok && dataStr == "@" {
		options = append(options, runner.WithDataFromReader(os.Stdin))
	} else if strings.TrimSpace(cfg.DataPath) != "" {
		options = append(options, runner.WithDataFromFile(strings.TrimSpace(cfg.DataPath)))
	} else {
		options = append(options, runner.WithData(cfg.Data))
	}

	// or binary data
	if len(cfg.BinData) > 0 {
		options = append(options, runner.WithBinaryData(cfg.BinData))
	}
	if len(cfg.BinDataPath) > 0 {
		options = append(options, runner.WithBinaryDataFromFile(cfg.BinDataPath))
	}

	report, err := runner.Run(cfg.Call, cfg.Host, options...)
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

	var metadata map[string]string
	*md = strings.TrimSpace(*md)
	if *md != "" {
		if err := json.Unmarshal([]byte(*md), &metadata); err != nil {
			return nil, fmt.Errorf("Error unmarshaling metadata '%v': %v", *md, err.Error())
		}
	}

	var dataObj interface{}
	if *data != "@" && strings.TrimSpace(*data) != "" {
		if err := json.Unmarshal([]byte(*data), &dataObj); err != nil {
			return nil, fmt.Errorf("Error unmarshaling data '%v': %v", *data, err.Error())
		}
	}

	var tagsMap map[string]string
	*tags = strings.TrimSpace(*tags)
	if *tags != "" {
		if err := json.Unmarshal([]byte(*tags), &tagsMap); err != nil {
			return nil, fmt.Errorf("Error unmarshaling tags '%v': %v", *md, err.Error())
		}
	}

	cfg := &config{
		Host:          host,
		Proto:         *proto,
		Protoset:      *protoset,
		Call:          *call,
		RootCert:      *cacert,
		Cert:          *cert,
		Key:           *key,
		SkipTLSVerify: *skipVerify,
		Insecure:      *insecure,
		Authority:     *authority,
		CName:         *cname,
		N:             *n,
		C:             *c,
		QPS:           *q,
		Z:             Duration(*z),
		X:             Duration(*x),
		Timeout:       *t,
		Data:          dataObj,
		DataPath:      *dataPath,
		BinData:       binaryData,
		BinDataPath:   *binPath,
		Metadata:      &metadata,
		MetadataPath:  *mdPath,
		SI:            Duration(*si),
		Output:        *output,
		Format:        *format,
		ImportPaths:   iPaths,
		DialTimeout:   *ct,
		KeepaliveTime: *kt,
		CPUs:          *cpus,
		Name:          *name,
		Tags:          &tagsMap,
	}

	return cfg, nil
}
