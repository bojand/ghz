---
id: usage
title: Usage
---

```
usage: ghz [<flags>] [<host>]

Flags:
  -h, --help                   Show context-sensitive help (also try --help-long and --help-man).
      --config=                Path to the JSON or TOML config file that specifies all the test run settings.
      --proto=                 The Protocol Buffer .proto file.
      --protoset=              The compiled protoset file. Alternative to proto. -proto takes precedence.
      --call=                  A fully-qualified method name in 'package.Service/method' or 'package.Service.Method' format.
  -i, --import-paths=          Comma separated list of proto import paths. The current working directory and the directory of the protocol buffer file are automatically added to the import list.
      --cacert=                File containing trusted root certificates for verifying the server.
      --cert=                  File containing client certificate (public key), to present to the server. Must also provide -key option.
      --key=                   File containing client private key, to present to the server. Must also provide -cert option.
      --cname=                 Server name override when validating TLS certificate - useful for self signed certs.
      --skipTLS                Skip TLS client verification of the server's certificate chain and host name.
      --insecure               Use plaintext and insecure connection.
      --authority=             Value to be used as the :authority pseudo-header. Only works if -insecure is used.
      --async                  Make requests asynchronous as soon as possible. Does not wait for request to finish before sending next one.
  -r, --rps=0                  Requests per second (RPS) rate limit for constant load schedule. Default is no rate limit.
      --load-schedule="const"  Specifies the load schedule. Options are const, step, or line. Default is const.
      --load-start=0           Specifies the RPS load start value for step or line schedules.
      --load-step=0            Specifies the load step value or slope value.
      --load-end=0             Specifies the load end value for step or line load schedules.
      --load-step-duration=0   Specifies the load step duration value for step load schedule.
      --load-max-duration=0    Specifies the max load duration value for step or line load schedule.
  -c, --concurrency=50         Number of request workers to run concurrently for const concurrency schedule. Default is 50.
      --concurrency-schedule="const"
                               Concurrency change schedule. Options are const, step, or line. Default is const.
      --concurrency-start=0    Concurrency start value for step and line concurrency schedules.
      --concurrency-end=0      Concurrency end value for step and line concurrency schedules.
      --concurrency-step=1     Concurrency step / slope value for step and line concurrency schedules.
      --concurrency-step-duration=0
                               Specifies the concurrency step duration value for step concurrency schedule.
      --concurrency-max-duration=0
                               Specifies the max concurrency adjustment duration value for step or line concurrency schedule.
  -n, --total=200              Number of requests to run. Default is 200.
  -t, --timeout=20s            Timeout for each request. Default is 20s, use 0 for infinite.
  -z, --duration=0             Duration of application to send requests. When duration is reached, application stops and exits. If duration is specified, n is ignored. Examples: -z 10s -z 3m.
  -x, --max-duration=0         Maximum duration of application to send requests with n setting respected. If duration is reached before n requests are completed, application stops and exits. Examples: -x 10s -x 3m.
      --duration-stop="close"  Specifies how duration stop is reported. Options are close, wait or ignore. Default is close.
  -d, --data=                  The call data as stringified JSON. If the value is '@' then the request contents are read from stdin.
  -D, --data-file=             File path for call data JSON file. Examples: /home/user/file.json or ./file.json.
  -b, --binary                 The call data comes as serialized binary message or multiple count-prefixed messages read from stdin.
  -B, --binary-file=           File path for the call data as serialized binary message or multiple count-prefixed messages.
  -m, --metadata=              Request metadata as stringified JSON.
  -M, --metadata-file=         File path for call metadata JSON file. Examples: /home/user/metadata.json or ./metadata.json.
      --stream-interval=0      Interval for stream requests between message sends.
      --reflect-metadata=      Reflect metadata as stringified JSON used only for reflection request.
  -o, --output=                Output path. If none provided stdout is used.
  -O, --format=                Output format. One of: summary, csv, json, pretty, html, influx-summary, influx-details. Default is summary.
      --skipFirst=0            Skip the first X requests when doing the results tally.
      --connections=1          Number of connections to use. Concurrency is distributed evenly among all the connections. Default is 1.
      --connect-timeout=10s    Connection timeout for the initial connection dial. Default is 10s.
      --keepalive=0            Keepalive time duration. Only used if present and above 0.
      --name=                  User specified name for the test.
      --tags=                  JSON representation of user-defined string tags.
      --cpus=12                Number of cpu cores to use.
      --debug=                 The path to debug log file.
  -e, --enable-compression     Enable Gzip compression on requests.
  -v, --version                Show application version.

Args:
  [<host>]  Host and port to test.
```
