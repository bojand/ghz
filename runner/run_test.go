package runner

import (
	"fmt"
	"strconv"
	"testing"
	"text/template"
	"time"

	"github.com/bojand/ghz/internal"
	"github.com/bojand/ghz/internal/helloworld"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/google/uuid"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/metadata"
)

func changeFunc(mtd *desc.MethodDescriptor, cd *CallData) []byte {
	msg := &helloworld.HelloRequest{}
	msg.Name = "bob"
	binData, _ := proto.Marshal(msg)
	return binData
}

func TestRunUnary(t *testing.T) {
	callType := helloworld.Unary

	gs, s, err := internal.StartServer(false)

	if err != nil {
		assert.FailNow(t, err.Error())
	}

	defer s.Stop()

	t.Run("test report", func(t *testing.T) {
		gs.ResetCounters()

		data := make(map[string]interface{})
		data["name"] = "bob"

		report, err := Run(
			"helloworld.Greeter.SayHello",
			internal.TestLocalhost,
			WithProtoFile("../testdata/greeter.proto", []string{}),
			WithTotalRequests(1),
			WithConcurrency(1),
			WithTimeout(time.Duration(20*time.Second)),
			WithDialTimeout(time.Duration(20*time.Second)),
			WithData(data),
			WithInsecure(true),
		)

		assert.NoError(t, err)

		assert.NotNil(t, report)

		assert.Equal(t, 1, int(report.Count))
		assert.NotZero(t, report.Average)
		assert.NotZero(t, report.Fastest)
		assert.NotZero(t, report.Slowest)
		assert.NotZero(t, report.Rps)
		assert.Empty(t, report.Name)
		assert.NotEmpty(t, report.Date)
		assert.NotEmpty(t, report.Options)
		assert.NotEmpty(t, report.Details)
		assert.Equal(t, true, report.Options.Insecure)
		assert.NotEmpty(t, report.LatencyDistribution)
		assert.Equal(t, ReasonNormalEnd, report.EndReason)
		assert.Empty(t, report.ErrorDist)

		assert.Equal(t, report.Average, report.Slowest)
		assert.Equal(t, report.Average, report.Fastest)

		count := gs.GetCount(callType)
		assert.Equal(t, 1, count)
	})

	t.Run("test predefined template functions", func(t *testing.T) {
		gs.ResetCounters()

		data := make(map[string]interface{})
		data["name"] = "{{ newUUID }}"

		report, err := Run(
			"helloworld.Greeter.SayHello",
			internal.TestLocalhost,
			WithProtoFile("../testdata/greeter.proto", []string{}),
			WithTotalRequests(1),
			WithConcurrency(1),
			WithTimeout(time.Duration(20*time.Second)),
			WithData(data),
			WithInsecure(true),
		)

		assert.NoError(t, err)
		assert.NotNil(t, report)

		count := gs.GetCount(callType)
		assert.Equal(t, 1, count)

		calls := gs.GetCalls(callType)
		assert.NotNil(t, calls)
		assert.Len(t, calls, 1)

		msg := calls[0][0]
		parsed, err := uuid.Parse(msg.GetName())
		assert.NoError(t, err)
		assert.NotEmpty(t, parsed)
	})

	t.Run("test sprig template functions", func(t *testing.T) {
		gs.ResetCounters()

		data := make(map[string]interface{})
		data["name"] = "{{ `test` | upper | repeat 5 }}"

		report, err := Run(
			"helloworld.Greeter.SayHello",
			internal.TestLocalhost,
			WithProtoFile("../testdata/greeter.proto", []string{}),
			WithTotalRequests(1),
			WithConcurrency(1),
			WithTimeout(time.Duration(20*time.Second)),
			WithData(data),
			WithInsecure(true),
		)

		assert.NoError(t, err)
		assert.NotNil(t, report)

		count := gs.GetCount(callType)
		assert.Equal(t, 1, count)

		calls := gs.GetCalls(callType)
		assert.NotNil(t, calls)
		assert.Len(t, calls, 1)

		msg := calls[0][0]
		assert.Equal(t, msg.GetName(), "TESTTESTTESTTESTTEST")
	})

	t.Run("test custom template functions", func(t *testing.T) {
		gs.ResetCounters()

		data := make(map[string]interface{})
		data["name"] = "{{customFunc}}"

		report, err := Run(
			"helloworld.Greeter.SayHello",
			internal.TestLocalhost,
			WithProtoFile("../testdata/greeter.proto", []string{}),
			WithTotalRequests(1),
			WithConcurrency(1),
			WithTimeout(time.Duration(20*time.Second)),
			WithData(data),
			WithInsecure(true),
			WithTemplateFuncs(template.FuncMap{
				"customFunc": func() string {
					return "custom-value"
				},
			}),
		)

		assert.NoError(t, err)
		assert.NotNil(t, report)

		count := gs.GetCount(callType)
		assert.Equal(t, 1, count)

		calls := gs.GetCalls(callType)
		assert.NotNil(t, calls)
		assert.Len(t, calls, 1)

		msg := calls[0][0]
		assert.Equal(t, msg.GetName(), "custom-value")
	})

	t.Run("test custom template functions are added to predefined functions", func(t *testing.T) {
		gs.ResetCounters()

		data := make(map[string]interface{})
		data["name"] = "{{newUUID}}"

		report, err := Run(
			"helloworld.Greeter.SayHello",
			internal.TestLocalhost,
			WithProtoFile("../testdata/greeter.proto", []string{}),
			WithTotalRequests(1),
			WithConcurrency(1),
			WithTimeout(time.Duration(20*time.Second)),
			WithData(data),
			WithInsecure(true),
			WithTemplateFuncs(template.FuncMap{
				"customFunc": func() string {
					return "custom-value"
				},
			}),
		)

		assert.NoError(t, err)
		assert.NotNil(t, report)

		count := gs.GetCount(callType)
		assert.Equal(t, 1, count)

		calls := gs.GetCalls(callType)
		assert.NotNil(t, calls)
		assert.Len(t, calls, 1)

		msg := calls[0][0]
		parsed, err := uuid.Parse(msg.GetName())
		assert.NoError(t, err)
		assert.NotEmpty(t, parsed)
	})

	t.Run("test skip first N", func(t *testing.T) {
		gs.ResetCounters()

		data := make(map[string]interface{})
		data["name"] = "bob"

		report, err := Run(
			"helloworld.Greeter.SayHello",
			internal.TestLocalhost,
			WithProtoFile("../testdata/greeter.proto", []string{}),
			WithTotalRequests(10),
			WithConcurrency(1),
			WithTimeout(time.Duration(20*time.Second)),
			WithDialTimeout(time.Duration(20*time.Second)),
			WithData(data),
			WithSkipFirst(5),
			WithInsecure(true),
		)

		assert.NoError(t, err)

		assert.NotNil(t, report)

		assert.Equal(t, 5, int(report.Count))
		assert.NotZero(t, report.Average)
		assert.NotZero(t, report.Fastest)
		assert.NotZero(t, report.Slowest)
		assert.NotZero(t, report.Rps)
		assert.Empty(t, report.Name)
		assert.NotEmpty(t, report.Date)
		assert.NotEmpty(t, report.Options)
		assert.NotEmpty(t, report.Details)
		assert.Equal(t, true, report.Options.Insecure)
		assert.NotEmpty(t, report.LatencyDistribution)
		assert.Equal(t, ReasonNormalEnd, report.EndReason)
		assert.Empty(t, report.ErrorDist)

		assert.NotEqual(t, report.Average, report.Slowest)
		assert.NotEqual(t, report.Average, report.Fastest)
		assert.NotEqual(t, report.Slowest, report.Fastest)

		count := gs.GetCount(callType)
		assert.Equal(t, 10, count)
	})

	t.Run("test N and Name", func(t *testing.T) {
		gs.ResetCounters()

		data := make(map[string]interface{})
		data["name"] = "bob"

		report, err := Run(
			"helloworld.Greeter.SayHello",
			internal.TestLocalhost,
			WithProtoFile("../testdata/greeter.proto", []string{}),
			WithTotalRequests(12),
			WithConcurrency(2),
			WithTimeout(time.Duration(20*time.Second)),
			WithDialTimeout(time.Duration(20*time.Second)),
			WithData(data),
			WithName("test123"),
			WithInsecure(true),
		)

		assert.NoError(t, err)

		assert.NotNil(t, report)

		assert.Equal(t, 12, int(report.Count))
		assert.NotZero(t, report.Average)
		assert.NotZero(t, report.Fastest)
		assert.NotZero(t, report.Slowest)
		assert.NotZero(t, report.Rps)
		assert.Equal(t, "test123", report.Name)
		assert.NotEmpty(t, report.Date)
		assert.NotEmpty(t, report.Options)
		assert.NotEmpty(t, report.Details)
		assert.Equal(t, true, report.Options.Insecure)
		assert.NotEmpty(t, report.LatencyDistribution)
		assert.Equal(t, ReasonNormalEnd, report.EndReason)
		assert.Empty(t, report.ErrorDist)

		assert.NotEqual(t, report.Average, report.Slowest)
		assert.NotEqual(t, report.Average, report.Fastest)
		assert.NotEqual(t, report.Slowest, report.Fastest)

		count := gs.GetCount(callType)
		assert.Equal(t, 12, count)

		connCount := gs.GetConnectionCount()
		assert.Equal(t, 1, connCount)
	})

	t.Run("test n default", func(t *testing.T) {
		gs.ResetCounters()

		data := make(map[string]interface{})
		data["name"] = "bob"

		report, err := Run(
			"helloworld.Greeter.SayHello",
			internal.TestLocalhost,
			WithProtoFile("../testdata/greeter.proto", []string{}),
			WithData(data),
			WithInsecure(true),
		)

		assert.NoError(t, err)

		assert.NotNil(t, report)

		assert.Equal(t, 200, int(report.Count))
		assert.NotZero(t, report.Average)
		assert.NotZero(t, report.Fastest)
		assert.NotZero(t, report.Slowest)
		assert.NotZero(t, report.Rps)
		assert.NotEmpty(t, report.Date)
		assert.NotEmpty(t, report.Options)
		assert.NotEmpty(t, report.Details)
		assert.Equal(t, true, report.Options.Insecure)
		assert.NotEmpty(t, report.LatencyDistribution)
		assert.Equal(t, ReasonNormalEnd, report.EndReason)
		assert.Empty(t, report.ErrorDist)

		assert.NotEqual(t, report.Average, report.Slowest)
		assert.NotEqual(t, report.Average, report.Fastest)
		assert.NotEqual(t, report.Slowest, report.Fastest)

		count := gs.GetCount(callType)
		assert.Equal(t, 200, count)

		connCount := gs.GetConnectionCount()
		assert.Equal(t, 1, connCount)
	})

	t.Run("test run duration", func(t *testing.T) {
		gs.ResetCounters()

		data := make(map[string]interface{})
		data["name"] = "bob"

		report, err := Run(
			"helloworld.Greeter.SayHello",
			internal.TestLocalhost,
			WithProtoFile("../testdata/greeter.proto", []string{}),
			WithRunDuration(3*time.Second),
			WithData(data),
			WithInsecure(true),
		)

		assert.NoError(t, err)

		assert.NotNil(t, report)

		assert.True(t, int(report.Count) > 200, fmt.Sprintf("%d not > 200", int64(report.Count)))
		assert.True(t, report.Total.Milliseconds() >= 3000 && report.Total.Milliseconds() < 3100, fmt.Sprintf("duration %d expected value", report.Total.Milliseconds()))
		assert.NotZero(t, report.Average)
		assert.NotZero(t, report.Fastest)
		assert.NotZero(t, report.Slowest)
		assert.NotZero(t, report.Rps)
		assert.NotEmpty(t, report.Date)
		assert.NotEmpty(t, report.Options)
		assert.NotEmpty(t, report.Details)
		assert.Equal(t, true, report.Options.Insecure)
		assert.NotEmpty(t, report.LatencyDistribution)
		assert.Equal(t, ReasonTimeout, report.EndReason)
		// assert.Empty(t, report.ErrorDist)

		assert.NotEqual(t, report.Average, report.Slowest)
		assert.NotEqual(t, report.Average, report.Fastest)
		assert.NotEqual(t, report.Slowest, report.Fastest)

		// count := gs.GetCount(callType)
		// assert.Equal(t, int64(report.Count), int64(count))

		connCount := gs.GetConnectionCount()
		assert.Equal(t, 1, connCount)
	})

	t.Run("test RPS", func(t *testing.T) {

		gs.ResetCounters()

		data := make(map[string]interface{})
		data["name"] = "bob"

		report, err := Run(
			"helloworld.Greeter.SayHello",
			internal.TestLocalhost,
			WithProtoFile("../testdata/greeter.proto", []string{}),
			WithTotalRequests(10),
			WithConcurrency(2),
			WithRPS(1),
			WithTimeout(time.Duration(20*time.Second)),
			WithDialTimeout(time.Duration(20*time.Second)),
			WithData(data),
			WithInsecure(true),
		)

		assert.NoError(t, err)

		assert.NotNil(t, report)

		assert.Equal(t, 10, int(report.Count))
		assert.NotZero(t, report.Average)
		assert.NotZero(t, report.Fastest)
		assert.NotZero(t, report.Slowest)
		assert.NotZero(t, report.Rps)
		assert.Empty(t, report.Name)
		assert.NotEmpty(t, report.Date)
		assert.NotEmpty(t, report.Options)
		assert.NotEmpty(t, report.Details)
		assert.Equal(t, true, report.Options.Insecure)
		assert.NotEmpty(t, report.LatencyDistribution)
		assert.Equal(t, ReasonNormalEnd, report.EndReason)
		assert.Empty(t, report.ErrorDist)

		assert.NotEqual(t, report.Average, report.Slowest)
		assert.NotEqual(t, report.Average, report.Fastest)
		assert.NotEqual(t, report.Slowest, report.Fastest)
	})

	t.Run("test binary", func(t *testing.T) {
		gs.ResetCounters()

		msg := &helloworld.HelloRequest{}
		msg.Name = "bob"

		binData, err := proto.Marshal(msg)
		assert.NoError(t, err)

		report, err := Run(
			"helloworld.Greeter.SayHello",
			internal.TestLocalhost,
			WithProtoFile("../testdata/greeter.proto", []string{}),
			WithTotalRequests(5),
			WithConcurrency(1),
			WithTimeout(time.Duration(20*time.Second)),
			WithDialTimeout(time.Duration(20*time.Second)),
			WithBinaryData(binData),
			WithInsecure(true),
		)

		assert.NoError(t, err)

		assert.NotNil(t, report)

		assert.Equal(t, 5, int(report.Count))
		assert.NotZero(t, report.Average)
		assert.NotZero(t, report.Fastest)
		assert.NotZero(t, report.Slowest)
		assert.NotZero(t, report.Rps)
		assert.Empty(t, report.Name)
		assert.NotEmpty(t, report.Date)
		assert.NotEmpty(t, report.Options)
		assert.NotEmpty(t, report.Details)
		assert.NotEmpty(t, report.LatencyDistribution)
		assert.Equal(t, ReasonNormalEnd, report.EndReason)
		assert.Empty(t, report.ErrorDist)

		assert.NotEqual(t, report.Average, report.Slowest)
		assert.NotEqual(t, report.Average, report.Fastest)
		assert.NotEqual(t, report.Slowest, report.Fastest)

		count := gs.GetCount(callType)
		assert.Equal(t, 5, count)

		connCount := gs.GetConnectionCount()
		assert.Equal(t, 1, connCount)
	})

	t.Run("test binary with func", func(t *testing.T) {

		gs.ResetCounters()

		report, err := Run(
			"helloworld.Greeter.SayHello",
			internal.TestLocalhost,
			WithProtoFile("../testdata/greeter.proto", []string{}),
			WithTotalRequests(5),
			WithBinaryDataFunc(changeFunc),
			WithConcurrency(1),
			WithTimeout(time.Duration(20*time.Second)),
			WithDialTimeout(time.Duration(20*time.Second)),
			WithInsecure(true),
		)

		assert.NoError(t, err)

		assert.NotNil(t, report)

		assert.Equal(t, 5, int(report.Count))
		assert.NotZero(t, report.Average)
		assert.NotZero(t, report.Fastest)
		assert.NotZero(t, report.Slowest)
		assert.NotZero(t, report.Rps)
		assert.Empty(t, report.Name)
		assert.NotEmpty(t, report.Date)
		assert.NotEmpty(t, report.Options)
		assert.NotEmpty(t, report.Details)
		assert.NotEmpty(t, report.LatencyDistribution)
		assert.Equal(t, ReasonNormalEnd, report.EndReason)
		assert.Empty(t, report.ErrorDist)

		assert.NotEqual(t, report.Average, report.Slowest)
		assert.NotEqual(t, report.Average, report.Fastest)
		assert.NotEqual(t, report.Slowest, report.Fastest)

		count := gs.GetCount(callType)
		assert.Equal(t, 5, count)

		connCount := gs.GetConnectionCount()
		assert.Equal(t, 1, connCount)
	})

	t.Run("test connections", func(t *testing.T) {
		gs.ResetCounters()

		msg := &helloworld.HelloRequest{}
		msg.Name = "bob"

		binData, err := proto.Marshal(msg)
		assert.NoError(t, err)

		report, err := Run(
			"helloworld.Greeter.SayHello",
			internal.TestLocalhost,
			WithProtoFile("../testdata/greeter.proto", []string{}),
			WithTotalRequests(5),
			WithConcurrency(5),
			WithConnections(5),
			WithTimeout(time.Duration(20*time.Second)),
			WithDialTimeout(time.Duration(20*time.Second)),
			WithBinaryData(binData),
			WithInsecure(true),
		)

		assert.NoError(t, err)

		assert.NotNil(t, report)

		assert.Equal(t, 5, int(report.Count))
		assert.NotZero(t, report.Average)
		assert.NotZero(t, report.Fastest)
		assert.NotZero(t, report.Slowest)
		assert.NotZero(t, report.Rps)
		assert.Empty(t, report.Name)
		assert.NotEmpty(t, report.Date)
		assert.NotEmpty(t, report.Options)
		assert.NotEmpty(t, report.Details)
		assert.NotEmpty(t, report.LatencyDistribution)
		assert.Equal(t, ReasonNormalEnd, report.EndReason)
		assert.Empty(t, report.ErrorDist)

		assert.NotEqual(t, report.Average, report.Slowest)
		assert.NotEqual(t, report.Average, report.Fastest)
		assert.NotEqual(t, report.Slowest, report.Fastest)

		count := gs.GetCount(callType)
		assert.Equal(t, 5, count)

		connCount := gs.GetConnectionCount()
		assert.Equal(t, 5, connCount)
	})

	t.Run("test round-robin c = 2", func(t *testing.T) {
		gs.ResetCounters()

		data := make([]map[string]interface{}, 3)
		for i := 0; i < 3; i++ {
			data[i] = make(map[string]interface{})
			data[i]["name"] = strconv.Itoa(i)
		}

		report, err := Run(
			"helloworld.Greeter.SayHello",
			internal.TestLocalhost,
			WithProtoFile("../testdata/greeter.proto", []string{}),
			WithTotalRequests(6),
			WithConcurrency(2),
			WithTimeout(time.Duration(20*time.Second)),
			WithDialTimeout(time.Duration(20*time.Second)),
			WithInsecure(true),
			WithData(data),
		)

		assert.NoError(t, err)
		assert.NotNil(t, report)

		count := gs.GetCount(callType)
		assert.Equal(t, 6, count)

		calls := gs.GetCalls(callType)
		assert.NotNil(t, calls)
		assert.Len(t, calls, 6)
		names := make([]string, 0)
		for _, msgs := range calls {
			for _, msg := range msgs {
				names = append(names, msg.GetName())
			}
		}

		// we don't expect to have the same order of elements since requests are concurrent
		assert.ElementsMatch(t, []string{"0", "1", "2", "0", "1", "2"}, names)
	})

	t.Run("test round-robin c = 1", func(t *testing.T) {
		gs.ResetCounters()

		data := make([]map[string]interface{}, 3)
		for i := 0; i < 3; i++ {
			data[i] = make(map[string]interface{})
			data[i]["name"] = strconv.Itoa(i)
		}

		report, err := Run(
			"helloworld.Greeter.SayHello",
			internal.TestLocalhost,
			WithProtoFile("../testdata/greeter.proto", []string{}),
			WithTotalRequests(6),
			WithConcurrency(1),
			WithTimeout(time.Duration(20*time.Second)),
			WithDialTimeout(time.Duration(20*time.Second)),
			WithInsecure(true),
			WithData(data),
		)

		assert.NoError(t, err)
		assert.NotNil(t, report)

		count := gs.GetCount(callType)
		assert.Equal(t, 6, count)

		calls := gs.GetCalls(callType)
		assert.NotNil(t, calls)
		assert.Len(t, calls, 6)
		names := make([]string, 0)
		for _, msgs := range calls {
			for _, msg := range msgs {
				names = append(names, msg.GetName())
			}
		}

		// we expect the same order of messages with single worker
		assert.Equal(t, []string{"0", "1", "2", "0", "1", "2"}, names)
	})

	t.Run("test round-robin binary", func(t *testing.T) {
		gs.ResetCounters()

		buf := proto.Buffer{}
		for i := 0; i < 3; i++ {
			msg := &helloworld.HelloRequest{}
			msg.Name = strconv.Itoa(i)
			err = buf.EncodeMessage(msg)
			assert.NoError(t, err)
		}
		binData := buf.Bytes()

		report, err := Run(
			"helloworld.Greeter.SayHello",
			internal.TestLocalhost,
			WithProtoFile("../testdata/greeter.proto", []string{}),
			WithTotalRequests(6),
			WithConcurrency(1),
			WithTimeout(time.Duration(20*time.Second)),
			WithDialTimeout(time.Duration(20*time.Second)),
			WithInsecure(true),
			WithBinaryData(binData),
		)

		assert.NoError(t, err)
		assert.NotNil(t, report)

		count := gs.GetCount(callType)
		assert.Equal(t, 6, count)

		calls := gs.GetCalls(callType)
		assert.NotNil(t, calls)
		assert.Len(t, calls, 6)
		names := make([]string, 0)
		for _, msgs := range calls {
			for _, msg := range msgs {
				names = append(names, msg.GetName())
			}
		}

		assert.Equal(t, []string{"0", "1", "2", "0", "1", "2"}, names)
	})

	t.Run("with data provider", func(t *testing.T) {
		gs.ResetCounters()

		callCounter := 0

		report, err := Run(
			"helloworld.Greeter.SayHello",
			internal.TestLocalhost,
			WithProtoFile("../testdata/greeter.proto", []string{}),
			WithTotalRequests(5),
			WithConcurrency(1),
			WithTimeout(time.Duration(20*time.Second)),
			WithDialTimeout(time.Duration(20*time.Second)),
			WithDataProvider(func(*CallData) ([]*dynamic.Message, error) {
				protoMsg := &helloworld.HelloRequest{Name: strconv.Itoa(callCounter)}
				dynamicMsg, err := dynamic.AsDynamicMessage(protoMsg)
				if err != nil {
					return nil, err
				}

				callCounter++

				return []*dynamic.Message{dynamicMsg}, nil
			}),
			WithInsecure(true),
		)

		assert.NoError(t, err)

		assert.NotNil(t, report)

		assert.Equal(t, 5, int(report.Count))
		assert.NotZero(t, report.Average)
		assert.NotZero(t, report.Fastest)
		assert.NotZero(t, report.Slowest)
		assert.NotZero(t, report.Rps)
		assert.Empty(t, report.Name)
		assert.NotEmpty(t, report.Date)
		assert.NotEmpty(t, report.Options)
		assert.NotEmpty(t, report.Details)
		assert.Equal(t, true, report.Options.Insecure)
		assert.NotEmpty(t, report.LatencyDistribution)
		assert.Equal(t, ReasonNormalEnd, report.EndReason)
		assert.Empty(t, report.ErrorDist)

		assert.NotEqual(t, report.Average, report.Slowest)
		assert.NotEqual(t, report.Average, report.Fastest)
		assert.NotEqual(t, report.Slowest, report.Fastest)

		count := gs.GetCount(callType)
		assert.Equal(t, 5, count)

		calls := gs.GetCalls(callType)
		assert.NotNil(t, calls)
		assert.Len(t, calls, 5)
		names := make([]string, 0)
		for _, msgs := range calls {
			for _, msg := range msgs {
				names = append(names, msg.GetName())
			}
		}

		assert.Equal(t, []string{"0", "1", "2", "3", "4"}, names)
	})

	t.Run("with data custom function", func(t *testing.T) {
		gs.ResetCounters()

		data := make(map[string]interface{})
		data["name"] = "{{customFunc}}"

		report, err := Run(
			"helloworld.Greeter.SayHello",
			internal.TestLocalhost,
			WithProtoFile("../testdata/greeter.proto", []string{}),
			WithTotalRequests(3),
			WithConcurrency(1),
			WithTimeout(time.Duration(20*time.Second)),
			WithDialTimeout(time.Duration(20*time.Second)),
			WithData(data),
			WithTemplateFuncs(template.FuncMap{
				"customFunc": func() string {
					return "custom-value"
				},
			}),
			WithInsecure(true),
		)

		assert.NoError(t, err)

		assert.NotNil(t, report)

		assert.Equal(t, 3, int(report.Count))
		assert.NotZero(t, report.Average)
		assert.NotZero(t, report.Fastest)
		assert.NotZero(t, report.Slowest)
		assert.NotZero(t, report.Rps)
		assert.Empty(t, report.Name)
		assert.NotEmpty(t, report.Date)
		assert.NotEmpty(t, report.Options)
		assert.NotEmpty(t, report.Details)
		assert.Equal(t, true, report.Options.Insecure)
		assert.NotEmpty(t, report.LatencyDistribution)
		assert.Equal(t, ReasonNormalEnd, report.EndReason)
		assert.Empty(t, report.ErrorDist)

		assert.NotEqual(t, report.Average, report.Slowest)
		assert.NotEqual(t, report.Average, report.Fastest)
		assert.NotEqual(t, report.Slowest, report.Fastest)

		count := gs.GetCount(callType)
		assert.Equal(t, 3, count)

		calls := gs.GetCalls(callType)
		assert.NotNil(t, calls)
		assert.Len(t, calls, 3)
		fmt.Printf("%+v\n", calls)
		assert.Equal(t, "custom-value", calls[0][0].GetName())
		assert.Equal(t, "custom-value", calls[1][0].GetName())
		assert.Equal(t, "custom-value", calls[2][0].GetName())
	})

	t.Run("with metadata provider", func(t *testing.T) {
		gs.ResetCounters()

		callCounter := 0

		report, err := Run(
			"helloworld.Greeter.SayHello",
			internal.TestLocalhost,
			WithProtoFile("../testdata/greeter.proto", []string{}),
			WithTotalRequests(5),
			WithConcurrency(1),
			WithTimeout(time.Duration(20*time.Second)),
			WithDialTimeout(time.Duration(20*time.Second)),
			WithDataProvider(func(*CallData) ([]*dynamic.Message, error) {
				name := strconv.Itoa(callCounter)
				if callCounter == 1 || callCounter == 3 {
					name = "__record_metadata__"
				}
				protoMsg := &helloworld.HelloRequest{Name: name}
				dynamicMsg, err := dynamic.AsDynamicMessage(protoMsg)
				if err != nil {
					return nil, err
				}

				callCounter++

				return []*dynamic.Message{dynamicMsg}, nil
			}),
			WithMetadataProvider(func(*CallData) (*metadata.MD, error) {
				mdv := "secret" + strconv.Itoa(callCounter)
				return &metadata.MD{"token": []string{mdv}}, nil
			}),
			WithInsecure(true),
		)

		assert.NoError(t, err)

		assert.NotNil(t, report)

		assert.Equal(t, 5, int(report.Count))
		assert.NotZero(t, report.Average)
		assert.NotZero(t, report.Fastest)
		assert.NotZero(t, report.Slowest)
		assert.NotZero(t, report.Rps)
		assert.Empty(t, report.Name)
		assert.NotEmpty(t, report.Date)
		assert.NotEmpty(t, report.Options)
		assert.NotEmpty(t, report.Details)
		assert.Equal(t, true, report.Options.Insecure)
		assert.NotEmpty(t, report.LatencyDistribution)
		assert.Equal(t, ReasonNormalEnd, report.EndReason)
		assert.Empty(t, report.ErrorDist)

		assert.NotEqual(t, report.Average, report.Slowest)
		assert.NotEqual(t, report.Average, report.Fastest)
		assert.NotEqual(t, report.Slowest, report.Fastest)

		count := gs.GetCount(callType)
		assert.Equal(t, 5, count)

		calls := gs.GetCalls(callType)
		assert.NotNil(t, calls)
		assert.Len(t, calls, 5)
		names := make([]string, 0)
		for _, msgs := range calls {
			for _, msg := range msgs {
				names = append(names, msg.GetName())
			}
		}

		assert.Equal(t, []string{"0", "__record_metadata__||token:secret1", "2", "__record_metadata__||token:secret3", "4"}, names)
	})

	t.Run("with metadata custom function", func(t *testing.T) {
		gs.ResetCounters()

		data := make(map[string]interface{})
		data["name"] = "__record_metadata__"

		metadata := make(map[string]string)
		metadata["token"] = "{{customFunc}}"

		report, err := Run(
			"helloworld.Greeter.SayHello",
			internal.TestLocalhost,
			WithProtoFile("../testdata/greeter.proto", []string{}),
			WithTotalRequests(3),
			WithConcurrency(1),
			WithTimeout(time.Duration(20*time.Second)),
			WithDialTimeout(time.Duration(20*time.Second)),
			WithData(data),
			WithMetadata(metadata),
			WithTemplateFuncs(template.FuncMap{
				"customFunc": func() string {
					return "custom-value"
				},
			}),
			WithInsecure(true),
		)

		assert.NoError(t, err)

		assert.NotNil(t, report)

		assert.Equal(t, 3, int(report.Count))
		assert.NotZero(t, report.Average)
		assert.NotZero(t, report.Fastest)
		assert.NotZero(t, report.Slowest)
		assert.NotZero(t, report.Rps)
		assert.Empty(t, report.Name)
		assert.NotEmpty(t, report.Date)
		assert.NotEmpty(t, report.Options)
		assert.NotEmpty(t, report.Details)
		assert.Equal(t, true, report.Options.Insecure)
		assert.NotEmpty(t, report.LatencyDistribution)
		assert.Equal(t, ReasonNormalEnd, report.EndReason)
		assert.Empty(t, report.ErrorDist)

		assert.NotEqual(t, report.Average, report.Slowest)
		assert.NotEqual(t, report.Average, report.Fastest)
		assert.NotEqual(t, report.Slowest, report.Fastest)

		count := gs.GetCount(callType)
		assert.Equal(t, 3, count)

		calls := gs.GetCalls(callType)
		assert.NotNil(t, calls)
		assert.Len(t, calls, 3)
		names := make([]string, 0)
		for _, msgs := range calls {
			for _, msg := range msgs {
				names = append(names, msg.GetName())
			}
		}

		assert.Equal(t, []string{
			"__record_metadata__||token:custom-value",
			"__record_metadata__||token:custom-value",
			"__record_metadata__||token:custom-value",
		}, names)
	})
}

