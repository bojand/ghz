package main

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/alecthomas/kingpin"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/bojand/ghz/printer"
	"github.com/bojand/ghz/runner"
)

var (
	// set by goreleaser with -ldflags="-X main.version=..."
	version = "dev"
)

func main() {

	// var (
	// 	nCPUs = runtime.GOMAXPROCS(-1)

	// 	// config
	// 	cPath string

	// 	// proto
	// 	proto    string
	// 	protoset string
	// 	call     string
	// 	paths    []string

	// 	// security
	// 	cacert     string
	// 	cert       string
	// 	key        string
	// 	cname      string
	// 	skipVerify bool
	// 	insecure   bool
	// 	authority  string

	// 	// run
	// 	async bool
	// 	rps   uint

	// 	// load
	// 	loadSchedule     string
	// 	loadStart        uint
	// 	loadStep         int
	// 	loadEnd          uint
	// 	loadStepDuration time.Duration
	// 	loadMaxDuration  time.Duration

	// 	// concurrency
	// 	c             uint
	// 	cschdule      string
	// 	cStart        uint
	// 	cEnd          uint
	// 	cStep         int
	// 	cStepDuration time.Duration
	// 	cMaxDuration  time.Duration

	// 	// other
	// 	total          uint
	// 	requestTimeout time.Duration
	// 	totalDuration  time.Duration
	// 	maxDuration    time.Duration
	// 	zstop          string

	// 	// data
	// 	data     string
	// 	dataPath string
	// 	binData  bool
	// 	binPath  string
	// 	md       string
	// 	mdPath   string
	// 	si       time.Duration
	// 	scd      time.Duration
	// 	scc      uint
	// 	sdm      bool
	// 	rmd      string

	// 	// output
	// 	output      string
	// 	format      string
	// 	skipFirst   uint
	// 	countErrors bool

	// 	// connection
	// 	conns             uint
	// 	ct                time.Duration
	// 	kt                time.Duration
	// 	enableCompression bool
	// 	lbStrategy        string

	// 	// meta
	// 	name  string
	// 	tags  string
	// 	cpus  uint
	// 	debug string
	// )

	var (
		nCPUs = runtime.GOMAXPROCS(-1)

		// config
		cPath  string
		config runner.Config

		data    string
		binData bool
	)

	rootCmd := &cobra.Command{
		Use:   "ghz [host]",
		Short: "ghz description",
		Args:  cobra.ExactArgs(1),
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("on init")
			v := viper.New()

			if cPath != "" {
				fmt.Println("on init have config... setting it")
				v.SetConfigFile(cPath)
			}

			err := v.ReadInConfig()
			if err == nil {
				fmt.Println("Using config file:", v.ConfigFileUsed())
			}

			mergeFlags(cmd, v)

			return err
		},
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("ghz args: %v\n", args)
			fmt.Println("flags:", cPath, config.Proto, config.Protoset, config.Call)

			fmt.Println("proto changed: ", cmd.Flag("proto").Changed)
			fmt.Println("call changed: ", cmd.Flag("call").Changed)

			var logger *zap.SugaredLogger

			options := []runner.Option{runner.WithConfig(&config)}
			if len(config.Debug) > 0 {
				var err error
				logger, err = createLogger(config.Debug)
				kingpin.FatalIfError(err, "")

				defer logger.Sync()

				options = append(options, runner.WithLogger(logger))
			}

			if cmd.Flag("lb-strategy").Changed && config.Host != "" && !strings.HasPrefix(config.Host, "dns:///") {
				logger.Warn("Load balancing strategy set without using DNS (dns:///) scheme. Strategy: %v. Host: %+v.", config.LBStrategy, config.Host)
			}

			if logger != nil {
				logger.Debugw("Start Run", "config", config)
			}

			report, err := runner.Run(config.Call, config.Host, options...)
			if err != nil {
				if logger != nil {
					logger.Errorf("Error from run: %+v", err.Error())
				}

				handleError(err)
			}

			output := os.Stdout
			outputPath := strings.TrimSpace(config.Output)

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

			handleError(p.Print(config.Format))
		},
	}

	rootCmd.Flags().SortFlags = false
	rootCmd.PersistentFlags().SortFlags = false

	rootCmd.PersistentFlags().StringVar(&cPath, "config", "",
		"Path to the JSON or TOML config file that specifies all the test run settings.")

	// proto
	rootCmd.PersistentFlags().StringVar(&config.Proto, "proto", "", "The Protocol Buffer .proto file.")
	rootCmd.PersistentFlags().StringVar(&config.Protoset, "protoset", "",
		"The compiled protoset file. Alternative to proto. -proto takes precedence.")
	rootCmd.PersistentFlags().StringVar(&config.Call, "call", "",
		"A fully-qualified method name in 'package.Service/method' or 'package.Service.Method' format.")
	rootCmd.PersistentFlags().StringSliceVarP(&config.ImportPaths, "import-paths", "i", []string{},
		"Comma separated list of proto import paths. The current working directory and the directory of the protocol buffer file are automatically added to the import list.")

	// security
	rootCmd.PersistentFlags().StringVar(&config.RootCert, "cacert", "",
		"File containing trusted root certificates for verifying the server.")
	rootCmd.PersistentFlags().StringVar(&config.Cert, "cert", "",
		"File containing client certificate (public key), to present to the server. Must also provide -key option.")
	rootCmd.PersistentFlags().StringVar(&config.Key, "key", "",
		"File containing client private key, to present to the server. Must also provide -cert option.")
	rootCmd.PersistentFlags().StringVar(&config.CName, "cname", "",
		"Server name override when validating TLS certificate - useful for self signed certs.")
	rootCmd.PersistentFlags().BoolVar(&config.SkipTLSVerify, "skip-verify", false,
		"Skip TLS client verification of the server's certificate chain and host name.")
	rootCmd.PersistentFlags().BoolVar(&config.Insecure, "insecure", false,
		"Use plaintext and insecure connection.")
	rootCmd.PersistentFlags().StringVar(&config.Authority, "authority", "",
		"Value to be used as the :authority pseudo-header. Only works if -insecure is used.")

	// run
	rootCmd.PersistentFlags().BoolVar(&config.Async, "async", false,
		"Make requests asynchronous as soon as possible. Does not wait for request to finish before sending next one.")
	rootCmd.PersistentFlags().UintVarP(&config.RPS, "rps", "r", 0,
		"Requests per second (RPS) rate limit for constant load schedule. Default is no rate limit.")
	rootCmd.PersistentFlags().StringVar(&config.LoadSchedule, "load-schedule", "const",
		"Specifies the load schedule. Options are const, step, or line. Default is const.")
	rootCmd.PersistentFlags().UintVar(&config.LoadStart, "load-start", 0,
		"Specifies the RPS load start value for step or line schedules.")
	rootCmd.PersistentFlags().IntVar(&config.LoadStep, "load-step", 0,
		"Specifies the load step value or slope value.")
	rootCmd.PersistentFlags().UintVar(&config.LoadEnd, "load-end", 0,
		"Specifies the load end value for step or line load schedules.")
	rootCmd.PersistentFlags().DurationVar(&config.LoadStepDuration, "load-step-duration", 0,
		"Specifies the load step duration value for step load schedule.")
	rootCmd.PersistentFlags().DurationVar(&config.LoadMaxDuration, "load-max-duration", 0,
		"Specifies the max load duration value for step or line load schedule.")

	// concurrency
	rootCmd.PersistentFlags().UintVarP(&config.C, "concurrency", "c", 50,
		"Number of request workers to run concurrently for const concurrency schedule. Default is 50.")
	rootCmd.PersistentFlags().StringVar(&config.CSchedule, "concurrency-schedule", "const",
		"Concurrency change schedule. Options are const, step, or line. Default is const.")
	rootCmd.PersistentFlags().UintVar(&config.CStart, "concurrency-start", 0,
		"Concurrency start value for step and line concurrency schedules.")
	rootCmd.PersistentFlags().UintVar(&config.CEnd, "concurrency-end", 0,
		"Concurrency end value for step and line concurrency schedules..")
	rootCmd.PersistentFlags().IntVar(&config.CStep, "concurrency-step", 1,
		"Concurrency step / slope value for step and line concurrency schedules.")
	rootCmd.PersistentFlags().DurationVar(&config.CStepDuration, "concurrency-step-duration", 0,
		"Specifies the concurrency step duration value for step concurrency schedule.")
	rootCmd.PersistentFlags().DurationVar(&config.CMaxDuration, "concurrency-max-duration", 0,
		"Specifies the max concurrency adjustment duration value for step or line concurrency schedule.")

	// other
	rootCmd.PersistentFlags().UintVarP(&config.Total, "total", "n", 200,
		"Number of requests to run. Default is 200.")
	rootCmd.PersistentFlags().DurationVar(&config.Timeout, "timeout", time.Second*20,
		"Timeout for each request. Default is 20s, use 0 for infinite.")
	rootCmd.PersistentFlags().DurationVarP(&config.TotalDuration, "duration", "z", 0,
		"Duration of application to send requests. When duration is reached, application stops and exits. If duration is specified, n is ignored. Examples: -z 10s -z 3m.")
	rootCmd.PersistentFlags().DurationVarP(&config.MaxDuration, "max-duration", "x", 0,
		"Maximum duration of application to send requests with n setting respected. If duration is reached before n requests are completed, application stops and exits. Examples: -x 10s -x 3m.")
	rootCmd.PersistentFlags().StringVar(&config.ZStop, "duration-stop", "close",
		"Specifies how duration stop is reported. Options are close, wait or ignore. Default is close.")

	// data
	rootCmd.PersistentFlags().StringVarP(&data, "data", "d", "",
		"The call data as stringified JSON. If the value is '@' then the request contents are read from stdin.")
	rootCmd.PersistentFlags().StringVarP(&config.DataPath, "data-file", "D", "",
		"File path for call data JSON file. Examples: /home/user/file.json or ./file.json.")
	rootCmd.PersistentFlags().BoolVarP(&binData, "binary", "b", false,
		"The call data comes as serialized binary message or multiple count-prefixed messages read from stdin.")
	rootCmd.PersistentFlags().StringVarP(&config.BinDataPath, "binary-file", "B", "",
		"File path for the call data as serialized binary message or multiple count-prefixed messages.")
	rootCmd.PersistentFlags().StringToStringVarP(&config.Metadata, "metadata", "m", map[string]string{},
		"Request metadata as stringified JSON.")
	rootCmd.PersistentFlags().StringVarP(&config.MetadataPath, "metadata-file", "M", "",
		"File path for call metadata JSON file. Examples: /home/user/metadata.json or ./metadata.json.")
	rootCmd.PersistentFlags().DurationVar(&config.SI, "stream-interval", 0,
		"Timeout interval for stream requests between individual message sends.")
	rootCmd.PersistentFlags().DurationVar(&config.StreamCallDuration, "stream-call-duration", 0,
		"Duration after which client will close the stream in each streaming call.")
	rootCmd.PersistentFlags().UintVar(&config.StreamCallCount, "stream-call-count", 0,
		"Count of messages sent, after which client will close the stream in each streaming call.")
	rootCmd.PersistentFlags().BoolVar(&config.StreamDynamicMessages, "stream-dynamic-messages", false,
		"In streaming calls, regenerate and apply call template data on every message send.")
	rootCmd.PersistentFlags().StringToStringVar(&config.ReflectMetadata, "reflect-metadata", map[string]string{},
		"Reflect metadata as stringified JSON used only for reflection request.")

	// output
	rootCmd.PersistentFlags().StringVarP(&config.Output, "output", "o", "",
		"Output path. If none provided stdout is used.")
	rootCmd.PersistentFlags().StringVarP(&config.Format, "format", "O", "summary",
		"Output format. One of: summary, csv, json, pretty, html, influx-summary, influx-details. Default is summary.")
	rootCmd.PersistentFlags().UintVar(&config.SkipFirst, "skip-first", 0, "Skip the first X requests when doing the results tally.")
	rootCmd.PersistentFlags().BoolVar(&config.CountErrors, "count-errors", false, "Count erroneous (non-OK) resoponses in stats calculations.")

	// connection
	rootCmd.PersistentFlags().UintVar(&config.Connections, "connections", 1,
		"Number of connections to use. Concurrency is distributed evenly among all the connections. Default is 1.")
	rootCmd.PersistentFlags().DurationVar(&config.DialTimeout, "connect-timeout", 10*time.Second,
		"Connection timeout for the initial connection dial. Default is 10s.")
	rootCmd.PersistentFlags().DurationVar(&config.KeepaliveTime, "keepalive", 0,
		"Keepalive time duration. Only used if present and above 0.")
	rootCmd.PersistentFlags().BoolVarP(&config.EnableCompression, "enable-compression", "e", false,
		"Enable Gzip compression on requests.")
	rootCmd.PersistentFlags().StringVar(&config.LBStrategy, "lb-strategy", "", "Client load balancing strategy.")

	// meta
	rootCmd.PersistentFlags().StringVar(&config.Name, "name", "", "User specified name for the test.")
	rootCmd.PersistentFlags().StringToStringVar(&config.Tags, "tags", map[string]string{}, "JSON representation of user-defined string tags.")
	rootCmd.PersistentFlags().UintVar(&config.CPUs, "cpus", uint(nCPUs), "Number of cpu cores to use.")
	rootCmd.PersistentFlags().StringVar(&config.Debug, "debug", "", "The path to debug log file.")

	viper.BindPFlags(rootCmd.PersistentFlags())

	rootCmd.Execute()
}

