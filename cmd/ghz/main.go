package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/alecthomas/kingpin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/bojand/ghz/printer"
	"github.com/bojand/ghz/runner"
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
	isAsyncSet = false
	async      = kingpin.Flag("async", "Make requests asynchronous as soon as possible. Does not wait for request to finish before sending next one.").
			Default("false").IsSetByUser(&isAsyncSet).Bool()

	isRPSSet = false
	rps      = kingpin.Flag("rps", "Requests per second (RPS) rate limit for constant load schedule. Default is no rate limit.").
			Default("0").Short('r').IsSetByUser(&isRPSSet).Uint()

	isScheduleSet = false
	schedule      = kingpin.Flag("load-schedule", "Specifies the load schedule. Options are const, step, or line. Default is const.").
			Default("const").IsSetByUser(&isScheduleSet).String()

	isLoadStartSet = false
	loadStart      = kingpin.Flag("load-start", "Specifies the RPS load start value for step or line schedules.").
			Default("0").IsSetByUser(&isLoadStartSet).Uint()

	isLoadStepSet = false
	loadStep      = kingpin.Flag("load-step", "Specifies the load step value or slope value.").
			Default("0").IsSetByUser(&isLoadStepSet).Int()

	isLoadEndSet = false
	loadEnd      = kingpin.Flag("load-end", "Specifies the load end value for step or line load schedules.").
			Default("0").IsSetByUser(&isLoadEndSet).Uint()

	isLoadStepDurSet = false
	loadStepDuration = kingpin.Flag("load-step-duration", "Specifies the load step duration value for step load schedule.").
				Default("0").IsSetByUser(&isLoadStepDurSet).Duration()

	isLoadMaxDurSet = false
	loadMaxDuration = kingpin.Flag("load-max-duration", "Specifies the max load duration value for step or line load schedule.").
			Default("0").IsSetByUser(&isLoadMaxDurSet).Duration()

	// Concurrency
	isCSet = false
	c      = kingpin.Flag("concurrency", "Number of request workers to run concurrently for const concurrency schedule. Default is 50.").
		Short('c').Default("50").IsSetByUser(&isCSet).Uint()

	isCScheduleSet = false
	cschdule       = kingpin.Flag("concurrency-schedule", "Concurrency change schedule. Options are const, step, or line. Default is const.").
			Default("const").IsSetByUser(&isCScheduleSet).String()

	isCStartSet = false
	cStart      = kingpin.Flag("concurrency-start", "Concurrency start value for step and line concurrency schedules.").
			Default("0").IsSetByUser(&isCStartSet).Uint()

	isCEndSet = false
	cEnd      = kingpin.Flag("concurrency-end", "Concurrency end value for step and line concurrency schedules.").
			Default("0").IsSetByUser(&isCEndSet).Uint()

	isCStepSet = false
	cstep      = kingpin.Flag("concurrency-step", "Concurrency step / slope value for step and line concurrency schedules.").
			Default("1").IsSetByUser(&isCStepSet).Int()

	isCStepDurSet = false
	cStepDuration = kingpin.Flag("concurrency-step-duration", "Specifies the concurrency step duration value for step concurrency schedule.").
			Default("0").IsSetByUser(&isCStepDurSet).Duration()

	isCMaxDurSet = false
	cMaxDuration = kingpin.Flag("concurrency-max-duration", "Specifies the max concurrency adjustment duration value for step or line concurrency schedule.").
			Default("0").IsSetByUser(&isCMaxDurSet).Duration()

	// Other
	isNSet = false
	n      = kingpin.Flag("total", "Number of requests to run. Default is 200.").
		Short('n').Default("200").IsSetByUser(&isNSet).Uint()

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
	zstop      = kingpin.Flag("duration-stop", "Specifies how duration stop is reported. Options are close, wait or ignore. Default is close.").
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

	isSkipFirstSet = false
	skipFirst      = kingpin.Flag("skipFirst", "Skip the first X requests when doing the results tally.").
			Default("0").IsSetByUser(&isSkipFirstSet).Uint()

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

	var cfg runner.Config

	if cfgPath != "" {
		err := runner.LoadConfig(cfgPath, &cfg)
		kingpin.FatalIfError(err, "")

		args := os.Args[1:]
		if len(args) > 1 {
			var cmdCfg runner.Config
			err = createConfigFromArgs(&cmdCfg)
			kingpin.FatalIfError(err, "")

			err = mergeConfig(&cfg, &cmdCfg)
			kingpin.FatalIfError(err, "")
		}
	} else {
		err := createConfigFromArgs(&cfg)

		kingpin.FatalIfError(err, "")
	}

	var logger *zap.SugaredLogger

	options := []runner.Option{runner.WithConfig(&cfg)}
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