func TestRunServerStreaming(t *testing.T) {

	callType := helloworld.ServerStream

	gs, s, err := internal.StartServer(false)

	if err != nil {
		assert.FailNow(t, err.Error())
	}

	defer s.Stop()

	t.Run("basic", func(t *testing.T) {

		gs.ResetCounters()

		data := make(map[string]interface{})
		data["name"] = "bob"

		report, err := Run(
			"helloworld.Greeter.SayHellos",
			internal.TestLocalhost,
			WithProtoFile("../testdata/greeter.proto", []string{}),
			WithTotalRequests(15),
			WithConcurrency(3),
			WithTimeout(time.Duration(20*time.Second)),
			WithDialTimeout(time.Duration(20*time.Second)),
			WithData(data),
			WithInsecure(true),
			WithName("server streaming test"),
		)

		assert.NoError(t, err)

		assert.NotNil(t, report)

		assert.Equal(t, 15, int(report.Count))
		assert.NotZero(t, report.Average)
		assert.NotZero(t, report.Fastest)
		assert.NotZero(t, report.Slowest)
		assert.NotZero(t, report.Rps)
		assert.Equal(t, "server streaming test", report.Name)
		assert.NotEmpty(t, report.Date)
		assert.NotEmpty(t, report.Options)
		assert.NotEmpty(t, report.Details)
		assert.Equal(t, true, report.Options.Insecure)
		assert.NotEmpty(t, report.LatencyDistribution)
		assert.Equal(t, ReasonNormalEnd, report.EndReason)
		assert.Empty(t, report.ErrorDist)

		assert.NotEqual(t, report.Average, report.Slowest)
		assert.NotEqual(t, report.Average, report.Fastest)
		assert.NotEqual(t, report.Slowest, report.Fastest)

		count := gs.GetCount(callType)
		assert.Equal(t, 15, count)

		connCount := gs.GetConnectionCount()
		assert.Equal(t, 1, connCount)
	})

	t.Run("with stream cancel", func(t *testing.T) {

		t.Skip("stream cancel returns errors normally")

		gs.ResetCounters()

		oldData := gs.StreamData

		nc := 100000
		gs.StreamData = make([]*helloworld.HelloReply, nc)
		for i := 0; i < nc; i++ {
			name := "name " + strconv.FormatInt(int64(i), 10)
			gs.StreamData[i] = &helloworld.HelloReply{Message: "Hello " + name}
		}

		data := make(map[string]interface{})
		data["name"] = "bob"

		report, err := Run(
			"helloworld.Greeter.SayHellos",
			internal.TestLocalhost,
			WithProtoFile("../testdata/greeter.proto", []string{}),
			WithTotalRequests(1),
			WithConcurrency(1),
			WithTimeout(time.Duration(20*time.Second)),
			WithDialTimeout(time.Duration(20*time.Second)),
			WithData(data),
			WithInsecure(true),
			WithName("server streaming test"),
			WithStreamCallDuration(200*time.Millisecond),
		)

		assert.NoError(t, err)

		assert.NotNil(t, report)

		assert.True(t, report.Total > 200*time.Millisecond, report.Total.String()+" not in interval")
		assert.Equal(t, 1, int(report.Count))
		assert.NotZero(t, report.Average)
		assert.NotZero(t, report.Fastest)
		assert.NotZero(t, report.Slowest)
		assert.NotZero(t, report.Rps)
		assert.Equal(t, "server streaming test", report.Name)
		assert.NotEmpty(t, report.Date)
		assert.NotEmpty(t, report.Options)
		assert.NotEmpty(t, report.Details)
		assert.Equal(t, true, report.Options.Insecure)
		assert.NotEmpty(t, report.LatencyDistribution)
		assert.Equal(t, ReasonNormalEnd, report.EndReason)
		assert.Empty(t, report.ErrorDist)

		assert.Equal(t, report.Average, report.Slowest)
		assert.Equal(t, report.Average, report.Fastest)
		assert.Equal(t, report.Slowest, report.Fastest)

		count := gs.GetCount(callType)
		assert.Equal(t, 1, count)

		connCount := gs.GetConnectionCount()
		assert.Equal(t, 1, connCount)

		assert.Len(t, report.Details, 1)
		dr := report.Details[0]
		assert.True(t, dr.Latency > 200*time.Millisecond && dr.Latency < 300*time.Millisecond, dr.Latency.String()+" not in interval")

		sends := gs.GetSendCounts(callType)
		assert.NotNil(t, sends)
		assert.Len(t, sends, 1)
		sendCount := sends[0]
		assert.True(t, sendCount < 20000, sendCount)

		// reset
		gs.StreamData = oldData
	})

	t.Run("with stream cancel count errors", func(t *testing.T) {

		gs.ResetCounters()

		oldData := gs.StreamData

		nc := 100000
		gs.StreamData = make([]*helloworld.HelloReply, nc)
		for i := 0; i < nc; i++ {
			name := "name " + strconv.FormatInt(int64(i), 10)
			gs.StreamData[i] = &helloworld.HelloReply{Message: "Hello " + name}
		}

		data := make(map[string]interface{})
		data["name"] = "bob"

		report, err := Run(
			"helloworld.Greeter.SayHellos",
			internal.TestLocalhost,
			WithProtoFile("../testdata/greeter.proto", []string{}),
			WithTotalRequests(1),
			WithConcurrency(1),
			WithTimeout(time.Duration(20*time.Second)),
			WithDialTimeout(time.Duration(20*time.Second)),
			WithData(data),
			WithInsecure(true),
			WithName("server streaming test"),
			WithStreamCallDuration(200*time.Millisecond),
			WithCountErrors(true),
		)

		assert.NoError(t, err)

		assert.NotNil(t, report)

		assert.True(t, report.Total > 200*time.Millisecond, report.Total.String()+" not in interval")
		assert.Equal(t, 1, int(report.Count))
		assert.NotZero(t, report.Average)
		assert.NotZero(t, report.Fastest)
		assert.NotZero(t, report.Slowest)
		assert.NotZero(t, report.Rps)
		assert.Equal(t, "server streaming test", report.Name)
		assert.NotEmpty(t, report.Date)
		assert.NotEmpty(t, report.Options)
		assert.NotEmpty(t, report.Details)
		assert.Equal(t, true, report.Options.Insecure)
		assert.NotEmpty(t, report.LatencyDistribution)
		assert.Equal(t, ReasonNormalEnd, report.EndReason)
		assert.Len(t, report.ErrorDist, 1)

		assert.Equal(t, report.Average, report.Slowest)
		assert.Equal(t, report.Average, report.Fastest)
		assert.Equal(t, report.Slowest, report.Fastest)

		count := gs.GetCount(callType)
		assert.Equal(t, 1, count)

		connCount := gs.GetConnectionCount()
		assert.Equal(t, 1, connCount)

		assert.Len(t, report.Details, 1)
		dr := report.Details[0]
		assert.True(t, dr.Latency > 200*time.Millisecond && dr.Latency < 300*time.Millisecond, dr.Latency.String()+" not in interval")

		sends := gs.GetSendCounts(callType)
		assert.NotNil(t, sends)
		assert.Len(t, sends, 1)
		sendCount := sends[0]
		assert.True(t, sendCount < 30000, sendCount)

		// reset
		gs.StreamData = oldData
	})

	t.Run("with stream call count and count errors", func(t *testing.T) {

		gs.ResetCounters()

		oldData := gs.StreamData

		nc := 5000
		gs.StreamData = make([]*helloworld.HelloReply, nc)
		for i := 0; i < nc; i++ {
			name := "name " + strconv.FormatInt(int64(i), 10)
			gs.StreamData[i] = &helloworld.HelloReply{Message: "Hello " + name}
		}

		data := make(map[string]interface{})
		data["name"] = "bob"

		report, err := Run(
			"helloworld.Greeter.SayHellos",
			internal.TestLocalhost,
			WithProtoFile("../testdata/greeter.proto", []string{}),
			WithTotalRequests(1),
			WithConcurrency(1),
			WithTimeout(time.Duration(20*time.Second)),
			WithDialTimeout(time.Duration(20*time.Second)),
			WithData(data),
			WithInsecure(true),
			WithName("server streaming test"),
			WithStreamCallCount(50),
			WithCountErrors(true),
		)

		assert.NoError(t, err)

		assert.NotNil(t, report)

		assert.Equal(t, 1, int(report.Count))
		assert.NotZero(t, report.Average)
		assert.NotZero(t, report.Fastest)
		assert.NotZero(t, report.Slowest)
		assert.NotZero(t, report.Rps)
		assert.Equal(t, "server streaming test", report.Name)
		assert.NotEmpty(t, report.Date)
		assert.NotEmpty(t, report.Options)
		assert.NotEmpty(t, report.Details)
		assert.Equal(t, true, report.Options.Insecure)
		assert.NotEmpty(t, report.LatencyDistribution)
		assert.Equal(t, ReasonNormalEnd, report.EndReason)
		assert.Len(t, report.ErrorDist, 1)

		assert.Equal(t, report.Average, report.Slowest)
		assert.Equal(t, report.Average, report.Fastest)
		assert.Equal(t, report.Slowest, report.Fastest)

		count := gs.GetCount(callType)
		assert.Equal(t, 1, count)

		connCount := gs.GetConnectionCount()
		assert.Equal(t, 1, connCount)

		sends := gs.GetSendCounts(callType)
		assert.NotNil(t, sends)
		assert.Len(t, sends, 1)
		sendCount := sends[0]
		assert.True(t, sendCount <= 3500, sendCount)

		// reset
		gs.StreamData = oldData
	})

	t.Run("with recv intercept", func(t *testing.T) {

		gs.ResetCounters()

		oldData := gs.StreamData

		nc := 10000
		gs.StreamData = make([]*helloworld.HelloReply, nc)
		for i := 0; i < nc; i++ {
			name := "name " + strconv.FormatInt(int64(i), 10)
			gs.StreamData[i] = &helloworld.HelloReply{Message: "Hello " + name}
		}

		data := make(map[string]interface{})
		data["name"] = "bob"

		report, err := Run(
			"helloworld.Greeter.SayHellos",
			internal.TestLocalhost,
			WithProtoFile("../testdata/greeter.proto", []string{}),
			WithTotalRequests(1),
			WithConcurrency(1),
			WithTimeout(time.Duration(20*time.Second)),
			WithDialTimeout(time.Duration(20*time.Second)),
			WithData(data),
			WithInsecure(true),
			WithName("server streaming test"),
			WithCountErrors(true),
			WithStreamRecvMsgIntercept(func(msg *dynamic.Message, err error) error {
				if err == nil {
					reply := &helloworld.HelloReply{}
					convertErr := msg.ConvertTo(reply)
					if convertErr == nil {
						if reply.GetMessage() == "Hello name 100" {
							return ErrEndStream
						}
					}
				}

				return nil
			}),
		)

		assert.NoError(t, err)

		assert.NotNil(t, report)

		assert.Equal(t, 1, int(report.Count))
		assert.NotZero(t, report.Average)
		assert.NotZero(t, report.Fastest)
		assert.NotZero(t, report.Slowest)
		assert.NotZero(t, report.Rps)
		assert.Equal(t, "server streaming test", report.Name)
		assert.NotEmpty(t, report.Date)
		assert.NotEmpty(t, report.Options)
		assert.NotEmpty(t, report.Details)
		assert.Equal(t, true, report.Options.Insecure)
		assert.NotEmpty(t, report.LatencyDistribution)
		assert.Equal(t, ReasonNormalEnd, report.EndReason)
		assert.Len(t, report.ErrorDist, 1)

		assert.Equal(t, report.Average, report.Slowest)
		assert.Equal(t, report.Average, report.Fastest)
		assert.Equal(t, report.Slowest, report.Fastest)

		count := gs.GetCount(callType)
		assert.Equal(t, 1, count)

		connCount := gs.GetConnectionCount()
		assert.Equal(t, 1, connCount)

		sends := gs.GetSendCounts(callType)
		assert.NotNil(t, sends)
		assert.Len(t, sends, 1)
		sendCount := sends[0]
		assert.True(t, sendCount <= 1500, sendCount)

		// reset
		gs.StreamData = oldData
	})
}

