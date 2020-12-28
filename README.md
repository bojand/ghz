<div align="center">
	<br>
	<img src="green_fwd2.svg" alt="Logo" width="100">
	<br>
</div>

# ghz

[![Release](https://img.shields.io/github/release/bojand/ghz.svg?style=flat-square)](https://github.com/bojand/ghz/releases/latest)
![Build Status](https://github.com/bojand/ghz/workflows/build/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/bojand/ghz?style=flat-square)](https://goreportcard.com/report/github.com/bojand/ghz)
[![License](https://img.shields.io/github/license/bojand/ghz.svg?style=flat-square)](https://raw.githubusercontent.com/bojand/ghz/master/LICENSE)
[![Donate](https://img.shields.io/badge/Donate-PayPal-green.svg?style=flat-square)](https://www.paypal.me/bojandj)
[![Buy me a coffee](https://img.shields.io/badge/buy%20me-a%20coffee-orange.svg?style=flat-square)](https://www.buymeacoffee.com/bojand)

[gRPC](http://grpc.io/) benchmarking and load testing tool.

## Documentation

All documentation at [ghz.sh](https://ghz.sh).

## Install

### Download

1. Download a prebuilt executable binary for your operating system from the [GitHub releases page](https://github.com/bojand/ghz/releases).
2. Unzip the archive and place the executable binary wherever you would like to run it from. Additionally consider adding the location directory in the `PATH` variable if you would like the `ghz` command to be available everywhere.

### Homebrew

```sh
brew install ghz
```

### Compile

**Clone**

```sh
git clone https://github.com/bojand/ghz
```

**Build using make**

```sh
make build
```

**Build using go**

```sh
cd cmd/ghz
go build .
```

## Usage

```
usage: ghz [<flags>] [<host>]

Flags:
  -h, --help                     Show context-sensitive help (also try --help-long and --help-man).
      --config=                  Path to the JSON or TOML config file that specifies all the test run settings.
      --proto=                   The Protocol Buffer .proto file.
      --protoset=                The compiled protoset file. Alternative to proto. -proto takes precedence.
      --call=                    A fully-qualified method name in 'package.Service/method' or 'package.Service.Method' format.
  -i, --import-paths=            Comma separated list of proto import paths. The current working directory and the directory of the protocol buffer file are automatically added to the import list.
      --cacert=                  File containing trusted root certificates for verifying the server.
      --cert=                    File containing client certificate (public key), to present to the server. Must also provide -key option.
      --key=                     File containing client private key, to present to the server. Must also provide -cert option.
      --cname=                   Server name override when validating TLS certificate - useful for self signed certs.
      --skipTLS                  Skip TLS client verification of the server's certificate chain and host name.
      --insecure                 Use plaintext and insecure connection.
      --authority=               Value to be used as the :authority pseudo-header. Only works if -insecure is used.
      --async                    Make requests asynchronous as soon as possible. Does not wait for request to finish before sending next one.
  -r, --rps=0                    Requests per second (RPS) rate limit for constant load schedule. Default is no rate limit.
      --load-schedule="const"    Specifies the load schedule. Options are const, step, or line. Default is const.
      --load-start=0             Specifies the RPS load start value for step or line schedules.
      --load-step=0              Specifies the load step value or slope value.
      --load-end=0               Specifies the load end value for step or line load schedules.
      --load-step-duration=0     Specifies the load step duration value for step load schedule.
      --load-max-duration=0      Specifies the max load duration value for step or line load schedule.
  -c, --concurrency=50           Number of request workers to run concurrently for const concurrency schedule. Default is 50.
      --concurrency-schedule="const"
                                 Concurrency change schedule. Options are const, step, or line. Default is const.
      --concurrency-start=0      Concurrency start value for step and line concurrency schedules.
      --concurrency-end=0        Concurrency end value for step and line concurrency schedules.
      --concurrency-step=1       Concurrency step / slope value for step and line concurrency schedules.
      --concurrency-step-duration=0
                                 Specifies the concurrency step duration value for step concurrency schedule.
      --concurrency-max-duration=0
                                 Specifies the max concurrency adjustment duration value for step or line concurrency schedule.
  -n, --total=200                Number of requests to run. Default is 200.
  -t, --timeout=20s              Timeout for each request. Default is 20s, use 0 for infinite.
  -z, --duration=0               Duration of application to send requests. When duration is reached, application stops and exits. If duration is specified, n is ignored. Examples: -z 10s -z 3m.
  -x, --max-duration=0           Maximum duration of application to send requests with n setting respected. If duration is reached before n requests are completed, application stops and exits. Examples: -x 10s -x 3m.
      --duration-stop="close"    Specifies how duration stop is reported. Options are close, wait or ignore. Default is close.
  -d, --data=                    The call data as stringified JSON. If the value is '@' then the request contents are read from stdin.
  -D, --data-file=               File path for call data JSON file. Examples: /home/user/file.json or ./file.json.
  -b, --binary                   The call data comes as serialized binary message or multiple count-prefixed messages read from stdin.
  -B, --binary-file=             File path for the call data as serialized binary message or multiple count-prefixed messages.
  -m, --metadata=                Request metadata as stringified JSON.
  -M, --metadata-file=           File path for call metadata JSON file. Examples: /home/user/metadata.json or ./metadata.json.
      --stream-interval=0        Interval for stream requests between message sends.
      --stream-call-duration=0   Duration after which client will close the stream in each streaming call.
      --stream-call-count=0      Count of messages sent, after which client will close the stream in each streaming call.
      --stream-dynamic-messages  In streaming calls, regenerate and apply call template data on every message send.
      --reflect-metadata=        Reflect metadata as stringified JSON used only for reflection request.
  -o, --output=                  Output path. If none provided stdout is used.
  -O, --format=                  Output format. One of: summary, csv, json, pretty, html, influx-summary, influx-details. Default is summary.
      --skipFirst=0              Skip the first X requests when doing the results tally.
      --count-errors             Count erroneous (non-OK) resoponses in stats calculations.
      --connections=1            Number of connections to use. Concurrency is distributed evenly among all the connections. Default is 1.
      --connect-timeout=10s      Connection timeout for the initial connection dial. Default is 10s.
      --keepalive=0              Keepalive time duration. Only used if present and above 0.
      --name=                    User specified name for the test.
      --tags=                    JSON representation of user-defined string tags.
      --cpus=12                  Number of cpu cores to use.
      --debug=                   The path to debug log file.
  -e, --enable-compression       Enable Gzip compression on requests.
  -v, --version                  Show application version.

Args:
  [<host>]  Host and port to test.
```

## Go Package

```go
report, err := runner.Run(
    "helloworld.Greeter.SayHello",
    "localhost:50051",
    runner.WithProtoFile("greeter.proto", []string{}),
    runner.WithDataFromFile("data.json"),
    runner.WithInsecure(true),
)

if err != nil {
    fmt.Println(err.Error())
    os.Exit(1)
}

printer := printer.ReportPrinter{
    Out:    os.Stdout,
    Report: report,
}

printer.Print("pretty")
```

## Development

Golang 1.11+ is required.

```
make # run all linters, tests, and produce code coverage
make build # build the binaries
make lint # run all linters
make test # run tests
make cover # run tests and produce code coverage

V=1 make # more verbosity
OPEN_COVERAGE=1 make cover # open code coverage.html after running
```

## Credit

Icon made by <a href="http://www.freepik.com" title="Freepik">Freepik</a> from <a href="https://www.flaticon.com/" title="Flaticon">www.flaticon.com</a> is licensed by <a href="http://creativecommons.org/licenses/by/3.0/" title="Creative Commons BY 3.0" target="_blank">CC 3.0 BY</a>

## License

Apache-2.0