func createConfigFromArgs(cfg *runner.Config) error {
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
	cfg.SkipFirst = *skipFirst
	cfg.Insecure = *insecure
	cfg.Authority = *authority
	cfg.CName = *cname
	cfg.N = *n
	cfg.C = *c
	cfg.RPS = *rps
	cfg.Z = runner.Duration(*z)
	cfg.X = runner.Duration(*x)
	cfg.Timeout = runner.Duration(*t)
	cfg.ZStop = *zstop
	cfg.Data = dataObj
	cfg.DataPath = *dataPath
	cfg.BinData = binaryData
	cfg.BinDataPath = *binPath
	cfg.Metadata = metadata
	cfg.MetadataPath = *mdPath
	cfg.SI = runner.Duration(*si)
	cfg.Output = *output
	cfg.Format = *format
	cfg.ImportPaths = iPaths
	cfg.Connections = *conns
	cfg.DialTimeout = runner.Duration(*ct)
	cfg.KeepaliveTime = runner.Duration(*kt)
	cfg.CPUs = *cpus
	cfg.Name = *name
	cfg.Tags = tagsMap
	cfg.ReflectMetadata = rmdMap
	cfg.Debug = *debug
	cfg.EnableCompression = *enableCompression
	cfg.LoadSchedule = *schedule
	cfg.LoadStart = *loadStart
	cfg.LoadStep = *loadStep
	cfg.LoadEnd = *loadEnd
	cfg.LoadStepDuration = runner.Duration(*loadStepDuration)
	cfg.LoadMaxDuration = runner.Duration(*loadMaxDuration)
	cfg.Async = *async
	cfg.CSchedule = *cschdule
	cfg.CStart = *cStart
	cfg.CStep = *cstep
	cfg.CEnd = *cEnd
	cfg.CStepDuration = runner.Duration(*cStepDuration)
	cfg.CMaxDuration = runner.Duration(*cMaxDuration)

	return nil
}

func mergeConfig(dest *runner.Config, src *runner.Config) error {
	if src == nil || dest == nil {
		return errors.New("config cannot be nil")
	}

	// proto

	if isProtoSet {
		dest.Proto = src.Proto
	}

	if isProtoSetSet {
		dest.Protoset = src.Protoset
	}

	if isCallSet {
		dest.Call = src.Call
	}

	// security

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

	if isSkipFirstSet {
		dest.SkipFirst = src.SkipFirst
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

	// run

	if isNSet {
		dest.N = src.N
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

	// data

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

	// other

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

	// load

	if isAsyncSet {
		dest.Async = src.Async
	}

	if isRPSSet {
		dest.RPS = src.RPS
	}

	if isScheduleSet {
		dest.LoadSchedule = src.LoadSchedule
	}

	if isLoadStartSet {
		dest.LoadStart = src.LoadStart
	}

	if isLoadStepSet {
		dest.LoadStep = src.LoadStep
	}

	if isLoadEndSet {
		dest.LoadEnd = src.LoadEnd
	}

	if isLoadStepDurSet {
		dest.LoadStepDuration = src.LoadStepDuration
	}

	if isLoadMaxDurSet {
		dest.LoadMaxDuration = src.LoadMaxDuration
	}

	// concurrency

	if isCSet {
		dest.C = src.C
	}

	if isCScheduleSet {
		dest.CSchedule = src.CSchedule
	}

	if isCStartSet {
		dest.CStart = src.CStart
	}

	if isCStepSet {
		dest.CStep = src.CStep
	}

	if isCEndSet {
		dest.CEnd = src.CEnd
	}

	if isCStepDurSet {
		dest.CStepDuration = src.CStepDuration
	}

	if isCMaxDurSet {
		dest.CMaxDuration = src.CMaxDuration
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