func TestRunClientStreaming(t *testing.T) {
	callType := helloworld.ClientStream

	gs, s, err := internal.StartServer(false)

	if err != nil {
		assert.FailNow(t, err.Error())
	}

	defer s.Stop()

	t.Run("basic", func(t *testing.T) {
		gs.ResetCounters()

		m1 := make(map[string]interface{})
		m1["name"] = "bob"

		m2 := make(map[string]interface{})
		m2["name"] = "Kate"

		m3 := make(map[string]interface{})
		m3["name"] = "foo"

		data := []interface{}{m1, m2, m3}

		report, err := Run(
			"helloworld.Greeter.SayHelloCS",
			internal.TestLocalhost,
			WithProtoFile("../testdata/greeter.proto", []string{}),
			WithTotalRequests(16),
			WithConcurrency(4),
			WithTimeout(time.Duration(20*time.Second)),
			WithDialTimeout(time.Duration(20*time.Second)),
			WithData(data),
			WithInsecure(true),
		)

		assert.NoError(t, err)

		assert.NotNil(t, report)

		assert.Equal(t, 16, int(report.Count))
		assert.NotZero(t, report.Average)
		assert.NotZero(t, report.Fastest)
		assert.NotZero(t, report.Slowest)
		assert.NotZero(t, report.Rps)
		assert.Empty(t, report.Name)
		assert.NotEmpty(t, report.Date)
		assert.NotEmpty(t, report.Details)
		assert.NotEmpty(t, report.Options)
		assert.Equal(t, true, report.Options.Insecure)
		assert.NotEmpty(t, report.LatencyDistribution)
		assert.Equal(t, ReasonNormalEnd, report.EndReason)
		assert.Empty(t, report.ErrorDist)

		assert.NotEqual(t, report.Average, report.Slowest)
		assert.NotEqual(t, report.Average, report.Fastest)
		assert.NotEqual(t, report.Slowest, report.Fastest)

		count := gs.GetCount(callType)
		assert.Equal(t, 16, count)

		connCount := gs.GetConnectionCount()
		assert.Equal(t, 1, connCount)
	})

	t.Run("with stream interval", func(t *testing.T) {
		gs.ResetCounters()

		m1 := make(map[string]interface{})
		m1["name"] = "bob"

		m2 := make(map[string]interface{})
		m2["name"] = "Kate"

		m3 := make(map[string]interface{})
		m3["name"] = "foo"

		m4 := make(map[string]interface{})
		m4["name"] = "bar"

		data := []interface{}{m1, m2, m3, m4}

		report, err := Run(
			"helloworld.Greeter.SayHelloCS",
			internal.TestLocalhost,
			WithProtoFile("../testdata/greeter.proto", []string{}),
			WithTotalRequests(1),
			WithConcurrency(1),
			WithTimeout(time.Duration(20*time.Second)),
			WithDialTimeout(time.Duration(20*time.Second)),
			WithStreamInterval(200*time.Millisecond),
			WithData(data),
			WithInsecure(true),
		)

		assert.NoError(t, err)

		assert.NotNil(t, report)

		assert.True(t, report.Total > 600*time.Millisecond, report.Total.String()+" not in interval")
		assert.Equal(t, 1, int(report.Count))
		assert.NotZero(t, report.Average)
		assert.NotZero(t, report.Fastest)
		assert.NotZero(t, report.Slowest)
		assert.NotZero(t, report.Rps)
		assert.Empty(t, report.Name)
		assert.NotEmpty(t, report.Date)
		assert.NotEmpty(t, report.Details)
		assert.NotEmpty(t, report.Options)
		assert.Equal(t, true, report.Options.Insecure)
		assert.NotEmpty(t, report.LatencyDistribution)
		assert.Equal(t, ReasonNormalEnd, report.EndReason)
		assert.Empty(t, report.ErrorDist)

		assert.Equal(t, report.Average, report.Slowest)
		assert.Equal(t, report.Average, report.Fastest)
		assert.Equal(t, report.Slowest, report.Fastest)

		count := gs.GetCount(callType)
		assert.Equal(t, 1, count)

		connCount := gs.GetConnectionCount()
		assert.Equal(t, 1, connCount)

		assert.Len(t, report.Details, 1)
		dr := report.Details[0]
		assert.True(t, dr.Latency > 600*time.Millisecond && dr.Latency < 800*time.Millisecond, dr.Latency.String()+" not in interval")
	})

	t.Run("with stream cancel", func(t *testing.T) {
		gs.ResetCounters()

		nc := 100000
		data := make([]interface{}, nc)
		for i := 0; i < nc; i++ {
			data[i] = map[string]interface{}{
				"name": "name " + strconv.FormatInt(int64(i), 10),
			}
		}

		report, err := Run(
			"helloworld.Greeter.SayHelloCS",
			internal.TestLocalhost,
			WithProtoFile("../testdata/greeter.proto", []string{}),
			WithTotalRequests(1),
			WithConcurrency(1),
			WithTimeout(time.Duration(20*time.Second)),
			WithDialTimeout(time.Duration(20*time.Second)),
			WithStreamCallDuration(200*time.Millisecond),
			WithData(data),
			WithInsecure(true),
		)

		assert.NoError(t, err)

		assert.NotNil(t, report)

		assert.True(t, report.Total > 200*time.Millisecond, report.Total.String()+" not in interval")
		assert.Equal(t, 1, int(report.Count))
		assert.NotZero(t, report.Average)
		assert.NotZero(t, report.Fastest)
		assert.NotZero(t, report.Slowest)
		assert.NotZero(t, report.Rps)
		assert.Empty(t, report.Name)
		assert.NotEmpty(t, report.Date)
		assert.NotEmpty(t, report.Details)
		assert.NotEmpty(t, report.Options)
		assert.Equal(t, true, report.Options.Insecure)
		assert.NotEmpty(t, report.LatencyDistribution)
		assert.Equal(t, ReasonNormalEnd, report.EndReason)
		assert.Empty(t, report.ErrorDist)

		assert.Equal(t, report.Average, report.Slowest)
		assert.Equal(t, report.Average, report.Fastest)
		assert.Equal(t, report.Slowest, report.Fastest)

		count := gs.GetCount(callType)
		assert.Equal(t, 1, count)

		connCount := gs.GetConnectionCount()
		assert.Equal(t, 1, connCount)

		assert.Len(t, report.Details, 1)
		dr := report.Details[0]
		assert.True(t, dr.Latency > 200*time.Millisecond && dr.Latency < 400*time.Millisecond, dr.Latency.String()+" not in interval")

		calls := gs.GetCalls(callType)
		assert.NotNil(t, calls)
		assert.Len(t, calls, 1)
		msgs := calls[0]
		assert.True(t, len(msgs) < 15000, len(msgs))
	})

	t.Run("with stream interval and cancel", func(t *testing.T) {
		gs.ResetCounters()

		m1 := make(map[string]interface{})
		m1["name"] = "bob"
		m2 := make(map[string]interface{})
		m2["name"] = "Kate"
		m3 := make(map[string]interface{})
		m3["name"] = "foo"
		m4 := make(map[string]interface{})
		m4["name"] = "bar"
		m5 := make(map[string]interface{})
		m5["name"] = "biz"

		data := []interface{}{m1, m2, m3, m4, m5}

		report, err := Run(
			"helloworld.Greeter.SayHelloCS",
			internal.TestLocalhost,
			WithProtoFile("../testdata/greeter.proto", []string{}),
			WithTotalRequests(1),
			WithConcurrency(1),
			WithTimeout(time.Duration(20*time.Second)),
			WithDialTimeout(time.Duration(20*time.Second)),
			WithStreamInterval(200*time.Millisecond),
			WithStreamCallDuration(650*time.Millisecond),
			WithData(data),
			WithInsecure(true),
		)

		assert.NoError(t, err)

		assert.NotNil(t, report)

		assert.True(t, report.Total > 650*time.Millisecond, report.Total.String()+" not in interval")
		assert.Equal(t, 1, int(report.Count))
		assert.NotZero(t, report.Average)
		assert.NotZero(t, report.Fastest)
		assert.NotZero(t, report.Slowest)
		assert.NotZero(t, report.Rps)
		assert.Empty(t, report.Name)
		assert.NotEmpty(t, report.Date)
		assert.NotEmpty(t, report.Details)
		assert.NotEmpty(t, report.Options)
		assert.Equal(t, true, report.Options.Insecure)
		assert.NotEmpty(t, report.LatencyDistribution)
		assert.Equal(t, ReasonNormalEnd, report.EndReason)
		assert.Empty(t, report.ErrorDist)

		assert.Equal(t, report.Average, report.Slowest)
		assert.Equal(t, report.Average, report.Fastest)
		assert.Equal(t, report.Slowest, report.Fastest)

		count := gs.GetCount(callType)
		assert.Equal(t, 1, count)

		connCount := gs.GetConnectionCount()
		assert.Equal(t, 1, connCount)

		assert.Len(t, report.Details, 1)
		dr := report.Details[0]
		assert.True(t, dr.Latency > 650*time.Millisecond && dr.Latency < 800*time.Millisecond, dr.Latency.String()+" not in interval")

		calls := gs.GetCalls(callType)
		assert.NotNil(t, calls)
		assert.Len(t, calls, 1)
		msgs := calls[0]
		assert.Len(t, msgs, 4)
	})

	t.Run("with stream interval and call count", func(t *testing.T) {
		gs.ResetCounters()

		m1 := make(map[string]interface{})
		m1["name"] = "bob"
		m2 := make(map[string]interface{})
		m2["name"] = "Kate"
		m3 := make(map[string]interface{})
		m3["name"] = "foo"
		m4 := make(map[string]interface{})
		m4["name"] = "bar"
		m5 := make(map[string]interface{})
		m5["name"] = "biz"

		data := []interface{}{m1, m2, m3, m4, m5}

		report, err := Run(
			"helloworld.Greeter.SayHelloCS",
			internal.TestLocalhost,
			WithProtoFile("../testdata/greeter.proto", []string{}),
			WithTotalRequests(1),
			WithConcurrency(1),
			WithTimeout(time.Duration(20*time.Second)),
			WithDialTimeout(time.Duration(20*time.Second)),
			WithStreamInterval(100*time.Millisecond),
			WithStreamCallCount(3),
			WithData(data),
			WithInsecure(true),
		)

		assert.NoError(t, err)

		assert.NotNil(t, report)

		assert.True(t, report.Total > 200*time.Millisecond, report.Total.String()+" not in interval")
		assert.Equal(t, 1, int(report.Count))
		assert.NotZero(t, report.Average)
		assert.NotZero(t, report.Fastest)
		assert.NotZero(t, report.Slowest)
		assert.NotZero(t, report.Rps)
		assert.Empty(t, report.Name)
		assert.NotEmpty(t, report.Date)
		assert.NotEmpty(t, report.Details)
		assert.NotEmpty(t, report.Options)
		assert.Equal(t, true, report.Options.Insecure)
		assert.NotEmpty(t, report.LatencyDistribution)
		assert.Equal(t, ReasonNormalEnd, report.EndReason)
		assert.Empty(t, report.ErrorDist)

		assert.Equal(t, report.Average, report.Slowest)
		assert.Equal(t, report.Average, report.Fastest)
		assert.Equal(t, report.Slowest, report.Fastest)

		count := gs.GetCount(callType)
		assert.Equal(t, 1, count)

		connCount := gs.GetConnectionCount()
		assert.Equal(t, 1, connCount)

		assert.Len(t, report.Details, 1)
		dr := report.Details[0]
		// allow extra time for last message send
		assert.True(t, dr.Latency > 200*time.Millisecond && dr.Latency < 400*time.Millisecond, dr.Latency.String()+" not in interval")

		calls := gs.GetCalls(callType)
		assert.NotNil(t, calls)
		assert.Len(t, calls, 1)
		msgs := calls[0]
		assert.Len(t, msgs, 3)
	})

	t.Run("with stream dynamic message true", func(t *testing.T) {
		gs.ResetCounters()

		data := make(map[string]interface{})
		data["name"] = "{{newUUID}}"

		report, err := Run(
			"helloworld.Greeter.SayHelloCS",
			internal.TestLocalhost,
			WithProtoFile("../testdata/greeter.proto", []string{}),
			WithTotalRequests(1),
			WithConcurrency(1),
			WithTimeout(time.Duration(20*time.Second)),
			WithDialTimeout(time.Duration(20*time.Second)),
			WithStreamInterval(100*time.Millisecond),
			WithStreamCallCount(4),
			WithStreamDynamicMessages(true),
			WithData(data),
			WithInsecure(true),
		)

		assert.NoError(t, err)

		assert.NotNil(t, report)

		assert.True(t, report.Total > 300*time.Millisecond, report.Total.String()+" not in interval")
		assert.Equal(t, 1, int(report.Count))
		assert.NotZero(t, report.Average)
		assert.NotZero(t, report.Fastest)
		assert.NotZero(t, report.Slowest)
		assert.NotZero(t, report.Rps)
		assert.Empty(t, report.Name)
		assert.NotEmpty(t, report.Date)
		assert.NotEmpty(t, report.Details)
		assert.NotEmpty(t, report.Options)
		assert.Equal(t, true, report.Options.Insecure)
		assert.NotEmpty(t, report.LatencyDistribution)
		assert.Equal(t, ReasonNormalEnd, report.EndReason)
		assert.Empty(t, report.ErrorDist)

		assert.Equal(t, report.Average, report.Slowest)
		assert.Equal(t, report.Average, report.Fastest)
		assert.Equal(t, report.Slowest, report.Fastest)

		count := gs.GetCount(callType)
		assert.Equal(t, 1, count)

		connCount := gs.GetConnectionCount()
		assert.Equal(t, 1, connCount)

		assert.Len(t, report.Details, 1)
		dr := report.Details[0]
		// allow 100ms for each dynamic message (re)generation
		assert.True(t, dr.Latency > 300*time.Millisecond && dr.Latency < 700*time.Millisecond, dr.Latency.String()+" not in interval")

		calls := gs.GetCalls(callType)
		assert.NotNil(t, calls)
		assert.Len(t, calls, 1)
		msgs := calls[0]
		assert.Len(t, msgs, 4)
		for i, m1 := range msgs {
			parsed, err := uuid.Parse(m1.GetName())
			assert.NoError(t, err)
			assert.NotEmpty(t, parsed)

			for j, m2 := range msgs {
				parsed, err = uuid.Parse(m2.GetName())
				assert.NoError(t, err)
				assert.NotEmpty(t, parsed)

				if i != j {
					assert.NotEqual(t, m1.GetName(), m2.GetName())
				}
			}
		}
	})

	t.Run("with stream dynamic message false", func(t *testing.T) {
		gs.ResetCounters()

		data := make(map[string]interface{})
		data["name"] = "{{newUUID}}"

		report, err := Run(
			"helloworld.Greeter.SayHelloCS",
			internal.TestLocalhost,
			WithProtoFile("../testdata/greeter.proto", []string{}),
			WithTotalRequests(1),
			WithConcurrency(1),
			WithTimeout(time.Duration(20*time.Second)),
			WithDialTimeout(time.Duration(20*time.Second)),
			WithStreamInterval(100*time.Millisecond),
			WithStreamCallCount(4),
			WithData(data),
			WithInsecure(true),
		)

		assert.NoError(t, err)

		assert.NotNil(t, report)

		assert.True(t, report.Total > 300*time.Millisecond, report.Total.String()+" not in interval")
		assert.Equal(t, 1, int(report.Count))
		assert.NotZero(t, report.Average)
		assert.NotZero(t, report.Fastest)
		assert.NotZero(t, report.Slowest)
		assert.NotZero(t, report.Rps)
		assert.Empty(t, report.Name)
		assert.NotEmpty(t, report.Date)
		assert.NotEmpty(t, report.Details)
		assert.NotEmpty(t, report.Options)
		assert.Equal(t, true, report.Options.Insecure)
		assert.NotEmpty(t, report.LatencyDistribution)
		assert.Equal(t, ReasonNormalEnd, report.EndReason)
		assert.Empty(t, report.ErrorDist)

		assert.Equal(t, report.Average, report.Slowest)
		assert.Equal(t, report.Average, report.Fastest)
		assert.Equal(t, report.Slowest, report.Fastest)

		count := gs.GetCount(callType)
		assert.Equal(t, 1, count)

		connCount := gs.GetConnectionCount()
		assert.Equal(t, 1, connCount)

		assert.Len(t, report.Details, 1)
		dr := report.Details[0]
		assert.True(t, dr.Latency > 300*time.Millisecond && dr.Latency < 400*time.Millisecond, dr.Latency.String()+" not in interval")

		calls := gs.GetCalls(callType)
		assert.NotNil(t, calls)
		assert.Len(t, calls, 1)
		msgs := calls[0]
		assert.Len(t, msgs, 4)
		for i, m1 := range msgs {
			parsed, err := uuid.Parse(m1.GetName())
			assert.NoError(t, err)
			assert.NotEmpty(t, parsed)

			for j, m2 := range msgs {
				parsed, err = uuid.Parse(m2.GetName())
				assert.NoError(t, err)
				assert.NotEmpty(t, parsed)

				if i != j {
					assert.Equal(t, m1.GetName(), m2.GetName())
				}
			}
		}
	})

	t.Run("with stream message provider", func(t *testing.T) {
		gs.ResetCounters()

		callCounter := 0

		report, err := Run(
			"helloworld.Greeter.SayHelloCS",
			internal.TestLocalhost,
			WithProtoFile("../testdata/greeter.proto", []string{}),
			WithTotalRequests(1),
			WithConcurrency(1),
			WithTimeout(time.Duration(20*time.Second)),
			WithDialTimeout(time.Duration(20*time.Second)),
			WithInsecure(true),
			WithStreamMessageProvider(func(cd *CallData) (*dynamic.Message, error) {
				protoMsg := &helloworld.HelloRequest{Name: cd.WorkerID + ": " + strconv.Itoa(callCounter)}
				dynamicMsg, err := dynamic.AsDynamicMessage(protoMsg)
				if err != nil {
					return nil, err
				}

				callCounter++

				if callCounter == 5 {
					err = ErrLastMessage
				}

				return dynamicMsg, err
			}),
		)

		assert.NoError(t, err)

		assert.NotNil(t, report)

		assert.NotZero(t, report.Total)
		assert.Equal(t, 1, int(report.Count))
		assert.NotZero(t, report.Average)
		assert.NotZero(t, report.Fastest)
		assert.NotZero(t, report.Slowest)
		assert.NotZero(t, report.Rps)
		assert.Empty(t, report.Name)
		assert.NotEmpty(t, report.Date)
		assert.NotEmpty(t, report.Details)
		assert.NotEmpty(t, report.Options)
		assert.Equal(t, true, report.Options.Insecure)
		assert.NotEmpty(t, report.LatencyDistribution)
		assert.Equal(t, ReasonNormalEnd, report.EndReason)
		assert.Empty(t, report.ErrorDist)

		assert.Equal(t, report.Average, report.Slowest)
		assert.Equal(t, report.Average, report.Fastest)
		assert.Equal(t, report.Slowest, report.Fastest)

		count := gs.GetCount(callType)
		assert.Equal(t, 1, count)

		connCount := gs.GetConnectionCount()
		assert.Equal(t, 1, connCount)

		calls := gs.GetCalls(callType)
		assert.NotNil(t, calls)
		assert.Len(t, calls, 1)
		msgs := calls[0]
		assert.Len(t, msgs, 5)

		assert.Equal(t, "g0c0: 0", msgs[0].GetName())
		assert.Equal(t, "g0c0: 4", msgs[4].GetName())
	})
}

