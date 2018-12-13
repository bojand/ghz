---
id: calldata
title: Call Template Data
---

Data and metadata can specify [template actions](https://golang.org/pkg/text/template/) that will be parsed and evaluated at every request. Each request gets a new instance of the data. The available variables / actions are:

```go
// call template data
type callTemplateData struct {

	// unique incremented request number for each request
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

	// timestamp of the call as unix time
	TimestampUnix      int64
}
```

This can be useful to inject variable information into the data or metadata payload for each request, such as timestamp or unique request number. See examples below.
