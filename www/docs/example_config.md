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
    "n": 2000,
    "c": 50,
    "d": {
        "name": "Joe"
    },
    "m": {
        "foo": "bar",
        "trace_id": "{{.RequestNumber}}",
        "timestamp": "{{.TimestampUnix}}"
    },
    "i": [
        "/path/to/protos"
    ],
    "x": "10s",
    "host": "0.0.0.0:50051"
}
```

An example TOML config file:

```toml
x = "7s"
n = 5000
c = 50
proto = "../../testdata/greeter.proto"
call = "helloworld.Greeter.SayHello"
host = "0.0.0.0:50051"
insecure = true
o = "pretty.json"
O = "pretty"

[d]
name = "Bob {{.TimestampUnix}}"

[m]
rn = "{{.RequestNumber}}"
```