func TestRunClientStreamingBinary(t *testing.T) {
	callType := helloworld.ClientStream

	gs, s, err := internal.StartServer(false)

	if err != nil {
		assert.FailNow(t, err.Error())
	}

	defer s.Stop()

	gs.ResetCounters()

	msg := &helloworld.HelloRequest{}
	msg.Name = "bob"

	binData, err := proto.Marshal(msg)
	assert.NoError(t, err)

	report, err := Run(
		"helloworld.Greeter.SayHelloCS",
		internal.TestLocalhost,
		WithProtoFile("../testdata/greeter.proto", []string{}),
		WithTotalRequests(24),
		WithConcurrency(4),
		WithTimeout(time.Duration(20*time.Second)),
		WithDialTimeout(time.Duration(20*time.Second)),
		WithBinaryData(binData),
		WithInsecure(true),
	)

	assert.NoError(t, err)

	assert.NotNil(t, report)

	assert.Equal(t, 24, int(report.Count))
	assert.NotZero(t, report.Average)
	assert.NotZero(t, report.Fastest)
	assert.NotZero(t, report.Slowest)
	assert.NotZero(t, report.Rps)
	assert.Empty(t, report.Name)
	assert.NotEmpty(t, report.Date)
	assert.NotEmpty(t, report.Details)
	assert.NotEmpty(t, report.Options)
	assert.Equal(t, true, report.Options.Insecure)
	assert.NotEmpty(t, report.LatencyDistribution)
	assert.Equal(t, ReasonNormalEnd, report.EndReason)
	assert.Empty(t, report.ErrorDist)

	assert.NotEqual(t, report.Average, report.Slowest)
	assert.NotEqual(t, report.Average, report.Fastest)
	assert.NotEqual(t, report.Slowest, report.Fastest)

	count := gs.GetCount(callType)
	assert.Equal(t, 24, count)

	connCount := gs.GetConnectionCount()
	assert.Equal(t, 1, connCount)
}

