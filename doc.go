/*
Package ghz can be used to perform benchmarking and load testing against gRPC services.

	report, err := runner.Run(
		"helloworld.Greeter.SayHello",
		"localhost:50051",
		runner.WithProtoFile("greeter.proto", []string{}),
		runner.WithDataFromFile("data.json"),
		runner.WithInsecure(true),
	)
*/
package ghz
