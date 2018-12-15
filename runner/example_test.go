package runner_test

import (
	"fmt"
	"os"

	"github.com/bojand/ghz/printer"
	"github.com/bojand/ghz/runner"
)

func ExampleRun() {
	report, err := runner.Run(
		"helloworld.Greeter.SayHello",
		"localhost:50051",
		runner.WithProtoFile("../testdata/greeter.proto", []string{}),
		runner.WithDataFromFile("../testdata/data.json"),
		runner.WithInsecure(true),
	)

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	printer := printer.ReportPrinter{
		Out:    os.Stdout,
		Report: report,
	}

	printer.Print("pretty")
}
