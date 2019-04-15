package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/bojand/ghz/printer"
	"github.com/bojand/ghz/runner"
	"github.com/jinzhu/configor"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	// set by goreleaser with -ldflags="-X main.version=..."
	version = "dev"

	nCPUs = runtime.GOMAXPROCS(-1)

	cPath = kingpin.Flag("config", "Path to the JSON or TOML config file that specifies all the test run settings.").PlaceHolder(" ").String()

	proto    = kingpin.Flag("proto", `The Protocol Buffer .proto file.`).PlaceHolder(" ").String()
	protoset = kingpin.Flag("protoset", "The compiled protoset file. Alternative to proto. -proto takes precedence.").PlaceHolder(" ").String()
	call     = kingpin.Flag("call", `A fully-qualified method name in 'package.Service/method' or 'package.Service.Method' format.`).PlaceHolder(" ").String()
	paths    = kingpin.Flag("import-paths", "Comma separated list of proto import paths. The current working directory and the directory of the protocol buffer file are automatically added to the import list.").Short('i').PlaceHolder(" ").String()

	cacert     = kingpin.Flag("cacert", "File containing trusted root certificates for verifying the server.").PlaceHolder(" ").String()
	cert       = kingpin.Flag("cert", "File containing client certificate (public key), to present to the server. Must also provide -key option.").PlaceHolder(" ").String()
	key        = kingpin.Flag("key", "File containing client private key, to present to the server. Must also provide -cert option.").PlaceHolder(" ").String()
	cname      = kingpin.Flag("cname", "Server name override when validating TLS certificate - useful for self signed certs.").PlaceHolder(" ").String()
	skipVerify = kingpin.Flag("skipTLS", "Skip TLS client verification of the server's certificate chain and host name.").Default("false").Bool()
	insecure   = kingpin.Flag("insecure", "Use plaintext and insecure connection.").Default("false").Bool()
	authority  = kingpin.Flag("authority", "Value to be used as the :authority pseudo-header. Only works if -insecure is used.").PlaceHolder(" ").String()

	c = kingpin.Flag("concurrency", "Number of requests to run concurrently. Total number of requests cannot be smaller than the concurrency level. Default is 50.").Short('c').Default("50").Uint()
	n = kingpin.Flag("total", "Number of requests to run. Default is 200.").Short('n').Default("200").Uint()
	q = kingpin.Flag("qps", "Rate limit, in queries per second (QPS). Default is no rate limit.").Default("0").Short('q').Uint()
	t = kingpin.Flag("timeout", "Timeout for each request in seconds. Default is 20, use 0 for infinite.").Default("20").Short('t').Uint()
	z = kingpin.Flag("duration", "Duration of application to send requests. When duration is reached, application stops and exits. If duration is specified, n is ignored. Examples: -z 10s -z 3m.").Short('z').Default("0").Duration()
	x = kingpin.Flag("max-duration", "Maximum duration of application to send requests with n setting respected. If duration is reached before n requests are completed, application stops and exits. Examples: -x 10s -x 3m.").Short('x').Default("0").Duration()

	conns = kingpin.Flag("connections", "Number of connections to use. Concurrency is distributed evenly among all the connections. Default is 1.").Default("1").Uint()

	data     = kingpin.Flag("data", "The call data as stringified JSON. If the value is '@' then the request contents are read from stdin.").Short('d').PlaceHolder(" ").String()
	dataPath = kingpin.Flag("data-file", "File path for call data JSON file. Examples: /home/user/file.json or ./file.json.").Short('D').PlaceHolder("PATH").PlaceHolder(" ").String()
	binData  = kingpin.Flag("binary", "The call data comes as serialized binary message read from stdin.").Short('b').Default("false").Bool()
	binPath  = kingpin.Flag("binary-file", "File path for the call data as serialized binary message.").Short('B').PlaceHolder(" ").String()
	md       = kingpin.Flag("metadata", "Request metadata as stringified JSON.").Short('m').PlaceHolder(" ").String()
	mdPath   = kingpin.Flag("metadata-file", "File path for call metadata JSON file. Examples: /home/user/metadata.json or ./metadata.json.").Short('M').PlaceHolder(" ").String()
	si       = kingpin.Flag("stream-interval", "Interval for stream requests between message sends.").Default("0").Duration()
	rmd      = kingpin.Flag("reflect-metadata", "Reflect metadata as stringified JSON used only for reflection request.").PlaceHolder(" ").String()

	output = kingpin.Flag("output", "Output path. If none provided stdout is used.").Short('o').PlaceHolder(" ").String()
	format = kingpin.Flag("format", "Output format. One of: summary, csv, json, pretty, html, influx-summary, influx-details. Default is summary.").Short('O').Default("summary").PlaceHolder(" ").Enum("summary", "csv", "json", "pretty", "html", "influx-summary", "influx-details")

	ct = kingpin.Flag("connect-timeout", "Connection timeout in seconds for the initial connection dial. Default is 10.").Default("10").Uint()
	kt = kingpin.Flag("keepalive", "Keepalive time in seconds. Only used if present and above 0.").Default("0").Uint()

	name = kingpin.Flag("name", "User specified name for the test.").PlaceHolder(" ").String()
	tags = kingpin.Flag("tags", "JSON representation of user-defined string tags.").PlaceHolder(" ").String()

	cpus = kingpin.Flag("cpus", "Number of cpu cores to use.").Default(strconv.FormatUint(uint64(nCPUs), 10)).Uint()

	host = kingpin.Arg("host", "Host and port to test.").String()
)

