# grpcannon

[![build status](https://img.shields.io/travis/bojand/grpcannon/master.svg?style=flat-square)](https://travis-ci.org/bojand/grpcannon)

Simple [gRPC](http://grpc.io/) benchmarking and load testing tool inspired by [hey](https://github.com/rakyll/hey/) and [grpcurl](https://github.com/fullstorydev/grpcurl).

## Demo

![demo](grpcannon.gif)

## Install

Download a prebuild executable binary from the [releases page](https://github.com/bojand/grpcannon/releases).

## Usage

```
Usage: grpcannon [options...] <host>
Options:
  -proto	The protocol buffer file.
  -call		A fully-qualified method name in 'service/method' or 'service.method' format.
  -cert		The file containing the CA root cert file.

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

  -cpus		Number of used cpu cores. (default for current machine is 8 cores)

  -v  Print the version.
```

Alternatively all settings can be set via `grpcannon.json` file if present in the same path as the `grpcannon` executable.

## Examples

A simple unary call:

```sh
grpcannon -proto ./greeter.proto -call helloworld.Greeter.SayHello -d '{"name":"Joe"}' 0.0.0.0:50051
```

Custom number of requests and concurrency:

```sh
grpcannon -proto ./greeter.proto -call helloworld.Greeter.SayHello -d '{"name":"Joe"}' -n 2000 -c 20 0.0.0.0:50051
```

Client streaming data can be sent as an array, each element representing a single message:

```sh
grpcannon -proto ./greeter.proto -call helloworld.Greeter.SayHelloCS -d '[{"name":"Joe"},{"name":"Kate"},{"name":"Sara"}]' 0.0.0.0:50051
```

If a single object is given for data it is sent as every message.

Example `grpcannon.json`

```json
{
    "proto": "/path/to/greeter.proto",
    "call": "helloworld.Greeter.SayHello",
    "n": 2000,
    "c": 50,
    "d": {
        "name": "Joe"
    },
    "i": [
        "/path/to/protos"
    ]
}
```

## Output

Sample standard output of summary of the results:

```
Summary:
  Count:	2000
  Total:	345.52 ms
  Slowest:	15.41 ms
  Fastest:	0.66 ms
  Average:	6.83 ms
  Requests/sec:	5788347.22

Response time histogram:
  0.664 [1]	|
  2.138 [36]	|∎
  3.613 [14]	|
  5.087 [65]	|∎∎
  6.561 [1305]	|∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎
  8.035 [274]	|∎∎∎∎∎∎∎∎
  9.509 [66]	|∎∎
  10.983 [0]	|
  12.458 [59]	|∎∎
  13.932 [130]	|∎∎∎∎
  15.406 [50]	|∎∎

Latency distribution:
  10% in 5.18 ms
  25% in 5.51 ms
  50% in 6.10 ms
  75% in 6.72 ms
  90% in 12.19 ms
  95% in 13.26 ms
  99% in 14.73 ms
Status code distribution:
  [OK]	2000 responses
```

Alternatively with `-O csv` flag we can get detailed listing in csv format:

```sh
duration (ms),status,error
1.43,OK,
0.39,OK,
0.36,OK,
0.50,OK,
0.36,OK,
0.40,OK,
0.37,OK,
0.34,OK,
0.35,OK,
0.32,OK,
...
```

## License

Apache-2.0
