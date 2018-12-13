---
id: package
title: Package
---

`ghz` can be used programatically as Go package within Go applications. See detailed [godoc](https://godoc.org/github.com/bojand/ghz) documentation. Example usage:


```go
package main

import (
	"fmt"
	"github.com/bojand/ghz/runner"
)

func main() {
    report, err := runner.Run(
		"helloworld.Greeter.SayHello",
		"localhost:50051",
		WithProtoFile("greeter.proto", []string{}),
		WithDataFromFile("data.json"),
		WithInsecure(true),
    )
    
    fmt.Println(report)
}

```
