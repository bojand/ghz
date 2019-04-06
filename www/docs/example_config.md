---
id: example_config
title: Configuration Files
---

All the call options can be specified in JSON or TOML config files and used as input via the `-config` option. 

An example JSON config file:

```json
{
    "proto": "/path/to/greeter.proto",
    "call": "helloworld.Greeter.SayHello",
    "total": 2000,
    "concurrency": 50,
    "data": {
        "name": "Joe"
    },
    "metadata": {
        "foo": "bar",
        "trace_id": "{{.RequestNumber}}",
        "timestamp": "{{.TimestampUnix}}"
    },
    "import-paths": [
        "/path/to/protos"
    ],
    "max-duration": "10s",
    "host": "0.0.0.0:50051"
}
```

An example TOML config file:

```toml
"max-duration" = "7s"
total = 5000
concurrency = 50
proto = "../../testdata/greeter.proto"
call = "helloworld.Greeter.SayHello"
host = "0.0.0.0:50051"
insecure = true
output = "pretty.json"
format = "pretty"

[data]
name = "Bob {{.TimestampUnix}}"

[metadata]
rn = "{{.RequestNumber}}"
```
