package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/alecthomas/kingpin"
	"github.com/bojand/ghz/printer"
	"github.com/bojand/ghz/runner"
	"github.com/jinzhu/configor"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	// set by goreleaser with -ldflags="-X main.version=..."
	version = "dev"

	nCPUs = runtime.GOMAXPROCS(-1)

	cPath = kingpin.Flag("config", "Path to the JSON or TOML config file that specifies all the test run settings.").PlaceHolder(" ").String()

	// Proto
	isProtoSet = false
	proto      = kingpin.Flag("proto", `The Protocol Buffer .proto file.`).
			PlaceHolder(" ").IsSetByUser(&isProtoSet).String()

	isProtoSetSet = false
	protoset      = kingpin.Flag("protoset", "The compiled protoset file. Alternative to proto. -proto takes precedence.").
			PlaceHolder(" ").IsSetByUser(&isProtoSetSet).String()

	isCallSet = false
	call      = kingpin.Flag("call", `A fully-qualified method name in 'package.Service/method' or 'package.Service.Method' format.`).
			PlaceHolder(" ").IsSetByUser(&isCallSet).String()

	isImportSet = false
	paths       = kingpin.Flag("import-paths", "Comma separated list of proto import paths. The current working directory and the directory of the protocol buffer file are automatically added to the import list.").
			Short('i').PlaceHolder(" ").IsSetByUser(&isImportSet).String()

	// Security
	isCACertSet = false
	cacert      = kingpin.Flag("cacert", "File containing trusted root certificates for verifying the server.").
			PlaceHolder(" ").IsSetByUser(&isCACertSet).String()

	isCertSet = false
	cert      = kingpin.Flag("cert", "File containing client certificate (public key), to present to the server. Must also provide -key option.").
			PlaceHolder(" ").IsSetByUser(&isCertSet).String()

	isKeySet = false
	key      = kingpin.Flag("key", "File containing client private key, to present to the server. Must also provide -cert option.").
			PlaceHolder(" ").IsSetByUser(&isKeySet).String()

	isCNameSet = false
	cname      = kingpin.Flag("cname", "Server name override when validating TLS certificate - useful for self signed certs.").
			PlaceHolder(" ").IsSetByUser(&isCNameSet).String()

	isSkipSet  = false
	skipVerify = kingpin.Flag("skipTLS", "Skip TLS client verification of the server's certificate chain and host name.").
			Default("false").IsSetByUser(&isSkipSet).Bool()

	isInsecSet = false
	insecure   = kingpin.Flag("insecure", "Use plaintext and insecure connection.").
			Default("false").IsSetByUser(&isInsecSet).Bool()

	isAuthSet = false
	authority = kingpin.Flag("authority", "Value to be used as the :authority pseudo-header. Only works if -insecure is used.").
			PlaceHolder(" ").IsSetByUser(&isAuthSet).String()

	// Run
	isCSet = false
	c      = kingpin.Flag("concurrency", "Number of requests to run concurrently. Total number of requests cannot be smaller than the concurrency level. Default is 50.").
		Short('c').Default("50").IsSetByUser(&isCSet).Uint()

	isNSet = false
	n      = kingpin.Flag("total", "Number of requests to run. Default is 200.").
		Short('n').Default("200").IsSetByUser(&isNSet).Uint()

	isQSet = false
	q      = kingpin.Flag("qps", "Rate limit, in queries per second (QPS). Default is no rate limit.").
		Default("0").Short('q').IsSetByUser(&isQSet).Uint()

	isTSet = false
	t      = kingpin.Flag("timeout", "Timeout for each request. Default is 20s, use 0 for infinite.").
		Default("20s").Short('t').IsSetByUser(&isTSet).Duration()

	isZSet = false
	z      = kingpin.Flag("duration", "Duration of application to send requests. When duration is reached, application stops and exits. If duration is specified, n is ignored. Examples: -z 10s -z 3m.").
		Short('z').Default("0").IsSetByUser(&isZSet).Duration()

	isXSet = false
	x      = kingpin.Flag("max-duration", "Maximum duration of application to send requests with n setting respected. If duration is reached before n requests are completed, application stops and exits. Examples: -x 10s -x 3m.").
		Short('x').Default("0").IsSetByUser(&isXSet).Duration()

	isZStopSet = false
	zstop      = kingpin.Flag("duration-stop", "Specifies how duration stop is reported. Options are close, wait or ignore.").
			Default("close").IsSetByUser(&isZStopSet).String()

	// Data
	isDataSet = false
	data      = kingpin.Flag("data", "The call data as stringified JSON. If the value is '@' then the request contents are read from stdin.").
			Short('d').PlaceHolder(" ").IsSetByUser(&isDataSet).String()

	isDataPathSet = false
	dataPath      = kingpin.Flag("data-file", "File path for call data JSON file. Examples: /home/user/file.json or ./file.json.").
			Short('D').PlaceHolder("PATH").PlaceHolder(" ").IsSetByUser(&isDataPathSet).String()

	isBinDataSet = false
	binData      = kingpin.Flag("binary", "The call data comes as serialized binary message or multiple count-prefixed messages read from stdin.").
			Short('b').Default("false").IsSetByUser(&isBinDataSet).Bool()

	isBinDataPathSet = false
	binPath          = kingpin.Flag("binary-file", "File path for the call data as serialized binary message or multiple count-prefixed messages.").
				Short('B').PlaceHolder(" ").IsSetByUser(&isBinDataPathSet).String()

	isMDSet = false
	md      = kingpin.Flag("metadata", "Request metadata as stringified JSON.").
		Short('m').PlaceHolder(" ").IsSetByUser(&isMDSet).String()

	isMDPathSet = false
	mdPath      = kingpin.Flag("metadata-file", "File path for call metadata JSON file. Examples: /home/user/metadata.json or ./metadata.json.").
			Short('M').PlaceHolder(" ").IsSetByUser(&isMDPathSet).String()

	isSISet = false
	si      = kingpin.Flag("stream-interval", "Interval for stream requests between message sends.").
		Default("0").IsSetByUser(&isSISet).Duration()

	isRMDSet = false
	rmd      = kingpin.Flag("reflect-metadata", "Reflect metadata as stringified JSON used only for reflection request.").
			PlaceHolder(" ").IsSetByUser(&isRMDSet).String()

	// Output
	isOutputSet = false
	output      = kingpin.Flag("output", "Output path. If none provided stdout is used.").
			Short('o').PlaceHolder(" ").IsSetByUser(&isOutputSet).String()

	isFormatSet = false
	format      = kingpin.Flag("format", "Output format. One of: summary, csv, json, pretty, html, influx-summary, influx-details. Default is summary.").
			Short('O').Default("summary").PlaceHolder(" ").IsSetByUser(&isFormatSet).Enum("summary", "csv", "json", "pretty", "html", "influx-summary", "influx-details")

	// Connection
	isConnSet = false
	conns     = kingpin.Flag("connections", "Number of connections to use. Concurrency is distributed evenly among all the connections. Default is 1.").
			Default("1").IsSetByUser(&isConnSet).Uint()

	isCTSet = false
	ct      = kingpin.Flag("connect-timeout", "Connection timeout for the initial connection dial. Default is 10s.").
		Default("10s").IsSetByUser(&isCTSet).Duration()

	isKTSet = false
	kt      = kingpin.Flag("keepalive", "Keepalive time duration. Only used if present and above 0.").
		Default("0").IsSetByUser(&isKTSet).Duration()

	// Meta
	isNameSet = false
	name      = kingpin.Flag("name", "User specified name for the test.").
			PlaceHolder(" ").IsSetByUser(&isNameSet).String()

	isTagsSet = false
	tags      = kingpin.Flag("tags", "JSON representation of user-defined string tags.").
			PlaceHolder(" ").IsSetByUser(&isTagsSet).String()

	isCPUSet = false
	cpus     = kingpin.Flag("cpus", "Number of cpu cores to use.").
			Default(strconv.FormatUint(uint64(nCPUs), 10)).IsSetByUser(&isCPUSet).Uint()

	// Debug
	isDebugSet = false
	debug      = kingpin.Flag("debug", "The path to debug log file.").
			PlaceHolder(" ").IsSetByUser(&isDebugSet).String()

	isHostSet = false
	host      = kingpin.Arg("host", "Host and port to test.").String()

	isEnableCompressionSet = false
	enableCompression      = kingpin.Flag("enable-compression", "Enable Gzip compression on requests.").
				Short('e').Default("false").IsSetByUser(&isEnableCompressionSet).Bool()
)