func TestRunBidi(t *testing.T) {
	callType := helloworld.Bidi

	gs, s, err := internal.StartServer(false)

	if err != nil {
		assert.FailNow(t, err.Error())
	}

	defer s.Stop()

	t.Run("basic", func(t *testing.T) {
		gs.ResetCounters()

		m1 := make(map[string]interface{})
		m1["name"] = "bob"

		m2 := make(map[string]interface{})
		m2["name"] = "Kate"

		m3 := make(map[string]interface{})
		m3["name"] = "foo"

		data := []interface{}{m1, m2, m3}

		report, err := Run(
			"helloworld.Greeter.SayHelloBidi",
			internal.TestLocalhost,
			WithProtoFile("../testdata/greeter.proto", []string{}),
			WithTotalRequests(20),
			WithConcurrency(4),
			WithTimeout(time.Duration(20*time.Second)),
			WithDialTimeout(time.Duration(20*time.Second)),
			WithData(data),
			WithInsecure(true),
		)

		assert.NoError(t, err)

		assert.NotNil(t, report)

		assert.Equal(t, 20, int(report.Count))
		assert.NotZero(t, report.Average)
		assert.NotZero(t, report.Fastest)
		assert.NotZero(t, report.Slowest)
		assert.NotZero(t, report.Rps)
		assert.Empty(t, report.Name)
		assert.NotEmpty(t, report.Date)
		assert.NotEmpty(t, report.Details)
		assert.NotEmpty(t, report.Options)
		assert.NotEmpty(t, report.LatencyDistribution)
		assert.Equal(t, ReasonNormalEnd, report.EndReason)
		assert.Equal(t, true, report.Options.Insecure)
		assert.Empty(t, report.ErrorDist)

		assert.NotEqual(t, report.Average, report.Slowest)
		assert.NotEqual(t, report.Average, report.Fastest)
		assert.NotEqual(t, report.Slowest, report.Fastest)

		count := gs.GetCount(callType)
		assert.Equal(t, 20, count)

		connCount := gs.GetConnectionCount()
		assert.Equal(t, 1, connCount)
	})

	t.Run("with stream interval", func(t *testing.T) {
		gs.ResetCounters()

		m1 := make(map[string]interface{})
		m1["name"] = "bob"

		m2 := make(map[string]interface{})
		m2["name"] = "Kate"

		m3 := make(map[string]interface{})
		m3["name"] = "foo"

		m4 := make(map[string]interface{})
		m4["name"] = "bar"

		data := []interface{}{m1, m2, m3, m4}

		report, err := Run(
			"helloworld.Greeter.SayHelloBidi",
			internal.TestLocalhost,
			WithProtoFile("../testdata/greeter.proto", []string{}),
			WithTotalRequests(1),
			WithConcurrency(1),
			WithTimeout(time.Duration(20*time.Second)),
			WithDialTimeout(time.Duration(20*time.Second)),
			WithStreamInterval(200*time.Millisecond),
			WithData(data),
			WithInsecure(true),
		)

		assert.NoError(t, err)

		assert.NotNil(t, report)

		assert.True(t, report.Total > 600*time.Millisecond, report.Total.String()+" not in interval")
		assert.Equal(t, 1, int(report.Count))
		assert.NotZero(t, report.Average)
		assert.NotZero(t, report.Fastest)
		assert.NotZero(t, report.Slowest)
		assert.NotZero(t, report.Rps)
		assert.Empty(t, report.Name)
		assert.NotEmpty(t, report.Date)
		assert.NotEmpty(t, report.Details)
		assert.NotEmpty(t, report.Options)
		assert.NotEmpty(t, report.LatencyDistribution)
		assert.Equal(t, ReasonNormalEnd, report.EndReason)
		assert.Equal(t, true, report.Options.Insecure)
		assert.Empty(t, report.ErrorDist)

		assert.Equal(t, report.Average, report.Slowest)
		assert.Equal(t, report.Average, report.Fastest)
		assert.Equal(t, report.Slowest, report.Fastest)

		assert.Len(t, report.Details, 1)
		dr := report.Details[0]
		// allow extra for the last send and receive
		assert.True(t, dr.Latency > 600*time.Millisecond && dr.Latency < 950*time.Millisecond, dr.Latency.String()+" not in interval")

		count := gs.GetCount(callType)
		assert.Equal(t, 1, count)

		connCount := gs.GetConnectionCount()
		assert.Equal(t, 1, connCount)
	})

	t.Run("with stream cancel", func(t *testing.T) {
		gs.ResetCounters()

		nc := 100000
		data := make([]interface{}, nc)
		for i := 0; i < nc; i++ {
			data[i] = map[string]interface{}{
				"name": "name " + strconv.FormatInt(int64(i), 10),
			}
		}

		report, err := Run(
			"helloworld.Greeter.SayHelloBidi",
			internal.TestLocalhost,
			WithProtoFile("../testdata/greeter.proto", []string{}),
			WithTotalRequests(1),
			WithConcurrency(1),
			WithTimeout(time.Duration(20*time.Second)),
			WithDialTimeout(time.Duration(20*time.Second)),
			WithStreamCallDuration(200*time.Millisecond),
			WithData(data),
			WithInsecure(true),
		)

		assert.NoError(t, err)

		assert.NotNil(t, report)

		assert.True(t, report.Total > 200*time.Millisecond, report.Total.String()+" not in interval")
		assert.Equal(t, 1, int(report.Count))
		assert.NotZero(t, report.Average)
		assert.NotZero(t, report.Fastest)
		assert.NotZero(t, report.Slowest)
		assert.NotZero(t, report.Rps)
		assert.Empty(t, report.Name)
		assert.NotEmpty(t, report.Date)
		assert.NotEmpty(t, report.Details)
		assert.NotEmpty(t, report.Options)
		assert.NotEmpty(t, report.LatencyDistribution)
		assert.Equal(t, ReasonNormalEnd, report.EndReason)
		assert.Equal(t, true, report.Options.Insecure)
		assert.Empty(t, report.ErrorDist)

		assert.Equal(t, report.Average, report.Slowest)
		assert.Equal(t, report.Average, report.Fastest)
		assert.Equal(t, report.Slowest, report.Fastest)

		count := gs.GetCount(callType)
		assert.Equal(t, 1, count)

		connCount := gs.GetConnectionCount()
		assert.Equal(t, 1, connCount)

		assert.Len(t, report.Details, 1)
		dr := report.Details[0]
		assert.True(t, dr.Latency > 200*time.Millisecond && dr.Latency < 1000*time.Millisecond, dr.Latency.String()+" not in interval")

		calls := gs.GetCalls(callType)
		assert.NotNil(t, calls)
		assert.Len(t, calls, 1)
		msgs := calls[0]
		assert.True(t, len(msgs) < 15000, len(msgs))
	})

	t.Run("with stream interval and cancel", func(t *testing.T) {
		gs.ResetCounters()

		m1 := make(map[string]interface{})
		m1["name"] = "bob"
		m2 := make(map[string]interface{})
		m2["name"] = "kate"
		m3 := make(map[string]interface{})
		m3["name"] = "foo"
		m4 := make(map[string]interface{})
		m4["name"] = "bar"
		m5 := make(map[string]interface{})
		m5["name"] = "biz"
		m6 := make(map[string]interface{})
		m6["name"] = "baz"
		m7 := make(map[string]interface{})
		m7["name"] = "alice"

		data := []interface{}{m1, m2, m3, m4, m5, m6, m7}

		report, err := Run(
			"helloworld.Greeter.SayHelloBidi",
			internal.TestLocalhost,
			WithProtoFile("../testdata/greeter.proto", []string{}),
			WithTotalRequests(1),
			WithConcurrency(1),
			WithTimeout(time.Duration(20*time.Second)),
			WithDialTimeout(time.Duration(20*time.Second)),
			WithStreamInterval(200*time.Millisecond),
			WithStreamCallDuration(620*time.Millisecond),
			WithData(data),
			WithInsecure(true),
		)

		assert.NoError(t, err)

		assert.NotNil(t, report)

		assert.True(t, report.Total > 620*time.Millisecond, report.Total.String()+" not in interval")
		assert.Equal(t, 1, int(report.Count))
		assert.NotZero(t, report.Average)
		assert.NotZero(t, report.Fastest)
		assert.NotZero(t, report.Slowest)
		assert.NotZero(t, report.Rps)
		assert.Empty(t, report.Name)
		assert.NotEmpty(t, report.Date)
		assert.NotEmpty(t, report.Details)
		assert.NotEmpty(t, report.Options)
		assert.NotEmpty(t, report.LatencyDistribution)
		assert.Equal(t, ReasonNormalEnd, report.EndReason)
		assert.Equal(t, true, report.Options.Insecure)
		assert.Empty(t, report.ErrorDist)

		assert.Equal(t, report.Average, report.Slowest)
		assert.Equal(t, report.Average, report.Fastest)
		assert.Equal(t, report.Slowest, report.Fastest)

		count := gs.GetCount(callType)
		assert.Equal(t, 1, count)

		connCount := gs.GetConnectionCount()
		assert.Equal(t, 1, connCount)

		assert.Len(t, report.Details, 1)
		dr := report.Details[0]
		assert.True(t, dr.Latency > 620*time.Millisecond && dr.Latency < 820*time.Millisecond, dr.Latency.String()+" not in interval")

		calls := gs.GetCalls(callType)
		assert.NotNil(t, calls)
		assert.Len(t, calls, 1)
		msgs := calls[0]
		assert.Len(t, msgs, 4)
	})

	t.Run("with stream count", func(t *testing.T) {
		gs.ResetCounters()

		m1 := make(map[string]interface{})
		m1["name"] = "bob"
		m2 := make(map[string]interface{})
		m2["name"] = "Kate"
		m3 := make(map[string]interface{})
		m3["name"] = "foo"
		m4 := make(map[string]interface{})
		m4["name"] = "bar"
		m5 := make(map[string]interface{})
		m5["name"] = "biz"

		data := []interface{}{m1, m2, m3, m4, m5}

		report, err := Run(
			"helloworld.Greeter.SayHelloBidi",
			internal.TestLocalhost,
			WithProtoFile("../testdata/greeter.proto", []string{}),
			WithTotalRequests(1),
			WithConcurrency(1),
			WithTimeout(time.Duration(20*time.Second)),
			WithDialTimeout(time.Duration(20*time.Second)),
			WithStreamInterval(100*time.Millisecond),
			WithStreamCallCount(3),
			WithData(data),
			WithInsecure(true),
		)

		assert.NoError(t, err)

		assert.NotNil(t, report)

		assert.True(t, report.Total > 200*time.Millisecond, report.Total.String()+" not in interval")
		assert.Equal(t, 1, int(report.Count))
		assert.NotZero(t, report.Average)
		assert.NotZero(t, report.Fastest)
		assert.NotZero(t, report.Slowest)
		assert.NotZero(t, report.Rps)
		assert.Empty(t, report.Name)
		assert.NotEmpty(t, report.Date)
		assert.NotEmpty(t, report.Details)
		assert.NotEmpty(t, report.Options)
		assert.NotEmpty(t, report.LatencyDistribution)
		assert.Equal(t, ReasonNormalEnd, report.EndReason)
		assert.Equal(t, true, report.Options.Insecure)
		assert.Empty(t, report.ErrorDist)

		assert.Equal(t, report.Average, report.Slowest)
		assert.Equal(t, report.Average, report.Fastest)
		assert.Equal(t, report.Slowest, report.Fastest)

		count := gs.GetCount(callType)
		assert.Equal(t, 1, count)

		connCount := gs.GetConnectionCount()
		assert.Equal(t, 1, connCount)

		assert.Len(t, report.Details, 1)
		dr := report.Details[0]
		assert.True(t, dr.Latency > 200*time.Millisecond && dr.Latency < 300*time.Millisecond, dr.Latency.String()+" not in interval")

		calls := gs.GetCalls(callType)
		assert.NotNil(t, calls)
		assert.Len(t, calls, 1)
		msgs := calls[0]
		assert.Len(t, msgs, 3)
	})

	t.Run("with stream dynamic message true", func(t *testing.T) {
		gs.ResetCounters()

		data := make(map[string]interface{})
		data["name"] = "{{newUUID}}"

		report, err := Run(
			"helloworld.Greeter.SayHelloBidi",
			internal.TestLocalhost,
			WithProtoFile("../testdata/greeter.proto", []string{}),
			WithTotalRequests(1),
			WithConcurrency(1),
			WithTimeout(time.Duration(20*time.Second)),
			WithDialTimeout(time.Duration(20*time.Second)),
			WithStreamInterval(100*time.Millisecond),
			WithStreamCallCount(4),
			WithStreamDynamicMessages(true),
			WithData(data),
			WithInsecure(true),
		)

		assert.NoError(t, err)

		assert.NotNil(t, report)

		assert.True(t, report.Total > 300*time.Millisecond, report.Total.String()+" not in interval")
		assert.Equal(t, 1, int(report.Count))
		assert.NotZero(t, report.Average)
		assert.NotZero(t, report.Fastest)
		assert.NotZero(t, report.Slowest)
		assert.NotZero(t, report.Rps)
		assert.Empty(t, report.Name)
		assert.NotEmpty(t, report.Date)
		assert.NotEmpty(t, report.Details)
		assert.NotEmpty(t, report.Options)
		assert.Equal(t, true, report.Options.Insecure)
		assert.NotEmpty(t, report.LatencyDistribution)
		assert.Equal(t, ReasonNormalEnd, report.EndReason)
		assert.Empty(t, report.ErrorDist)

		assert.Equal(t, report.Average, report.Slowest)
		assert.Equal(t, report.Average, report.Fastest)
		assert.Equal(t, report.Slowest, report.Fastest)

		count := gs.GetCount(callType)
		assert.Equal(t, 1, count)

		connCount := gs.GetConnectionCount()
		assert.Equal(t, 1, connCount)

		assert.Len(t, report.Details, 1)
		dr := report.Details[0]
		assert.True(t, dr.Latency > 300*time.Millisecond && dr.Latency < 400*time.Millisecond, dr.Latency.String()+" not in interval")

		calls := gs.GetCalls(callType)
		assert.NotNil(t, calls)
		assert.Len(t, calls, 1)
		msgs := calls[0]
		assert.Len(t, msgs, 4)
		for i, m1 := range msgs {
			parsed, err := uuid.Parse(m1.GetName())
			assert.NoError(t, err)
			assert.NotEmpty(t, parsed)

			for j, m2 := range msgs {
				parsed, err = uuid.Parse(m2.GetName())
				assert.NoError(t, err)
				assert.NotEmpty(t, parsed)

				if i != j {
					assert.NotEqual(t, m1.GetName(), m2.GetName())
				}
			}
		}
	})

	t.Run("with stream dynamic message false", func(t *testing.T) {
		gs.ResetCounters()

		data := make(map[string]interface{})
		data["name"] = "{{newUUID}}"

		report, err := Run(
			"helloworld.Greeter.SayHelloBidi",
			internal.TestLocalhost,
			WithProtoFile("../testdata/greeter.proto", []string{}),
			WithTotalRequests(1),
			WithConcurrency(1),
			WithTimeout(time.Duration(20*time.Second)),
			WithDialTimeout(time.Duration(20*time.Second)),
			WithStreamInterval(100*time.Millisecond),
			WithStreamCallCount(4),
			WithData(data),
			WithInsecure(true),
		)

		assert.NoError(t, err)

		assert.NotNil(t, report)

		assert.True(t, report.Total > 300*time.Millisecond, report.Total.String()+" not in interval")
		assert.Equal(t, 1, int(report.Count))
		assert.NotZero(t, report.Average)
		assert.NotZero(t, report.Fastest)
		assert.NotZero(t, report.Slowest)
		assert.NotZero(t, report.Rps)
		assert.Empty(t, report.Name)
		assert.NotEmpty(t, report.Date)
		assert.NotEmpty(t, report.Details)
		assert.NotEmpty(t, report.Options)
		assert.Equal(t, true, report.Options.Insecure)
		assert.NotEmpty(t, report.LatencyDistribution)
		assert.Equal(t, ReasonNormalEnd, report.EndReason)
		assert.Empty(t, report.ErrorDist)

		assert.Equal(t, report.Average, report.Slowest)
		assert.Equal(t, report.Average, report.Fastest)
		assert.Equal(t, report.Slowest, report.Fastest)

		count := gs.GetCount(callType)
		assert.Equal(t, 1, count)

		connCount := gs.GetConnectionCount()
		assert.Equal(t, 1, connCount)

		assert.Len(t, report.Details, 1)
		dr := report.Details[0]
		assert.True(t, dr.Latency > 300*time.Millisecond && dr.Latency < 400*time.Millisecond, dr.Latency.String()+" not in interval")

		calls := gs.GetCalls(callType)
		assert.NotNil(t, calls)
		assert.Len(t, calls, 1)
		msgs := calls[0]
		assert.Len(t, msgs, 4)
		for i, m1 := range msgs {
			parsed, err := uuid.Parse(m1.GetName())
			assert.NoError(t, err)
			assert.NotEmpty(t, parsed)

			for j, m2 := range msgs {
				parsed, err = uuid.Parse(m2.GetName())
				assert.NoError(t, err)
				assert.NotEmpty(t, parsed)

				if i != j {
					assert.Equal(t, m1.GetName(), m2.GetName())
				}
			}
		}
	})

	t.Run("with stream recv message", func(t *testing.T) {
		gs.ResetCounters()

		m1 := make(map[string]interface{})
		m1["name"] = "bob"
		m2 := make(map[string]interface{})
		m2["name"] = "Kate"
		m3 := make(map[string]interface{})
		m3["name"] = "foo"
		m4 := make(map[string]interface{})
		m4["name"] = "bar"
		m5 := make(map[string]interface{})
		m5["name"] = "biz"
		m6 := make(map[string]interface{})
		m6["name"] = "baz"

		data := []interface{}{m1, m2, m3, m4, m5, m6}

		report, err := Run(
			"helloworld.Greeter.SayHelloBidi",
			internal.TestLocalhost,
			WithProtoFile("../testdata/greeter.proto", []string{}),
			WithTotalRequests(1),
			WithConcurrency(1),
			WithTimeout(time.Duration(20*time.Second)),
			WithDialTimeout(time.Duration(20*time.Second)),
			WithStreamInterval(100*time.Millisecond),
			WithData(data),
			WithInsecure(true),
			WithStreamRecvMsgIntercept(func(msg *dynamic.Message, err error) error {
				if err == nil {
					reply := &helloworld.HelloReply{}
					convertErr := msg.ConvertTo(reply)
					if convertErr == nil {
						if reply.GetMessage() == "Hello bar" {
							return ErrEndStream
						}
					}
				}

				return nil
			}),
		)

		assert.NoError(t, err)

		assert.NotNil(t, report)

		assert.True(t, report.Total > 300*time.Millisecond)
		assert.Equal(t, 1, int(report.Count))
		assert.NotZero(t, report.Average)
		assert.NotZero(t, report.Fastest)
		assert.NotZero(t, report.Slowest)
		assert.NotZero(t, report.Rps)
		assert.Empty(t, report.Name)
		assert.NotEmpty(t, report.Date)
		assert.NotEmpty(t, report.Details)
		assert.NotEmpty(t, report.Options)
		assert.NotEmpty(t, report.LatencyDistribution)
		assert.Equal(t, ReasonNormalEnd, report.EndReason)
		assert.Equal(t, true, report.Options.Insecure)
		assert.Empty(t, report.ErrorDist)

		assert.Equal(t, report.Average, report.Slowest)
		assert.Equal(t, report.Average, report.Fastest)
		assert.Equal(t, report.Slowest, report.Fastest)

		count := gs.GetCount(callType)
		assert.Equal(t, 1, count)

		connCount := gs.GetConnectionCount()
		assert.Equal(t, 1, connCount)

		assert.Len(t, report.Details, 1)
		dr := report.Details[0]
		assert.True(t, dr.Latency > 300*time.Millisecond && dr.Latency < 400*time.Millisecond, dr.Latency.String()+" not in interval")

		calls := gs.GetCalls(callType)
		assert.NotNil(t, calls)
		assert.Len(t, calls, 1)
		msgs := calls[0]
		assert.Len(t, msgs, 4)
	})

	t.Run("with stream recv message no stop", func(t *testing.T) {

		gs.ResetCounters()

		m1 := make(map[string]interface{})
		m1["name"] = "bob"
		m2 := make(map[string]interface{})
		m2["name"] = "Kate"
		m3 := make(map[string]interface{})
		m3["name"] = "foo"
		m4 := make(map[string]interface{})
		m4["name"] = "bar"
		m5 := make(map[string]interface{})
		m5["name"] = "biz"
		m6 := make(map[string]interface{})
		m6["name"] = "baz"

		data := []interface{}{m1, m2, m3, m4, m5, m6}

		report, err := Run(
			"helloworld.Greeter.SayHelloBidi",
			internal.TestLocalhost,
			WithProtoFile("../testdata/greeter.proto", []string{}),
			WithTotalRequests(1),
			WithConcurrency(1),
			WithTimeout(time.Duration(20*time.Second)),
			WithDialTimeout(time.Duration(20*time.Second)),
			WithStreamInterval(100*time.Millisecond),
			WithData(data),
			WithInsecure(true),
			WithStreamRecvMsgIntercept(func(msg *dynamic.Message, err error) error {
				if err == nil {
					reply := &helloworld.HelloReply{}
					convertErr := msg.ConvertTo(reply)
					if convertErr == nil {
						// bat doesn't exist
						if reply.GetMessage() == "Hello bat" {
							return ErrEndStream
						}
					}
				}

				return nil
			}),
		)

		assert.NoError(t, err)

		assert.NotNil(t, report)

		assert.True(t, report.Total > 500*time.Millisecond)
		assert.Equal(t, 1, int(report.Count))
		assert.NotZero(t, report.Average)
		assert.NotZero(t, report.Fastest)
		assert.NotZero(t, report.Slowest)
		assert.NotZero(t, report.Rps)
		assert.Empty(t, report.Name)
		assert.NotEmpty(t, report.Date)
		assert.NotEmpty(t, report.Details)
		assert.NotEmpty(t, report.Options)
		assert.NotEmpty(t, report.LatencyDistribution)
		assert.Equal(t, ReasonNormalEnd, report.EndReason)
		assert.Equal(t, true, report.Options.Insecure)
		assert.Empty(t, report.ErrorDist)

		assert.Equal(t, report.Average, report.Slowest)
		assert.Equal(t, report.Average, report.Fastest)
		assert.Equal(t, report.Slowest, report.Fastest)

		count := gs.GetCount(callType)
		assert.Equal(t, 1, count)

		connCount := gs.GetConnectionCount()
		assert.Equal(t, 1, connCount)

		assert.Len(t, report.Details, 1)
		dr := report.Details[0]
		assert.True(t, dr.Latency > 500*time.Millisecond && dr.Latency < 600*time.Millisecond, dr.Latency.String()+" not in interval")

		calls := gs.GetCalls(callType)
		assert.NotNil(t, calls)
		assert.Len(t, calls, 1)
		msgs := calls[0]
		assert.Len(t, msgs, 6)
	})

	t.Run("with stream message provider", func(t *testing.T) {
		gs.ResetCounters()

		callCounter := 0

		report, err := Run(
			"helloworld.Greeter.SayHelloBidi",
			internal.TestLocalhost,
			WithProtoFile("../testdata/greeter.proto", []string{}),
			WithTotalRequests(1),
			WithConcurrency(1),
			WithTimeout(time.Duration(20*time.Second)),
			WithDialTimeout(time.Duration(20*time.Second)),
			WithInsecure(true),
			WithStreamMessageProvider(func(cd *CallData) (*dynamic.Message, error) {
				protoMsg := &helloworld.HelloRequest{Name: cd.WorkerID + ": " + strconv.Itoa(callCounter)}
				dynamicMsg, err := dynamic.AsDynamicMessage(protoMsg)
				if err != nil {
					return nil, err
				}

				callCounter++

				if callCounter == 7 {
					err = ErrLastMessage
				}

				return dynamicMsg, err
			}),
		)

		assert.NoError(t, err)

		assert.NotNil(t, report)

		assert.NotZero(t, report.Total)
		assert.Equal(t, 1, int(report.Count))
		assert.NotZero(t, report.Average)
		assert.NotZero(t, report.Fastest)
		assert.NotZero(t, report.Slowest)
		assert.NotZero(t, report.Rps)
		assert.Empty(t, report.Name)
		assert.NotEmpty(t, report.Date)
		assert.NotEmpty(t, report.Details)
		assert.NotEmpty(t, report.Options)
		assert.NotEmpty(t, report.LatencyDistribution)
		assert.Equal(t, ReasonNormalEnd, report.EndReason)
		assert.Equal(t, true, report.Options.Insecure)
		assert.Empty(t, report.ErrorDist)

		assert.Equal(t, report.Average, report.Slowest)
		assert.Equal(t, report.Average, report.Fastest)
		assert.Equal(t, report.Slowest, report.Fastest)

		count := gs.GetCount(callType)
		assert.Equal(t, 1, count)

		connCount := gs.GetConnectionCount()
		assert.Equal(t, 1, connCount)

		calls := gs.GetCalls(callType)
		assert.NotNil(t, calls)
		assert.Len(t, calls, 1)
		msgs := calls[0]
		assert.Len(t, msgs, 7)

		assert.Equal(t, "g0c0: 0", msgs[0].GetName())
		assert.Equal(t, "g0c0: 6", msgs[6].GetName())
	})
}

