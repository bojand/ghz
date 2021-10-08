---
id: calldata
title: Call Data
---

Data and metadata can specify [template actions](https://golang.org/pkg/text/template/) that will be parsed and evaluated at every request. Each request gets a new instance of the data. The available variables / actions are:

```go
// CallData represents contextualized data available for templating
type CallData struct {

	// unique worker ID
	WorkerID		   string

	// unique incremented request number for each call
	RequestNumber      int64

	// fully-qualified name of the method call
	FullyQualifiedName string

	// shorter call method name
	MethodName         string

	// the service name
	ServiceName        string

	// name of the input message type
	InputName          string

	// name of the output message type
	OutputName         string

	// whether this call is client streaming
	IsClientStreaming  bool

	// whether this call is server streaming
	IsServerStreaming  bool

	// timestamp of the call in RFC3339 format
	Timestamp          string

	// timestamp of the call as unix time in seconds
	TimestampUnix      int64

	// timestamp of the call as unix time in milliseconds
	TimestampUnixMilli      int64

	// timestamp of the call as unix time in nanoseconds
	TimestampUnixNano      int64

	// UUID v4 for each call
	UUID	string
}
```

**Template Functions**

There are also template functions available:

`func newUUID() string`  
Generates a new UUID for each invocation.

`func randomString(length int) string`  
Generates a new random string for each incovation. Accepts a length parameter. If the argument is `<= 0` then a random string is generated with a random length between length of `2` and `16`.

`func randomInt(min, max int) int`  
Generates a new non-negative pseudo-random number in range `[min, max)`.

You can also use [sprig functions](http://masterminds.github.io/sprig/) within a template.

**Examples**

This can be useful to inject variable information into the message data JSON or metadata JSON payloads for each request, such as timestamp or unique request number. For example:

```sh
-m '{"request-id":"{{.RequestNumber}}", "timestamp":"{{.TimestampUnix}}"}'
```

Would result in server getting the following metadata map represented here in JSON:

```json
{
  "user-agent": "grpc-go/1.11.1",
  "request-id": "1",
  "timestamp": "1544890252"
}
```

```sh
-d '{"order_id":"{{newUUID}}", "item_id":"{{newUUID}}", "sku":"{{randomString 8 }}", "product_name":"{{randomString 0}}"}'
```

Would result in data with JSON representation:

```json
{
  "order_id": "3974e7b3-5946-4df5-bed3-8c3dc9a0be19",
  "item_id": "cd9c2604-cd9b-43a8-9cbb-d1ad26ca93a4",
  "sku": "HlFTAxcm",
  "product_name": "xg3NEC"
}
```

See [example calls](examples.md) for some more usage examples.

### Data Function API

When using the `ghz/runner` package programmatically, we can dynamically create data for each request using `WithBinaryDataFunc()` API:

```go
func dataFunc(mtd *desc.MethodDescriptor, cd *runner.CallData) []byte {
	msg := &helloworld.HelloRequest{}
	msg.Name = cd.WorkerID
	binData, err := proto.Marshal(msg)
	return binData
}

report, err := runner.Run(
	"helloworld.Greeter.SayHello",
	"0.0.0.0:50051",
	runner.WithProtoFile("./testdata/greeter.proto", []string{}),
	runner.WithInsecure(true),
	runner.WithBinaryDataFunc(dataFunc),
)
```
