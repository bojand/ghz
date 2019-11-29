---
id: examples
title: Examples
---

A simple insecure unary call:

```sh
ghz --insecure \
  --proto ./greeter.proto \
  --call helloworld.Greeter.SayHello \
  -d '{"name":"Joe"}' \
  0.0.0.0:50051
```

Or same test using [server reflection](https://github.com/grpc/grpc/blob/master/doc/server-reflection.md) (just omit `-proto` option):

```sh
ghz --insecure \
  --call helloworld.Greeter.SayHello \
  -d '{"name":"Joe"}' \
  0.0.0.0:50051
```

A simple unary call with metadata using template actions:

```sh
ghz --insecure \
  --proto ./greeter.proto \
  --call helloworld.Greeter.SayHello \
  -d '{"name":"Joe"}' \
  -m '{"trace_id":"{{.RequestNumber}}", "timestamp":"{{.TimestampUnix}}"}' \
  0.0.0.0:50051
```

Using binary data file (see [writing a message](https://developers.google.com/protocol-buffers/docs/gotutorial#writing-a-message)):

```sh
ghz --proto ./greeter.proto \
  --call helloworld.Greeter.SayHello \
  -B ./hello_request_data.bin \
  0.0.0.0:50051
```

Or using binary from stdin:

```sh
ghz --proto ./greeter.proto \
  --call helloworld.Greeter.SayHello \
  0.0.0.0:50051 < ./hello_request_data.bin
```

Round-robin of messages for unary call:

```sh
ghz --insecure \
  --proto ./greeter.proto \
  --call helloworld.Greeter.SayHello \
  -d '[{"name":"Joe"},{"name":"Bob"}]' \
  0.0.0.0:50051
```

Custom number of requests and concurrency:

```sh
ghz --proto ./greeter.proto \
  --call helloworld.Greeter.SayHello \
  -d '{"name":"Joe"}' \
  -n 2000 \
  -c 20 \
  0.0.0.0:50051
```

Using custom number of connections:

```sh
ghz --proto ./greeter.proto \
  --call helloworld.Greeter.SayHello \
  -d '{"name":"Joe"}' \
  -n 2000 \
  -c 20 \
  --connections=10 \
  0.0.0.0:50051
```

`10` connections will be shared among `20` goroutine workers. Each pair of `2` goroutines will share a single connection.

Client streaming data can be sent as an array, each element representing a single message:

```sh
ghz --proto ./greeter.proto \
  --call helloworld.Greeter.SayHelloCS \
  -d '[{"name":"Joe"},{"name":"Kate"},{"name":"Sara"}]' \
  0.0.0.0:50051
```

If a single object is given for data it is sent as every message.

We can also use `.protoset` files which can bundle multiple protocol buffer files into one binary file.

Create a protoset

```sh
protoc --proto_path=. --descriptor_set_out=bundle.protoset *.proto
```

And then use it as input to `ghz` with `-protoset` option:

```sh
ghz --protoset ./bundle.protoset \
  --call helloworld.Greeter.SayHello \
  -d '{"name":"Bob"}' \
  -n 1000 -c 10 \
  0.0.0.0:50051
```

Note that only one of `-proto` or `-protoset` options will be used. `-proto` takes precedence.

Alternatively `ghz` can be used with [Prototool](https://github.com/uber/prototool) using the [`descriptor-set`](https://github.com/uber/prototool/tree/dev/docs#prototool-descriptor-set) command:

```
ghz --protoset $(prototool descriptor-set --include-imports --tmp) ...
```

Finally we can specify all settings, including the target host, conveniently in a JSON or TOML config file.

```sh
ghz --config ./config.json
```

Config file settings can be combined with command line arguments. CLI options overwrite config file options.

```sh
ghz --config ./config.json -c 20 -n 1000
```