func TestRunUnarySecure(t *testing.T) {

	callType := helloworld.Unary

	gs, s, err := internal.StartServer(true)

	if err != nil {
		assert.FailNow(t, err.Error())
	}

	defer s.Stop()

	gs.ResetCounters()

	data := make(map[string]interface{})
	data["name"] = "bob"

	report, err := Run(
		"helloworld.Greeter.SayHello",
		internal.TestLocalhost,
		WithProtoFile("../testdata/greeter.proto", []string{}),
		WithTotalRequests(18),
		WithConcurrency(3),
		WithTimeout(time.Duration(20*time.Second)),
		WithDialTimeout(time.Duration(20*time.Second)),
		WithData(data),
		WithRootCertificate("../testdata/localhost.crt"),
	)

	assert.NoError(t, err)

	assert.NotNil(t, report)

	assert.Equal(t, 18, int(report.Count))
	assert.NotZero(t, report.Average)
	assert.NotZero(t, report.Fastest)
	assert.NotZero(t, report.Slowest)
	assert.NotZero(t, report.Rps)
	assert.Empty(t, report.Name)
	assert.NotEmpty(t, report.Date)
	assert.NotEmpty(t, report.Details)
	assert.NotEmpty(t, report.Options)
	assert.NotEmpty(t, report.LatencyDistribution)
	assert.Equal(t, ReasonNormalEnd, report.EndReason)
	assert.Equal(t, false, report.Options.Insecure)
	assert.Empty(t, report.ErrorDist)

	assert.NotEqual(t, report.Average, report.Slowest)
	assert.NotEqual(t, report.Average, report.Fastest)
	assert.NotEqual(t, report.Slowest, report.Fastest)

	count := gs.GetCount(callType)
	assert.Equal(t, 18, count)

	connCount := gs.GetConnectionCount()
	assert.Equal(t, 1, connCount)
}

