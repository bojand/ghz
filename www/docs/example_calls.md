---
id: example_call
title: Examples
---

A simple insecure unary call:

```sh
ghz -insecure \
  -proto ./greeter.proto \
  -call helloworld.Greeter.SayHello \
  -d '{"name":"Joe"}' \
  0.0.0.0:50051
```

A simple unary call with metadata using template actions:

```sh
ghz -insecure \
  -proto ./greeter.proto \
  -call helloworld.Greeter.SayHello \
  -d '{"name":"Joe"}' \
  -m '{"trace_id":"{{.RequestNumber}}", "timestamp":"{{.TimestampUnix}}"}' \
  0.0.0.0:50051
```

Using binary data file (see [writing a message](https://developers.google.com/protocol-buffers/docs/gotutorial#writing-a-message)):

```sh
ghz -proto ./testdata/greeter.proto \
  -call helloworld.Greeter.SayHello \
  -B ./hello_request_data.bin \
  0.0.0.0:50051
```

Or using binary from stdin:

```sh
ghz -proto ./greeter.proto \
  -call helloworld.Greeter.SayHello \
  0.0.0.0:50051 < ./hello_request_data.bin
```

Custom number of requests and concurrency:

```sh
ghz -proto ./greeter.proto \
  -call helloworld.Greeter.SayHello \
  -d '{"name":"Joe"}' \
  -n 2000 \
  -c 20 \
  0.0.0.0:50051
```

Client streaming data can be sent as an array, each element representing a single message:

```sh
ghz -proto ./greeter.proto \
  -call helloworld.Greeter.SayHelloCS \
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
ghz -protoset ./bundle.protoset \
  -call helloworld.Greeter.SayHello \
  -d '{"name":"Bob"}' \
  -n 1000 -c 10 \
  0.0.0.0:50051
```

Note that only one of `-proto` or `-protoset` options will be used. `-proto` takes precedence.

Finally we can specify all settings, including the target host, conviniently in a JSON or TOML config file.

```sh
ghz -config ./config.json
```
