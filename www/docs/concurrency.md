---
id: concurrency
title: Concurrency Options
---

This is a walkthrough of different concurrency options available to control the number of concurrent workers that `ghz` utilizes to make requests to the server. All examples are done using a simple unary gRPC call.

Many of these options are similar to the load control options, but independently control the concurrent workers utilized.

## Step Up Concurrency

```
./dist/ghz --insecure --async --proto /protos/helloworld.proto \
  --call helloworld.Greeter/SayHello \
  -n 10000 --rps 200 \
  --concurrency-schedule=step --concurrency-start=5 --concurrency-step=5 --concurrency-end=50 --concurrency-step-duration=5s \
  -d '{"name":"{{.WorkerID}}"}' 0.0.0.0:50051

Summary:
  Count:	10000
  Total:	50.05 s
  Slowest:	52.04 ms
  Fastest:	50.19 ms
  Average:	50.59 ms
  Requests/sec:	199.79

Response time histogram:
  50.187 [1]	|
  50.373 [1786]	|∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎
  50.558 [3032]	|∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎
  50.743 [2822]	|∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎
  50.929 [1536]	|∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎
  51.114 [562]	|∎∎∎∎∎∎∎
  51.299 [194]	|∎∎∎
  51.485 [42]	|∎
  51.670 [15]	|
  51.855 [6]	|
  52.041 [4]	|

Latency distribution:
  10 % in 50.33 ms
  25 % in 50.42 ms
  50 % in 50.57 ms
  75 % in 50.73 ms
  90 % in 50.89 ms
  95 % in 51.01 ms
  99 % in 51.24 ms

Status code distribution:
```

This test performs a constant load at `200` RPS, starting with `5` workers, and increasing concurrency by `5` workers every `5s` until we have `50` workers. At that point all `50` workers will be used to sustain the constant `200` RPS until `10000` total request limit is reached. Worker count over time would look something like:

![Step Up Concurrency Constant Load](/images/step_up_c_const_rps_wc.svg)

## Step Down Concurrency

```
./dist/ghz --insecure --async --proto /protos/helloworld.proto \
  --call helloworld.Greeter/SayHello \
  -n 10000 --rps 200 \
  --concurrency-schedule=step --concurrency-start=50 --concurrency-step=-5 \
  --concurrency-step-duration=5s --concurrency-max-duration=30s \
  -d '{"name":"{{.WorkerID}}"}' 0.0.0.0:50051

Summary:
  Count:	10000
  Total:	50.05 s
  Slowest:	52.13 ms
  Fastest:	50.15 ms
  Average:	50.63 ms
  Requests/sec:	199.79

Response time histogram:
  50.152 [1]	|
  50.350 [1145]	|∎∎∎∎∎∎∎∎∎∎∎∎∎
  50.548 [2476]	|∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎
  50.746 [3491]	|∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎
  50.943 [2202]	|∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎
  51.141 [490]	|∎∎∎∎∎∎
  51.339 [148]	|∎∎
  51.536 [30]	|
  51.734 [10]	|
  51.932 [4]	|
  52.130 [3]	|

Latency distribution:
  10 % in 50.34 ms
  25 % in 50.47 ms
  50 % in 50.63 ms
  75 % in 50.77 ms
  90 % in 50.89 ms
  95 % in 50.99 ms
  99 % in 51.24 ms

Status code distribution:
  [OK]   10000 responses
```

This test performs a constant load at `200` RPS, starting with `50` workers, and decreasing concurrency by `5` workers every `5s` until `30s` has elapsed. At that point all remaining workers will be used to sustain the constant `200` RPS until `10000` total request limit is reached. Worker count over time would look something like:

![Step Down Concurrency Constant Load](/images/step_down_c_const_rps_wc.svg)

## Linear increase of concurrency

```
./dist/ghz --insecure --async --proto /protos/helloworld.proto \
  --call helloworld.Greeter/SayHello \
  -n 10000 --rps 200 \
  --concurrency-schedule=line --concurrency-start=20 --concurrency-step=2 --concurrency-max-duration=30s \
  -d '{"name":"{{.WorkerID}}"}' 0.0.0.0:50051

Summary:
  Count:	10000
  Total:	50.05 s
  Slowest:	58.54 ms
  Fastest:	50.16 ms
  Average:	50.60 ms
  Requests/sec:	199.79

Response time histogram:
  50.157 [1]	|
  50.995 [9515]	|∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎
  51.834 [477]	|∎∎
  52.672 [3]	|
  53.510 [0]	|
  54.349 [0]	|
  55.187 [1]	|
  56.025 [0]	|
  56.864 [0]	|
  57.702 [1]	|
  58.540 [2]	|

Latency distribution:
  10 % in 50.31 ms
  25 % in 50.40 ms
  50 % in 50.60 ms
  75 % in 50.75 ms
  90 % in 50.89 ms
  95 % in 50.99 ms
  99 % in 51.25 ms

Status code distribution:
  [OK]   10000 responses
```

This test performs a constant load at `200` RPS, starting with `20` workers, and increasing concurrency linearly every `1s` by `2` workers until `30s` has elapsed. At that point all remaining workers will be used to sustain the constant `200` RPS until `10000` total request limit is reached.

![Lene Up Concurrency Constant Load](/images/line_up_c_const_rps_wc.svg)
