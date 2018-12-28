---
id: usage
title: Usage
---

```sh
Usage: ghz [options...] host
Options:

-proto		The Protocol Buffer .proto file.
-protoset	The compiled protoset file. Alternative to proto. -proto takes precedence.
-call		A fully-qualified method name in 'package/service/method' or 'package.service.method' format.
-cert		The file containing the CA root cert file. Ignored if -insecure is specified.
-cname		An server name override.
-insecure	Specify for non TLS connection.
-config		Path to the JSON or TOML config file that specifies all the test settings.

-c  Number of requests to run concurrently.
    Total number of requests cannot be smaller than the concurrency level. Default is 50.
-n  Number of requests to run. Default is 200.
-q  Rate limit, in queries per second (QPS). Default is no rate limit.
-t  Timeout for each request in seconds. Default is 20, use 0 for infinite.
-z  Duration of application to send requests. When duration is reached,
	application stops and exits. If duration is specified, n is ignored.
	Examples: -z 10s -z 3m.
-x  Maximum duration of application to send requests with n setting respected.
    If duration is reached before n requests are completed, application stops and exits.
    Examples: -x 10s -x 3m.

-d  The call data as stringified JSON.
    If the value is '@' then the request contents are read from stdin.
-D  Path for call data JSON file. For example, /home/user/file.json or ./file.json.
-b  The call data comes as serialized binary message read from stdin.
-B  Path for the call data as serialized binary message.
-m  Request metadata as stringified JSON.
-M  Path for call metadata JSON file. For example, /home/user/metadata.json or ./metadata.json.

-o  Output path. If none provided stdout is used.
-O  Output type. If none provided, a summary is printed.
    "csv" outputs the response metrics in comma-separated values format.
    "json" outputs the metrics report in JSON format.
    "pretty" outputs the metrics report in pretty JSON format.
    "html" outputs the metrics report as HTML.
    "influx-summary" outputs the metrics summary as influxdb line protocol.
    "influx-details" outputs the metrics details as influxdb line protocol.

-i  Comma separated list of proto import paths. The current working directory and the directory
    of the protocol buffer file are automatically added to the import list.

-T  Connection timeout in seconds for the initial connection dial. Default is 10.
-L  Keepalive time in seconds. Only used if present and above 0.

-name  User specified name for the test.
-tags  JSON representation of user-defined string tags.

-cpus  Number of used cpu cores. (default for current machine is 8 cores)

-v  Print the version.
```
