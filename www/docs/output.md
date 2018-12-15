---
id: output
title: Output Formats
---

### Summary

Sample standard output of summary of the results:

```
Summary:
  Count:	2000
  Total:	345.52 ms
  Slowest:	15.41 ms
  Fastest:	0.66 ms
  Average:	6.83 ms
  Requests/sec:	5788.35

Response time histogram:
  0.664 [1]	    |
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

HTML output can be generated using `html` as format in the `-O` option. See [sample output](http://bojand.github.io/sample.html).

### JSON

Using `-O json` outputs JSON data, and `-O pretty` outputs JSON in pretty format.

### InfluxDB Line Protocol

Using `-O influx-summary` outputs the summary data as [InfluxDB Line Protocol](https://docs.influxdata.com/influxdb/v1.6/concepts/glossary/#line-protocol). Sample output:

```
ghz_run,proto="/testdata/greeter.proto",call="helloworld.Greeter.SayHello",host="0.0.0.0:50051",n=1000,c=50,qps=0,z=0,timeout=20,dial_timeout=10,keepalive=0,data="{\"name\":\"{{.InputName}}\"}",metadata="{\"rn\":\"{{.RequestNumber}}\"}",errors=74,has_errors=true count=1000,total=50000556,average=1771308,fastest=248603,slowest=7241944,rps=19999.78,median=1715940,p95=4354194,errors=74 128802790
```

Use `-O influx-details` to get the individual details for each request:

```
ghz_detail,proto="/testdata/greeter.proto",call="helloworld.Greeter.SayHello",host="0.0.0.0:50051",n=1000,c=50,qps=0,z=0,timeout=20,dial_timeout=10,keepalive=0,data="{\"name\":\"{{.InputName}}\"}",metadata="{\"rn\":\"{{.RequestNumber}}\"}",hasError=false latency=5157328,error=,status=OK 681023506
ghz_detail,proto="/testdata/greeter.proto",call="helloworld.Greeter.SayHello",host="0.0.0.0:50051",n=1000,c=50,qps=0,z=0,timeout=20,dial_timeout=10,keepalive=0,data="{\"name\":\"{{.InputName}}\"}",metadata="{\"rn\":\"{{.RequestNumber}}\"}",hasError=false latency=4990499,error=,status=OK 681029613
```