func mergeFlags(cmd *cobra.Command, v *viper.Viper) {
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		// Apply the viper config value to the flag when the flag is not set and viper has a value
		if !f.Changed && v.IsSet(f.Name) {
			val := v.Get(f.Name)
			cmd.Flags().Set(f.Name, fmt.Sprintf("%v", val))
		}
	})
}

func handleError(err error) {
	if err != nil {
		if errString := err.Error(); errString != "" {
			fmt.Fprintln(os.Stderr, errString)
		}
		os.Exit(1)
	}
}

// func createConfigFromArgs(cfg *runner.Config) error {
// 	if cfg == nil {
// 		return errors.New("config cannot be nil")
// 	}

// 	iPaths := []string{}
// 	pathsTrimmed := strings.TrimSpace(*paths)
// 	if pathsTrimmed != "" {
// 		iPaths = strings.Split(pathsTrimmed, ",")
// 	}

// 	var binaryData []byte
// 	if *binData {
// 		b, err := ioutil.ReadAll(os.Stdin)
// 		if err != nil {
// 			return err
// 		}

// 		binaryData = b
// 	}

// 	var metadata map[string]string
// 	*md = strings.TrimSpace(*md)
// 	if *md != "" {
// 		if err := json.Unmarshal([]byte(*md), &metadata); err != nil {
// 			return fmt.Errorf("Error unmarshaling metadata '%v': %v", *md, err.Error())
// 		}
// 	}

