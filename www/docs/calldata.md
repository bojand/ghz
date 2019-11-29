---
id: calldata
title: Call Template Data
---

Data and metadata can specify [template actions](https://golang.org/pkg/text/template/) that will be parsed and evaluated at every request. Each request gets a new instance of the data. The available variables / actions are:

```go
// call template data
type callTemplateData struct {

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

This can be useful to inject variable information into the message data JSON or metadata JSON payloads for each request, such as timestamp or unique request number. For example:

```
-m '{"request-id":"{{.RequestNumber}}", "timestamp":"{{.TimestampUnix}}"}'
```

Would result in server getting the following metadata map represented here in JSON:

```json
{
  "user-agent": "grpc-go/1.11.1",
  "request-id": "1",
  "timestamp": "1544890251"
}
```

See [example calls](examples.md) for some more usage examples.