func TestRunUnaryProtoset(t *testing.T) {
	callType := helloworld.Unary

	gs, s, err := internal.StartServer(false)

	if err != nil {
		assert.FailNow(t, err.Error())
	}

	defer s.Stop()

	gs.ResetCounters()

	data := make(map[string]interface{})
	data["name"] = "bob"

	report, err := Run(
		"helloworld.Greeter.SayHello",
		internal.TestLocalhost,
		WithProtoset("../testdata/bundle.protoset"),
		WithTotalRequests(21),
		WithConcurrency(3),
		WithTimeout(time.Duration(20*time.Second)),
		WithDialTimeout(time.Duration(20*time.Second)),
		WithData(data),
		WithInsecure(true),
		WithKeepalive(time.Duration(5*time.Minute)),
		WithMetadataFromFile("../testdata/metadata.json"),
	)

	assert.NoError(t, err)

	assert.NotNil(t, report)

	md := make(map[string]string)
	md["request-id"] = "{{.RequestNumber}}"

	assert.Equal(t, 21, int(report.Count))
	assert.NotZero(t, report.Average)
	assert.NotZero(t, report.Fastest)
	assert.NotZero(t, report.Slowest)
	assert.NotZero(t, report.Rps)
	assert.Empty(t, report.Name)
	assert.NotEmpty(t, report.Date)
	assert.NotEmpty(t, report.Details)
	assert.NotEmpty(t, report.Options)
	assert.Equal(t, md, *report.Options.Metadata)
	assert.NotEmpty(t, report.LatencyDistribution)
	assert.Equal(t, ReasonNormalEnd, report.EndReason)
	assert.Equal(t, true, report.Options.Insecure)
	assert.Empty(t, report.ErrorDist)

	assert.NotEqual(t, report.Average, report.Slowest)
	assert.NotEqual(t, report.Average, report.Fastest)
	assert.NotEqual(t, report.Slowest, report.Fastest)

	count := gs.GetCount(callType)
	assert.Equal(t, 21, count)

	connCount := gs.GetConnectionCount()
	assert.Equal(t, 1, connCount)
}

func TestRunUnaryReflection(t *testing.T) {

	t.Run("Unknown method", func(t *testing.T) {

		gs, s, err := internal.StartServer(false)

		if err != nil {
			assert.FailNow(t, err.Error())
		}

		defer s.Stop()

		gs.ResetCounters()

		data := make(map[string]interface{})
		data["name"] = "bob"

		report, err := Run(
			"helloworld.Greeter.SayHelloAsdf",
			internal.TestLocalhost,
			WithTotalRequests(21),
			WithConcurrency(3),
			WithTimeout(time.Duration(20*time.Second)),
			WithDialTimeout(time.Duration(20*time.Second)),
			WithData(data),
			WithInsecure(true),
			WithKeepalive(time.Duration(1*time.Minute)),
			WithMetadataFromFile("../testdata/metadata.json"),
		)

		assert.Error(t, err)
		assert.Nil(t, report)

		connCount := gs.GetConnectionCount()
		assert.Equal(t, 1, connCount)
	})

	t.Run("Unary streaming", func(t *testing.T) {
		callType := helloworld.Unary

		gs, s, err := internal.StartServer(false)

		if err != nil {
			assert.FailNow(t, err.Error())
		}

		defer s.Stop()

		gs.ResetCounters()

		data := make(map[string]interface{})
		data["name"] = "bob"

		report, err := Run(
			"helloworld.Greeter.SayHello",
			internal.TestLocalhost,
			WithTotalRequests(21),
			WithConcurrency(3),
			WithTimeout(time.Duration(20*time.Second)),
			WithDialTimeout(time.Duration(20*time.Second)),
			WithData(data),
			WithInsecure(true),
			WithKeepalive(time.Duration(1*time.Minute)),
			WithMetadataFromFile("../testdata/metadata.json"),
		)

		assert.NoError(t, err)
		if err != nil {
			assert.FailNow(t, err.Error())
		}

		assert.NotNil(t, report)

		md := make(map[string]string)
		md["request-id"] = "{{.RequestNumber}}"

		assert.Equal(t, 21, int(report.Count))
		assert.NotZero(t, report.Average)
		assert.NotZero(t, report.Fastest)
		assert.NotZero(t, report.Slowest)
		assert.NotZero(t, report.Rps)
		assert.Empty(t, report.Name)
		assert.NotEmpty(t, report.Date)
		assert.NotEmpty(t, report.Details)
		assert.NotEmpty(t, report.Options)
		assert.Equal(t, md, *report.Options.Metadata)
		assert.NotEmpty(t, report.LatencyDistribution)
		assert.Equal(t, ReasonNormalEnd, report.EndReason)
		assert.Equal(t, true, report.Options.Insecure)
		assert.Empty(t, report.ErrorDist)

		assert.NotEqual(t, report.Average, report.Slowest)
		assert.NotEqual(t, report.Average, report.Fastest)
		assert.NotEqual(t, report.Slowest, report.Fastest)

		count := gs.GetCount(callType)
		assert.Equal(t, 21, count)

		connCount := gs.GetConnectionCount()
		assert.Equal(t, 2, connCount) // 1 extra connection for reflection
	})

	t.Run("Client streaming", func(t *testing.T) {
		callType := helloworld.ClientStream

		gs, s, err := internal.StartServer(false)

		if err != nil {
			assert.FailNow(t, err.Error())
		}

		defer s.Stop()

		gs.ResetCounters()

		m1 := make(map[string]interface{})
		m1["name"] = "bob"

		m2 := make(map[string]interface{})
		m2["name"] = "Kate"

		m3 := make(map[string]interface{})
		m3["name"] = "foo"

		data := []interface{}{m1, m2, m3}

		report, err := Run(
			"helloworld.Greeter.SayHelloCS",
			internal.TestLocalhost,
			WithTotalRequests(16),
			WithConcurrency(4),
			WithTimeout(time.Duration(20*time.Second)),
			WithDialTimeout(time.Duration(20*time.Second)),
			WithData(data),
			WithInsecure(true),
		)

		assert.NoError(t, err)

		assert.NotNil(t, report)

		assert.Equal(t, 16, int(report.Count))
		assert.NotZero(t, report.Average)
		assert.NotZero(t, report.Fastest)
		assert.NotZero(t, report.Slowest)
		assert.NotZero(t, report.Rps)
		assert.Empty(t, report.Name)
		assert.NotEmpty(t, report.Date)
		assert.NotEmpty(t, report.Details)
		assert.NotEmpty(t, report.Options)
		assert.Equal(t, true, report.Options.Insecure)
		assert.NotEmpty(t, report.LatencyDistribution)
		assert.Equal(t, ReasonNormalEnd, report.EndReason)
		assert.Empty(t, report.ErrorDist)

		assert.NotEqual(t, report.Average, report.Slowest)
		assert.NotEqual(t, report.Average, report.Fastest)
		assert.NotEqual(t, report.Slowest, report.Fastest)

		count := gs.GetCount(callType)
		assert.Equal(t, 16, count)

		connCount := gs.GetConnectionCount()
		assert.Equal(t, 2, connCount) // 1 extra connection for reflection
	})
}

