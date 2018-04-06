# grpcannon

[![build status](https://img.shields.io/travis/bojand/grpcannon/master.svg?style=flat-square)](https://travis-ci.org/bojand/grpcannon)

Simple [gRPC](http://grpc.io/) benchmarking and load testing tool inspired by [hey](https://github.com/rakyll/hey/) and [grpcurl](https://github.com/fullstorydev/grpcurl).

## Demo

![demo](grpcannon.gif)

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
  -M  Path for call data JSON file. For example, /home/user/metadata.json or ./metadata.json.

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
/grpcannon -proto ./greeter.proto -call helloworld.Greeter.SayHello -d '{"name":"Joe"}' 0.0.0.0:50051
```

Custom number of requests and concurrency:

```sh
/grpcannon -proto ./greeter.proto -call helloworld.Greeter.SayHello -d '{"name":"Joe"}' -n 2000 -c 20 0.0.0.0:50051
```

Client streaming data can be sent as an array, each element representing a single message:

```sh
/grpcannon -proto ./greeter.proto -call helloworld.Greeter.SayHelloCS -d '[{"name":"Joe"},{"name":"Kate"},{"name":"Sara"}]' 0.0.0.0:50051
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

## License

Apache-2.0