func main() {
	kingpin.Version(version)
	kingpin.CommandLine.HelpFlag.Short('h')
	kingpin.CommandLine.VersionFlag.Short('v')
	kingpin.Parse()

	cfgPath := strings.TrimSpace(*cPath)

	var cfg *config

	if cfgPath != "" {
		var conf config
		err := configor.Load(&conf, cfgPath)
		kingpin.FatalIfError(err, "")

		cfg = &conf
	} else {

		var err error
		cfg, err = createConfigFromArgs()
		kingpin.FatalIfError(err, "")
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
		runner.WithReflectionMetadata(cfg.ReflectMetadata),
		runner.WithConnections(cfg.Connections),
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
		handleError(err)
	}

	output := os.Stdout
	outputPath := strings.TrimSpace(cfg.Output)
	if outputPath != "" {
		f, err := os.Create(outputPath)
		if err != nil {
			handleError(err)
		}

		defer func() {
			handleError(f.Close())
		}()

		output = f
	}

	p := printer.ReportPrinter{
		Report: report,
		Out:    output,
	}

	handleError(p.Print(cfg.Format))
}

func handleError(err error) {
	if err != nil {
		if errString := err.Error(); errString != "" {
			fmt.Fprintln(os.Stderr, errString)
		}
		os.Exit(1)
	}
}

func createConfigFromArgs() (*config, error) {
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
			return nil, fmt.Errorf("Error unmarshaling tags '%v': %v", *tags, err.Error())
		}
	}

	var rmdMap map[string]string
	*rmd = strings.TrimSpace(*rmd)
	if *rmd != "" {
		if err := json.Unmarshal([]byte(*rmd), &rmdMap); err != nil {
			return nil, fmt.Errorf("Error unmarshaling reflection metadata '%v': %v", *rmd, err.Error())
		}
	}

	cfg := &config{
		Host:            *host,
		Proto:           *proto,
		Protoset:        *protoset,
		Call:            *call,
		RootCert:        *cacert,
		Cert:            *cert,
		Key:             *key,
		SkipTLSVerify:   *skipVerify,
		Insecure:        *insecure,
		Authority:       *authority,
		CName:           *cname,
		N:               *n,
		C:               *c,
		Connections:     *conns,
		QPS:             *q,
		Z:               Duration(*z),
		X:               Duration(*x),
		Timeout:         *t,
		Data:            dataObj,
		DataPath:        *dataPath,
		BinData:         binaryData,
		BinDataPath:     *binPath,
		Metadata:        &metadata,
		MetadataPath:    *mdPath,
		SI:              Duration(*si),
		Output:          *output,
		Format:          *format,
		ImportPaths:     iPaths,
		DialTimeout:     *ct,
		KeepaliveTime:   *kt,
		CPUs:            *cpus,
		Name:            *name,
		Tags:            &tagsMap,
		ReflectMetadata: &rmdMap,
	}

	return cfg, nil
}
