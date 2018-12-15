package runner

import (
	"sync"
	"testing"
	"time"

	"github.com/bojand/ghz/internal/helloworld"
	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"
)

func TestRunUnary(t *testing.T) {
	callType := helloworld.Unary

	gs, s, err := startServer(false)

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
			localhost,
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

	t.Run("test N and Name", func(t *testing.T) {
		gs.ResetCounters()

		data := make(map[string]interface{})
		data["name"] = "bob"

		report, err := Run(
			"helloworld.Greeter.SayHello",
			localhost,
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
	})

	t.Run("test QPS", func(t *testing.T) {
		gs.ResetCounters()

		var wg sync.WaitGroup

		wg.Add(1)

		time.AfterFunc(time.Duration(1500*time.Millisecond), func() {
			count := gs.GetCount(callType)
			assert.Equal(t, 2, count)
		})

		go func() {
			data := make(map[string]interface{})
			data["name"] = "bob"

			report, err := Run(
				"helloworld.Greeter.SayHello",
				localhost,
				WithProtoFile("../testdata/greeter.proto", []string{}),
				WithTotalRequests(10),
				WithConcurrency(2),
				WithQPS(1),
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

			wg.Done()

		}()
		wg.Wait()
	})

	t.Run("test binary", func(t *testing.T) {
		gs.ResetCounters()

		msg := &helloworld.HelloRequest{}
		msg.Name = "bob"

		binData, err := proto.Marshal(msg)

		report, err := Run(
			"helloworld.Greeter.SayHello",
			localhost,
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
	})
}

func TestRunServerStreaming(t *testing.T) {
	callType := helloworld.ServerStream

	gs, s, err := startServer(false)

	if err != nil {
		assert.FailNow(t, err.Error())
	}

	defer s.Stop()

	gs.ResetCounters()

	data := make(map[string]interface{})
	data["name"] = "bob"

	report, err := Run(
		"helloworld.Greeter.SayHellos",
		localhost,
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
}

func TestRunClientStreaming(t *testing.T) {
	callType := helloworld.ClientStream

	gs, s, err := startServer(false)

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
		localhost,
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
}

func TestRunClientStreamingBinary(t *testing.T) {
	callType := helloworld.ClientStream

	gs, s, err := startServer(false)

	if err != nil {
		assert.FailNow(t, err.Error())
	}

	defer s.Stop()

	gs.ResetCounters()

	msg := &helloworld.HelloRequest{}
	msg.Name = "bob"

	binData, err := proto.Marshal(msg)

	report, err := Run(
		"helloworld.Greeter.SayHelloCS",
		localhost,
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
}

func TestRunBidi(t *testing.T) {
	callType := helloworld.Bidi

	gs, s, err := startServer(false)

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
		"helloworld.Greeter.SayHelloBidi",
		localhost,
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
}

func TestRunUnarySecure(t *testing.T) {
	callType := helloworld.Unary

	gs, s, err := startServer(true)

	if err != nil {
		assert.FailNow(t, err.Error())
	}

	defer s.Stop()

	gs.ResetCounters()

	data := make(map[string]interface{})
	data["name"] = "bob"

	report, err := Run(
		"helloworld.Greeter.SayHello",
		localhost,
		WithProtoFile("../testdata/greeter.proto", []string{}),
		WithTotalRequests(18),
		WithConcurrency(3),
		WithTimeout(time.Duration(20*time.Second)),
		WithDialTimeout(time.Duration(20*time.Second)),
		WithData(data),
		WithCertificate("../testdata/localhost.crt", ""),
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
}

func TestRunUnaryProtoset(t *testing.T) {
	callType := helloworld.Unary

	gs, s, err := startServer(false)

	if err != nil {
		assert.FailNow(t, err.Error())
	}

	defer s.Stop()

	gs.ResetCounters()

	data := make(map[string]interface{})
	data["name"] = "bob"

	report, err := Run(
		"helloworld.Greeter.SayHello",
		localhost,
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
}
