---
id: output
title: Output
---

### Summary

Sample standard output of summary of the results:

```
Summary:
  Count:		200
  Total:		181.57 ms
  Slowest:		69.60 ms
  Fastest:		26.09 ms
  Average:		32.01 ms
  Requests/sec:	1101.53

Response time histogram:
  26.093 [1]	|∎
  30.444 [52]	|∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎
  34.794 [78]	|∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎
  39.145 [40]	|∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎
  43.495 [1]	|∎
  47.846 [0]	|
  52.196 [2]	|∎
  56.547 [5]	|∎∎∎
  60.897 [3]	|∎∎
  65.248 [2]	|∎
  69.598 [2]	|∎

Latency distribution:
  10% in 28.48 ms
  25% in 30.08 ms
  50% in 33.23 ms
  75% in 35.43 ms
  90% in 38.89 ms
  95% in 55.45 ms
  99% in 69.60 ms

Status code distribution:
  [Unavailable]        3 responses
  [PermissionDenied]   3 responses
  [OK]                 186 responses
  [Internal]           8 responses
Error distribution:
  [8]	rpc error: code = Internal desc = Internal error.
  [3]	rpc error: code = PermissionDenied desc = Permission denied.
  [3]	rpc error: code = Unavailable desc = Service unavialable.
```

Explanation of the summary:

- `count` - The total number of completed requests including successful and failed requests.
- `total` - The total time spent running the test within `ghz` from start to finish. This is a single measurement from start of the test run to the completion of the final request of the test run.
- `slowest` - The measurement of the slowest request.
- `fastest` - The measurement of the fastest request.
- `average` - The mathematical average computed by taking the _sum_ of the _individual_ response times of _all_ requests and dividing it by the total number of requests.
- `requests/sec` - Theoretical computed RPS computed by taking the total number of requests (successful and failed) and dividing it by the total duration of the test. That is: `count` / `total`.

With regard to measurement, we use [WithStatsHandler](https://godoc.org/google.golang.org/grpc#WithStatsHandler) option to capture call metrics. Specifically we only capture the [End](https://godoc.org/google.golang.org/grpc/stats#End) event which contains stats when an RPC ends. This should include the download of the payload and deserializing of the data.

### CSV

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

### HTML

HTML output can be generated using `html` as format in the `-O` option. [Sample HTML output](/sample.html).

### JSON

Using `-O json` outputs JSON data, and `-O pretty` outputs JSON in pretty format. [Sample pretty JSON output](/pretty.json).

### Prometheus

Using `-O prometheus` outputs the summary data as [Prometheus text exposition format
](https://prometheus.io/docs/instrumenting/exposition_formats/). [Sample Prometheus output](/prometheus.txt).

### InfluxDB Line Protocol

Using `-O influx-summary` outputs the summary data as [InfluxDB Line Protocol](https://docs.influxdata.com/influxdb/v1.6/concepts/glossary/#line-protocol). Sample output:

```
ghz_run,name="Greeter\ SayHello",proto="./greeter.proto",call="helloworld.Greeter.SayHello",host="0.0.0.0:50051",n=200,c=50,rps=0,z=0,timeout=20,dial_timeout=10,keepalive=0,data="{\"name\":\"Bob\ Smith\"}",metadata="",tags="{\"created\ by\":\"Joe\ Developer\"\,\"env\":\"staging\"}",errors=0,has_errors=false count=200,total=214737065,average=37806598,fastest=25759157,slowest=77504712,rps=931.37,median=36947515,p95=47421426,errors=0 1548107303068421000
```

Use `-O influx-details` to get the individual details for each request:

```
ghz_detail,name="Greeter\ SayHello",proto="./greeter.proto",call="helloworld.Greeter.SayHello",host="0.0.0.0:50051",n=200,c=50,rps=0,z=0,timeout=20,dial_timeout=10,keepalive=0,data="{\"name\":\"Bob\ Smith\"}",metadata="",tags="{\"created\ by\":\"Joe\ Developer\"\,\"env\":\"staging\"}",hasError=false latency=79044469,error="",status="OK" 1548107176979991000
ghz_detail,name="Greeter\ SayHello",proto="./greeter.proto",call="helloworld.Greeter.SayHello",host="0.0.0.0:50051",n=200,c=50,rps=0,z=0,timeout=20,dial_timeout=10,keepalive=0,data="{\"name\":\"Bob\ Smith\"}",metadata="",tags="{\"created\ by\":\"Joe\ Developer\"\,\"env\":\"staging\"}",hasError=false latency=43011582,error="",status="OK" 1548107177023123000
```