func main() {
	kingpin.Version(version)
	kingpin.CommandLine.HelpFlag.Short('h')
	kingpin.CommandLine.VersionFlag.Short('v')
	kingpin.Parse()

	isHostSet = *host != ""

	cfgPath := strings.TrimSpace(*cPath)

	var cfg config

	if cfgPath != "" {
		err := configor.Load(&cfg, cfgPath)
		kingpin.FatalIfError(err, "")

		args := os.Args[1:]
		if len(args) > 1 {
			var cmdCfg config
			err = createConfigFromArgs(&cmdCfg)
			kingpin.FatalIfError(err, "")

			err = mergeConfig(&cfg, &cmdCfg)
			kingpin.FatalIfError(err, "")
		}
	} else {
		err := createConfigFromArgs(&cfg)

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
		runner.WithTimeout(time.Duration(cfg.Timeout)),
		runner.WithRunDuration(time.Duration(cfg.Z)),
		runner.WithDurationStopAction(cfg.ZStop),
		runner.WithDialTimeout(time.Duration(cfg.DialTimeout)),
		runner.WithKeepalive(time.Duration(cfg.KeepaliveTime)),
		runner.WithName(cfg.Name),
		runner.WithCPUs(cfg.CPUs),
		runner.WithMetadata(cfg.Metadata),
		runner.WithTags(cfg.Tags),
		runner.WithStreamInterval(time.Duration(cfg.SI)),
		runner.WithReflectionMetadata(cfg.ReflectMetadata),
		runner.WithConnections(cfg.Connections),
		runner.WithEnableCompression(cfg.EnableCompression),
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

	var logger *zap.SugaredLogger

	if len(cfg.Debug) > 0 {
		var err error
		logger, err = createLogger(cfg.Debug)
		kingpin.FatalIfError(err, "")

		defer logger.Sync()

		options = append(options, runner.WithLogger(logger))
	}

	if logger != nil {
		logger.Debugw("Start Run", "config", cfg)
	}

	report, err := runner.Run(cfg.Call, cfg.Host, options...)
	if err != nil {
		if logger != nil {
			logger.Errorf("Error from run: %+v", err.Error())
		}

		handleError(err)
	}

	output := os.Stdout
	outputPath := strings.TrimSpace(cfg.Output)

	if logger != nil {
		logger.Debug("Run finished")
	}

	if outputPath != "" {
		f, err := os.Create(outputPath)
		if err != nil {
			if logger != nil {
				logger.Errorw("Error opening file "+outputPath+": "+err.Error(),
					"error", err)
			}

			handleError(err)
		}

		defer func() {
			handleError(f.Close())
		}()

		output = f
	}

	if logger != nil {
		logPath := "stdout"
		if outputPath != "" {
			logPath = outputPath
		}

		logger.Debugw("Printing report to "+logPath, "path", logPath)
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

func createConfigFromArgs(cfg *config) error {
	if cfg == nil {
		return errors.New("config cannot be nil")
	}

	iPaths := []string{}
	pathsTrimmed := strings.TrimSpace(*paths)
	if pathsTrimmed != "" {
		iPaths = strings.Split(pathsTrimmed, ",")
	}

	var binaryData []byte
	if *binData {
		b, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			return err
		}

		binaryData = b
	}

	var metadata map[string]string
	*md = strings.TrimSpace(*md)
	if *md != "" {
		if err := json.Unmarshal([]byte(*md), &metadata); err != nil {
			return fmt.Errorf("Error unmarshaling metadata '%v': %v", *md, err.Error())
		}
	}

	var dataObj interface{}
	if *data != "@" && strings.TrimSpace(*data) != "" {
		if err := json.Unmarshal([]byte(*data), &dataObj); err != nil {
			return fmt.Errorf("Error unmarshaling data '%v': %v", *data, err.Error())
		}
	}

	var tagsMap map[string]string
	*tags = strings.TrimSpace(*tags)
	if *tags != "" {
		if err := json.Unmarshal([]byte(*tags), &tagsMap); err != nil {
			return fmt.Errorf("Error unmarshaling tags '%v': %v", *tags, err.Error())
		}
	}

	var rmdMap map[string]string
	*rmd = strings.TrimSpace(*rmd)
	if *rmd != "" {
		if err := json.Unmarshal([]byte(*rmd), &rmdMap); err != nil {
			return fmt.Errorf("Error unmarshaling reflection metadata '%v': %v", *rmd, err.Error())
		}
	}

	cfg.Host = *host
	cfg.Proto = *proto
	cfg.Protoset = *protoset
	cfg.Call = *call
	cfg.RootCert = *cacert
	cfg.Cert = *cert
	cfg.Key = *key
	cfg.SkipTLSVerify = *skipVerify
	cfg.Insecure = *insecure
	cfg.Authority = *authority
	cfg.CName = *cname
	cfg.N = *n
	cfg.C = *c
	cfg.QPS = *q
	cfg.Z = Duration(*z)
	cfg.X = Duration(*x)
	cfg.Timeout = Duration(*t)
	cfg.ZStop = *zstop
	cfg.Data = dataObj
	cfg.DataPath = *dataPath
	cfg.BinData = binaryData
	cfg.BinDataPath = *binPath
	cfg.Metadata = &metadata
	cfg.MetadataPath = *mdPath
	cfg.SI = Duration(*si)
	cfg.Output = *output
	cfg.Format = *format
	cfg.ImportPaths = iPaths
	cfg.Connections = *conns
	cfg.DialTimeout = Duration(*ct)
	cfg.KeepaliveTime = Duration(*kt)
	cfg.CPUs = *cpus
	cfg.Name = *name
	cfg.Tags = &tagsMap
	cfg.ReflectMetadata = &rmdMap
	cfg.Debug = *debug
	cfg.EnableCompression = *enableCompression

	return nil
}

func mergeConfig(dest *config, src *config) error {
	if src == nil || dest == nil {
		return errors.New("config cannot be nil")
	}

	if isProtoSet {
		dest.Proto = src.Proto
	}

	if isProtoSetSet {
		dest.Protoset = src.Protoset
	}

	if isCallSet {
		dest.Call = src.Call
	}

	if isCACertSet {
		dest.RootCert = src.RootCert
	}

	if isCertSet {
		dest.Cert = src.Cert
	}

	if isKeySet {
		dest.Key = src.Key
	}

	if isSkipSet {
		dest.SkipTLSVerify = src.SkipTLSVerify
	}

	if isInsecSet {
		dest.Insecure = src.Insecure
	}

	if isAuthSet {
		dest.Authority = src.Authority
	}

	if isCNameSet {
		dest.CName = src.CName
	}

	if isNSet {
		dest.N = src.N
	}

	if isCSet {
		dest.C = src.C
	}

	if isQSet {
		dest.QPS = src.QPS
	}

	if isZSet {
		dest.Z = src.Z
	}

	if isXSet {
		dest.X = src.X
	}

	if isTSet {
		dest.Timeout = src.Timeout
	}

	if isZStopSet {
		dest.ZStop = src.ZStop
	}

	if isDataSet {
		dest.Data = src.Data
	}

	if isDataPathSet {
		dest.DataPath = src.DataPath
	}

	if isBinDataSet {
		dest.BinData = src.BinData
	}

	if isBinDataPathSet {
		dest.BinDataPath = src.BinDataPath
	}

	if isMDSet {
		dest.Metadata = src.Metadata
	}

	if isMDPathSet {
		dest.MetadataPath = src.MetadataPath
	}

	if isSISet {
		dest.SI = src.SI
	}

	if isOutputSet {
		dest.Output = src.Output
	}

	if isFormatSet {
		dest.Format = src.Format
	}

	if isImportSet {
		dest.ImportPaths = src.ImportPaths
	}

	if isConnSet {
		dest.Connections = src.Connections
	}

	if isCTSet {
		dest.DialTimeout = src.DialTimeout
	}

	if isKTSet {
		dest.KeepaliveTime = src.KeepaliveTime
	}

	if isCPUSet {
		dest.CPUs = src.CPUs
	}

	if isNameSet {
		dest.Name = src.Name
	}

	if isTagsSet {
		dest.Tags = src.Tags
	}

	if isRMDSet {
		dest.ReflectMetadata = src.ReflectMetadata
	}

	if isDebugSet {
		dest.Debug = src.Debug
	}

	if isHostSet {
		dest.Host = src.Host
	}

	return nil
}

// createLogger creates a new zap logger
func createLogger(path string) (*zap.SugaredLogger, error) {

	var encoderCfg zapcore.EncoderConfig
	var cfg zap.Config

	encoderCfg = zap.NewProductionEncoderConfig()
	cfg = zap.NewProductionConfig()

	encoderCfg.LevelKey = "level"
	encoderCfg.MessageKey = "message"
	encoderCfg.CallerKey = ""
	encoderCfg.TimeKey = "time"
	encoderCfg.EncodeTime = zapcore.RFC3339NanoTimeEncoder
	encoderCfg.EncodeCaller = nil

	cfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	cfg.EncoderConfig = encoderCfg
	cfg.OutputPaths = []string{path}
	cfg.ErrorOutputPaths = []string{path}

	dl, err := cfg.Build()
	if err != nil {
		return nil, err
	}

	return dl.Sugar(), nil
}
