package ghz

import (
	"fmt"
	"testing"
	"time"

	"github.com/bojand/ghz/internal/helloworld"
	"github.com/stretchr/testify/assert"
)

func TestRunUnary(t *testing.T) {
	callType := helloworld.Unary

	gs, s, err := startServer(false)

	if err != nil {
		assert.FailNow(t, err.Error())
	}

	defer s.Stop()

	data := make(map[string]interface{})
	data["name"] = "bob"

	t.Run("test report", func(t *testing.T) {
		gs.ResetCounters()

		report, err := Run(
			"helloworld.Greeter.SayHello",
			localhost,
			WithProtoFile("./testdata/greeter.proto", []string{}),
			WithN(1),
			WithC(1),
			WithTimeout(time.Duration(20*time.Second)),
			WithData(data),
			WithInsecure,
		)

		fmt.Printf("%#v\n", report)

		assert.NoError(t, err)
		assert.NotNil(t, report)
		assert.Equal(t, 1, int(report.Count))
		assert.Len(t, report.ErrorDist, 0)

		count := gs.GetCount(callType)
		assert.Equal(t, 1, count)
	})
}
