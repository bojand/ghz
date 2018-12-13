---
id: example_config
title: Configuration Files
---

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