// 	var dataObj interface{}
// 	if *data != "@" && strings.TrimSpace(*data) != "" {
// 		if err := json.Unmarshal([]byte(*data), &dataObj); err != nil {
// 			return fmt.Errorf("Error unmarshaling data '%v': %v", *data, err.Error())
// 		}
// 	}

// 	var tagsMap map[string]string
// 	*tags = strings.TrimSpace(*tags)
// 	if *tags != "" {
// 		if err := json.Unmarshal([]byte(*tags), &tagsMap); err != nil {
// 			return fmt.Errorf("Error unmarshaling tags '%v': %v", *tags, err.Error())
// 		}
// 	}

// 	var rmdMap map[string]string
// 	*rmd = strings.TrimSpace(*rmd)
// 	if *rmd != "" {
// 		if err := json.Unmarshal([]byte(*rmd), &rmdMap); err != nil {
// 			return fmt.Errorf("Error unmarshaling reflection metadata '%v': %v", *rmd, err.Error())
// 		}
// 	}

// 	cfg.Host = *host
// 	cfg.Proto = *proto
// 	cfg.Protoset = *protoset
// 	cfg.Call = *call
// 	cfg.RootCert = *cacert
// 	cfg.Cert = *cert
// 	cfg.Key = *key
// 	cfg.SkipTLSVerify = *skipVerify
// 	cfg.SkipFirst = *skipFirst
// 	cfg.Insecure = *insecure
// 	cfg.Authority = *authority
// 	cfg.CName = *cname
// 	cfg.N = *n
// 	cfg.C = *c
// 	cfg.RPS = *rps
// 	cfg.Z = runner.Duration(*z)
// 	cfg.X = runner.Duration(*x)
// 	cfg.Timeout = runner.Duration(*t)
// 	cfg.ZStop = *zstop
// 	cfg.Data = dataObj
// 	cfg.DataPath = *dataPath
// 	cfg.BinData = binaryData
// 	cfg.BinDataPath = *binPath
// 	cfg.Metadata = metadata
// 	cfg.MetadataPath = *mdPath
// 	cfg.SI = runner.Duration(*si)
// 	cfg.StreamCallDuration = runner.Duration(*scd)
// 	cfg.StreamCallCount = *scc
// 	cfg.StreamDynamicMessages = *sdm
// 	cfg.Output = *output
// 	cfg.Format = *format
// 	cfg.ImportPaths = iPaths
// 	cfg.Connections = *conns
// 	cfg.DialTimeout = runner.Duration(*ct)
// 	cfg.KeepaliveTime = runner.Duration(*kt)
// 	cfg.CPUs = *cpus
// 	cfg.Name = *name
// 	cfg.Tags = tagsMap
// 	cfg.ReflectMetadata = rmdMap
// 	cfg.Debug = *debug
// 	cfg.EnableCompression = *enableCompression
// 	cfg.LoadSchedule = *schedule
// 	cfg.LoadStart = *loadStart
// 	cfg.LoadStep = *loadStep
// 	cfg.LoadEnd = *loadEnd
// 	cfg.LoadStepDuration = runner.Duration(*loadStepDuration)
// 	cfg.LoadMaxDuration = runner.Duration(*loadMaxDuration)
// 	cfg.Async = *async
// 	cfg.CSchedule = *cschdule
// 	cfg.CStart = *cStart
// 	cfg.CStep = *cstep
// 	cfg.CEnd = *cEnd
// 	cfg.CStepDuration = runner.Duration(*cStepDuration)
// 	cfg.CMaxDuration = runner.Duration(*cMaxDuration)
// 	cfg.CountErrors = *countErrors
// 	cfg.LBStrategy = *lbStrategy

// 	return nil
// }

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