func TestRunUnaryDurationStop(t *testing.T) {

	_, s, err := internal.StartSleepServer(false)

	if err != nil {
		assert.FailNow(t, err.Error())
	}

	defer s.Stop()

	t.Run("test close", func(t *testing.T) {

		data := make(map[string]interface{})
		data["Milliseconds"] = "150"

		report, err := Run(
			"main.SleepService.SleepFor",
			internal.TestLocalhost,
			WithProtoFile("../testdata/sleep.proto", []string{}),
			WithConnections(1),
			WithConcurrency(1),
			WithData(data),
			WithRunDuration(time.Duration(1*time.Second)),
			WithDurationStopAction("close"),
			WithTimeout(time.Duration(200*time.Millisecond)),
			WithInsecure(true),
		)

		assert.NoError(t, err)

		assert.NotNil(t, report)

		assert.Equal(t, 7, int(report.Count))
		assert.NotZero(t, report.Average)
		assert.NotZero(t, report.Fastest)
		assert.NotZero(t, report.Slowest)
		assert.NotZero(t, report.Rps)
		assert.Empty(t, report.Name)
		assert.NotEmpty(t, report.Date)
		assert.NotEmpty(t, report.Options)
		assert.NotEmpty(t, report.Details)
		assert.Equal(t, true, report.Options.Insecure)
		assert.NotEmpty(t, report.LatencyDistribution)
		assert.Equal(t, ReasonTimeout, report.EndReason)
		assert.Len(t, report.ErrorDist, 1)
		assert.Len(t, report.StatusCodeDist, 2)
		assert.Equal(t, 6, report.StatusCodeDist["OK"])
		assert.Equal(t, 1, report.StatusCodeDist["Unavailable"])
	})

	t.Run("test wait", func(t *testing.T) {

		data := make(map[string]interface{})
		data["Milliseconds"] = "150"

		report, err := Run(
			"main.SleepService.SleepFor",
			internal.TestLocalhost,
			WithProtoFile("../testdata/sleep.proto", []string{}),
			WithConnections(1),
			WithConcurrency(1),
			WithData(data),
			WithRunDuration(time.Duration(1*time.Second)),
			WithDurationStopAction("wait"),
			WithTimeout(time.Duration(200*time.Millisecond)),
			WithInsecure(true),
		)

		assert.NoError(t, err)

		assert.NotNil(t, report)

		assert.Equal(t, 7, int(report.Count))
		assert.NotZero(t, report.Average)
		assert.NotZero(t, report.Fastest)
		assert.NotZero(t, report.Slowest)
		assert.NotZero(t, report.Rps)
		assert.Empty(t, report.Name)
		assert.NotEmpty(t, report.Date)
		assert.NotEmpty(t, report.Options)
		assert.NotEmpty(t, report.Details)
		assert.Equal(t, true, report.Options.Insecure)
		assert.NotEmpty(t, report.LatencyDistribution)
		assert.Equal(t, ReasonTimeout, report.EndReason)
		assert.Len(t, report.ErrorDist, 0)
		assert.Len(t, report.StatusCodeDist, 1)
		assert.Equal(t, 7, report.StatusCodeDist["OK"])
	})

	t.Run("test ignore", func(t *testing.T) {

		data := make(map[string]interface{})
		data["Milliseconds"] = "150"

		report, err := Run(
			"main.SleepService.SleepFor",
			internal.TestLocalhost,
			WithProtoFile("../testdata/sleep.proto", []string{}),
			WithConnections(1),
			WithConcurrency(1),
			WithData(data),
			WithRunDuration(time.Duration(1*time.Second)),
			WithDurationStopAction("ignore"),
			WithTimeout(time.Duration(200*time.Millisecond)),
			WithInsecure(true),
		)

		assert.NoError(t, err)

		assert.NotNil(t, report)

		assert.Equal(t, 6, int(report.Count))
		assert.NotZero(t, report.Average)
		assert.NotZero(t, report.Fastest)
		assert.NotZero(t, report.Slowest)
		assert.NotZero(t, report.Rps)
		assert.Empty(t, report.Name)
		assert.NotEmpty(t, report.Date)
		assert.NotEmpty(t, report.Options)
		assert.NotEmpty(t, report.Details)
		assert.Equal(t, true, report.Options.Insecure)
		assert.NotEmpty(t, report.LatencyDistribution)
		assert.Equal(t, ReasonTimeout, report.EndReason)
		assert.Len(t, report.ErrorDist, 0)
		assert.Len(t, report.StatusCodeDist, 1)
		assert.Equal(t, 6, report.StatusCodeDist["OK"])
	})
}

func TestRunWrappedUnary(t *testing.T) {

	_, s, err := internal.StartWrappedServer(false)

	if err != nil {
		assert.FailNow(t, err.Error())
	}

	defer s.Stop()

	t.Run("json string data", func(t *testing.T) {
		report, err := Run(
			"wrapped.WrappedService.GetMessage",
			internal.TestLocalhost,
			WithProtoFile("../testdata/wrapped.proto", []string{"../testdata"}),
			WithTotalRequests(1),
			WithConcurrency(1),
			WithTimeout(time.Duration(20*time.Second)),
			WithDialTimeout(time.Duration(20*time.Second)),
			WithDataFromJSON(`"foo"`),
			WithInsecure(true),
		)

		assert.NoError(t, err)

		assert.NotNil(t, report)

		assert.Equal(t, 1, int(report.Count))
		assert.NotZero(t, report.Average)
		assert.NotZero(t, report.Fastest)
		assert.NotZero(t, report.Slowest)
		assert.NotZero(t, report.Rps)
		assert.Empty(t, report.Name)
		assert.NotEmpty(t, report.Date)
		assert.NotEmpty(t, report.Options)
		assert.NotEmpty(t, report.Details)
		assert.Equal(t, true, report.Options.Insecure)
		assert.NotEmpty(t, report.LatencyDistribution)
		assert.Equal(t, ReasonNormalEnd, report.EndReason)
		assert.Empty(t, report.ErrorDist)

		assert.Equal(t, report.Average, report.Slowest)
		assert.Equal(t, report.Average, report.Fastest)
	})

	t.Run("wrapped bin data with template chars", func(t *testing.T) {
		protoMsg := wrappers.BytesValue{Value: []byte{'\x7b', '\x7b', '\xc2', '\x7d', '\x7d'}}
		msg, err := proto.Marshal(&protoMsg)
		assert.NoError(t, err)

		report, err := Run(
			"wrapped.WrappedService.GetBytesMessage",
			internal.TestLocalhost,
			WithProtoFile("../testdata/wrapped.proto", []string{"../testdata"}),
			WithTotalRequests(1),
			WithConcurrency(1),
			WithTimeout(time.Duration(20*time.Second)),
			WithDialTimeout(time.Duration(20*time.Second)),
			WithBinaryData(msg),
			WithInsecure(true),
		)

		assert.NoError(t, err)

		assert.NotNil(t, report)

		assert.Equal(t, 1, int(report.Count))
	})

	t.Run("json string data from file", func(t *testing.T) {
		report, err := Run(
			"wrapped.WrappedService.GetMessage",
			internal.TestLocalhost,
			WithProtoFile("../testdata/wrapped.proto", []string{"../testdata"}),
			WithTotalRequests(1),
			WithConcurrency(1),
			WithTimeout(time.Duration(20*time.Second)),
			WithDialTimeout(time.Duration(20*time.Second)),
			WithDataFromFile(`../testdata/wrapped_data.json`),
			WithInsecure(true),
		)

		assert.NoError(t, err)

		assert.NotNil(t, report)

		assert.Equal(t, 1, int(report.Count))
		assert.NotZero(t, report.Average)
		assert.NotZero(t, report.Fastest)
		assert.NotZero(t, report.Slowest)
		assert.NotZero(t, report.Rps)
		assert.Empty(t, report.Name)
		assert.NotEmpty(t, report.Date)
		assert.NotEmpty(t, report.Options)
		assert.NotEmpty(t, report.Details)
		assert.Equal(t, true, report.Options.Insecure)
		assert.NotEmpty(t, report.LatencyDistribution)
		assert.Equal(t, ReasonNormalEnd, report.EndReason)
		assert.Empty(t, report.ErrorDist)

		assert.Equal(t, report.Average, report.Slowest)
		assert.Equal(t, report.Average, report.Fastest)
	})

	t.Run("json string data from empty file", func(t *testing.T) {
		report, err := Run(
			"wrapped.WrappedService.GetMessage",
			internal.TestLocalhost,
			WithProtoFile("../testdata/wrapped.proto", []string{"../testdata"}),
			WithDataFromFile(`../testdata/data_empty.json`),
			WithInsecure(true),
		)

		assert.Error(t, err)

		assert.NotNil(t, report)

		assert.Equal(t, 0, int(report.Count))
		assert.Zero(t, report.Average)
		assert.Zero(t, report.Fastest)
		assert.Zero(t, report.Slowest)
		assert.Zero(t, report.Rps)
		assert.Empty(t, report.Name)
		assert.NotEmpty(t, report.Date)
		assert.NotEmpty(t, report.Options)
		assert.Empty(t, report.Details)
		assert.Equal(t, true, report.Options.Insecure)
		assert.Empty(t, report.LatencyDistribution)
		assert.Equal(t, ReasonNormalEnd, report.EndReason)
		assert.Empty(t, report.ErrorDist)

		assert.Equal(t, report.Average, report.Slowest)
		assert.Equal(t, report.Average, report.Fastest)
	})

}

func TestRunGtimeUnary(t *testing.T) {

	ts, s, err := internal.StartTimeServer(false)

	if err != nil {
		assert.FailNow(t, err.Error())
	}

	defer s.Stop()

	t.Run("json string data", func(t *testing.T) {
		data := `{"ts":"2020-01-20T01:30:30.01Z", "dur":"30s", "user_id":"12"}`

		report, err := Run(
			"gtime.TimeService.TestCall",
			internal.TestLocalhost,
			WithProtoFile("../testdata/gtime.proto", []string{"../testdata"}),
			WithTotalRequests(1),
			WithConcurrency(1),
			WithTimeout(time.Duration(20*time.Second)),
			WithDialTimeout(time.Duration(20*time.Second)),
			WithDataFromJSON(data),
			WithInsecure(true),
		)

		assert.NoError(t, err)

		assert.NotNil(t, report)

		assert.Equal(t, 1, int(report.Count))
		assert.NotZero(t, report.Average)
		assert.NotZero(t, report.Fastest)
		assert.NotZero(t, report.Slowest)
		assert.NotZero(t, report.Rps)
		assert.Empty(t, report.Name)
		assert.NotEmpty(t, report.Date)
		assert.NotEmpty(t, report.Options)
		assert.NotEmpty(t, report.Details)
		assert.Equal(t, true, report.Options.Insecure)
		assert.NotEmpty(t, report.LatencyDistribution)
		assert.Equal(t, ReasonNormalEnd, report.EndReason)
		assert.Empty(t, report.ErrorDist)

		assert.Equal(t, report.Average, report.Slowest)
		assert.Equal(t, report.Average, report.Fastest)

		expectedTime, err := time.Parse(time.RFC3339, "2020-01-20T01:30:30.01Z")
		assert.NoError(t, err)

		expectedDur := time.Duration(30 * time.Second)

		assert.Equal(t, expectedTime, ts.LastTimestamp)
		assert.Equal(t, expectedDur, ts.LastDuration)
	})

	t.Run("json file data", func(t *testing.T) {

		report, err := Run(
			"",
			internal.TestLocalhost,
			WithConfigFromFile("../testdata/tsconfig.json"),
		)

		assert.NoError(t, err)
		assert.NotNil(t, report)

		assert.Equal(t, 1, int(report.Count))
		assert.NotZero(t, report.Average)
		assert.NotZero(t, report.Fastest)
		assert.NotZero(t, report.Slowest)
		assert.NotZero(t, report.Rps)
		assert.Equal(t, "TimeService - TestCall", report.Name)
		assert.NotEmpty(t, report.Date)
		assert.NotEmpty(t, report.Options)
		assert.NotEmpty(t, report.Details)
		assert.Equal(t, true, report.Options.Insecure)
		assert.NotEmpty(t, report.LatencyDistribution)
		assert.Equal(t, ReasonNormalEnd, report.EndReason)
		assert.Empty(t, report.ErrorDist)

		assert.Equal(t, report.Average, report.Slowest)
		assert.Equal(t, report.Average, report.Fastest)

		expectedTime, err := time.Parse(time.RFC3339, "2020-01-22T02:30:30.01Z")
		assert.NoError(t, err)

		expectedDur := time.Duration(40 * time.Second)

		assert.Equal(t, expectedTime, ts.LastTimestamp)
		assert.Equal(t, expectedDur, ts.LastDuration)
	})
}
