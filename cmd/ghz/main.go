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
	"github.com/bojand/ghz/config"
	"github.com/jinzhu/configor"
)

var (
	// set by goreleaser with -ldflags="-X main.version=..."
	version = "dev"

	host     = flag.String("host", "", `The target host`)
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

	localConfigName = "ghz.json"
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

	var cfg *config.Config

	if cfgPath != "" {
		err := configor.Load(cfg, cfgPath)
		if err != nil {
			errAndExit(err.Error())
		}
	} else if _, err := os.Stat(localConfigName); err == nil {
		err := configor.Load(cfg, localConfigName)
		if err != nil {
			errAndExit(err.Error())
		}
	} else {
		if flag.NArg() < 1 {
			usageAndExit("")
		}

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
				errAndExit(err.Error())
			}

			binaryData = b
		}

		var metadata *map[string]string
		*md = strings.TrimSpace(*md)
		if *md != "" {
			if err := json.Unmarshal([]byte(*md), metadata); err != nil {
				errAndExit(err.Error())
			}
		}

		var dataObj interface{}
		if strings.TrimSpace(*data) != "@" && strings.TrimSpace(*data) != "" {
			if err := json.Unmarshal([]byte(*data), dataObj); err != nil {
				errAndExit(err.Error())
			}
		}

		cfg := &config.Config{
			Host:          host,
			Proto:         *proto,
			Protoset:      *protoset,
			Call:          *call,
			Cert:          *cert,
			CName:         *cname,
			N:             *n,
			C:             *c,
			QPS:           *q,
			Z:             *z,
			X:             *x,
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

		if cfg.X > 0 {
			cfg.Z = cfg.X
		} else if cfg.Z > 0 {
			cfg.N = math.MaxInt32
		}
	}

	options := make([]ghz.Option, 0, 10)

	options = append(options,
		ghz.WithProtoFile(*proto, cfg.ImportPaths),
		ghz.WithProtoset(*protoset),
		ghz.WithCertificate(*cert, *cname),
		ghz.WithInsecure(*insecure),
		ghz.WithConcurrency(*c),
		ghz.WithTotalRequests(*n),
		ghz.WithQPS(*q),
		ghz.WithTimeout(time.Duration(*t)*time.Second),
		ghz.WithRunDuration(*z),
		ghz.WithDataFromJSON(*data),
	)

	if strings.TrimSpace(*data) == "@" {
		options = append(options, ghz.WithDataFromReader(os.Stdin))
	} else if strings.TrimSpace(cfg.DataPath) != "" {
		options = append(options, ghz.WithDataFromFile(cfg.DataPath))
	} else {
		options = append(options, ghz.WithDataFromFile(cfg.DataPath))
	}

	ghz.Run(*call, *host, options...)
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

func customParse() (bool, string, []string) {

	givenArgs := os.Args[1:]
	nArgs := len(givenArgs)
	args := make([]string, nArgs)

	v := false
	var cfgPath string
	for i, f := range givenArgs {
		if f == "-v" {
			v = true
		} else if f == "-config" && nArgs > i+1 {
			cfgPath = givenArgs[i+1]
		} else {
			args = append(args, f)
		}
	}

	return v, cfgPath, args
}

func printArgs() {
	fmt.Println("proto:", *proto)
	fmt.Println("protoset:", *protoset)
	fmt.Println("call:", *call)
	fmt.Println("cert:", *cert)
	fmt.Println("cname:", *cname)
	fmt.Printf("n: %+v\n", *n)
	fmt.Println("c:", *c)
	fmt.Println("q:", *q)
	fmt.Println("z:", *z)
	fmt.Println("x:", *x)
	fmt.Println("t:", *t)
	fmt.Println("data:", *data)
	fmt.Println("dataPath:", *dataPath)
	fmt.Println("binData:", *binData)
	fmt.Println("binPath:", *binPath)
	fmt.Println("md:", *md)
	fmt.Println("mdPath:", *mdPath)
	fmt.Println("output:", *output)
	fmt.Println("format:", *format)
	fmt.Println("host:", *host)
	fmt.Println("ct:", *ct)
	fmt.Println("kt:", *kt)
	fmt.Println("cpus:", *cpus)
	fmt.Println("insecure:", *insecure)
	fmt.Println("name:", *name)
	fmt.Println("paths:", *paths)
}
