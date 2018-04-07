package grpcannon

import (
	"net"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/bojand/grpcannon/internal/helloworld"
	"github.com/bojand/grpcannon/protodesc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const port = ":50051"
const localhost = "localhost:50051"

func startServer(secure bool) (*helloworld.Greeter, *grpc.Server, error) {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		return nil, nil, err
	}

	var opts []grpc.ServerOption

	if secure {
		creds, err := credentials.NewServerTLSFromFile("./testdata/localhost.crt", "./testdata/localhost.key")
		if err != nil {
			return nil, nil, err
		}
		opts = []grpc.ServerOption{grpc.Creds(creds)}
	}

	s := grpc.NewServer(opts...)

	gs := helloworld.NewGreeter()
	helloworld.RegisterGreeterServer(s, gs)
	// reflection.Register(s)
	go func() {
		s.Serve(lis)
	}()
	return gs, s, err
}

func TestRequesterUnary(t *testing.T) {
	callType := helloworld.Unary

	gs, s, err := startServer(false)

	if err != nil {
		assert.FailNow(t, err.Error())
	}

	defer s.Stop()

	md, err := protodesc.GetMethodDesc("helloworld.Greeter.SayHello", "./testdata/greeter.proto", []string{})

	data := make(map[string]interface{})
	data["name"] = "bob"

	t.Run("test report", func(t *testing.T) {
		gs.ResetCounters()

		reqr, err := New(md, &Options{
			Host:        localhost,
			N:           1,
			C:           1,
			Timeout:     20,
			DialTimtout: 20,
			Data:        data,
		})
		assert.NoError(t, err)

		report, err := reqr.Run()
		assert.NoError(t, err)
		assert.NotNil(t, report)
		assert.Equal(t, 1, int(report.Count))
		assert.Len(t, report.ErrorDist, 0)

		count := gs.GetCount(callType)
		assert.Equal(t, 1, count)
	})

	t.Run("test N", func(t *testing.T) {
		gs.ResetCounters()

		reqr, err := New(md, &Options{
			Host:        localhost,
			N:           12,
			C:           2,
			Timeout:     20,
			DialTimtout: 20,
			Data:        data,
		})
		assert.NoError(t, err)

		report, err := reqr.Run()
		assert.NoError(t, err)
		assert.NotNil(t, report)
		assert.Equal(t, 12, int(report.Count))
		assert.Len(t, report.ErrorDist, 0)

		count := gs.GetCount(callType)
		assert.Equal(t, 12, count)
	})

	t.Run("test QPS", func(t *testing.T) {
		gs.ResetCounters()

		var wg sync.WaitGroup

		reqr, err := New(md, &Options{
			Host:        localhost,
			N:           10,
			C:           2,
			QPS:         1,
			Timeout:     20,
			DialTimtout: 20,
			Data:        data,
		})
		assert.NoError(t, err)

		wg.Add(1)

		time.AfterFunc(time.Duration(1500*time.Millisecond), func() {
			count := gs.GetCount(callType)
			assert.Equal(t, 2, count)
		})

		go func() {
			report, err := reqr.Run()
			assert.NoError(t, err)
			assert.NotNil(t, report)
			assert.Len(t, report.ErrorDist, 0)
			wg.Done()
		}()
		wg.Wait()
	})
}

func TestRequesterServerStreaming(t *testing.T) {
	callType := helloworld.ServerStream

	gs, s, err := startServer(false)

	if err != nil {
		assert.FailNow(t, err.Error())
	}

	defer s.Stop()

	md, err := protodesc.GetMethodDesc("helloworld.Greeter.SayHellos", "./testdata/greeter.proto", []string{})

	data := make(map[string]interface{})
	data["name"] = "bob"

	gs.ResetCounters()

	reqr, err := New(md, &Options{
		Host:        localhost,
		N:           15,
		C:           3,
		Timeout:     20,
		DialTimtout: 20,
		Data:        data,
	})
	assert.NoError(t, err)

	report, err := reqr.Run()
	assert.NoError(t, err)
	assert.NotNil(t, report)
	assert.Equal(t, 15, int(report.Count))
	assert.Len(t, report.ErrorDist, 0)

	count := gs.GetCount(callType)
	assert.Equal(t, 15, count)
}

func TestRequesterClientStreaming(t *testing.T) {
	callType := helloworld.ClientStream

	gs, s, err := startServer(false)

	if err != nil {
		assert.FailNow(t, err.Error())
	}

	defer s.Stop()

	md, err := protodesc.GetMethodDesc("helloworld.Greeter.SayHelloCS", "./testdata/greeter.proto", []string{})

	m1 := make(map[string]interface{})
	m1["name"] = "bob"

	m2 := make(map[string]interface{})
	m2["name"] = "Kate"

	m3 := make(map[string]interface{})
	m3["name"] = "foo"

	data := []interface{}{m1, m2, m3}

	gs.ResetCounters()

	reqr, err := New(md, &Options{
		Host:        localhost,
		N:           16,
		C:           4,
		Timeout:     20,
		DialTimtout: 20,
		Data:        data,
	})
	assert.NoError(t, err)

	report, err := reqr.Run()
	assert.NoError(t, err)
	assert.NotNil(t, report)
	assert.Equal(t, 16, int(report.Count))
	assert.True(t, len(report.LatencyDistribution) > 1)
	assert.True(t, len(report.Histogram) > 1)
	assert.Len(t, report.ErrorDist, 0)

	count := gs.GetCount(callType)
	assert.Equal(t, 16, count)
}

func TestRequesterBidi(t *testing.T) {
	callType := helloworld.Bidi

	gs, s, err := startServer(false)

	if err != nil {
		assert.FailNow(t, err.Error())
	}

	defer s.Stop()

	md, err := protodesc.GetMethodDesc("helloworld.Greeter.SayHelloBidi", "./testdata/greeter.proto", []string{})

	m1 := make(map[string]interface{})
	m1["name"] = "bob"

	m2 := make(map[string]interface{})
	m2["name"] = "Kate"

	m3 := make(map[string]interface{})
	m3["name"] = "foo"

	data := []interface{}{m1, m2, m3}

	gs.ResetCounters()

	reqr, err := New(md, &Options{
		Host:        localhost,
		N:           20,
		C:           4,
		Timeout:     20,
		DialTimtout: 20,
		Data:        data,
	})
	assert.NoError(t, err)

	report, err := reqr.Run()
	assert.NoError(t, err)
	assert.NotNil(t, report)
	assert.Equal(t, 20, int(report.Count))
	assert.True(t, len(report.LatencyDistribution) > 1)
	assert.True(t, len(report.Histogram) > 1)
	assert.Len(t, report.ErrorDist, 0)

	count := gs.GetCount(callType)
	assert.Equal(t, 20, count)
}

func TestRequesterUnarySecure(t *testing.T) {
	callType := helloworld.Unary

	gs, s, err := startServer(true)

	if err != nil {
		assert.FailNow(t, err.Error())
	}

	defer s.Stop()

	md, err := protodesc.GetMethodDesc("helloworld.Greeter.SayHello", "./testdata/greeter.proto", []string{})

	data := make(map[string]interface{})
	data["name"] = "bob"

	gs.ResetCounters()

	reqr, err := New(md, &Options{
		Host:        localhost,
		N:           18,
		C:           3,
		Timeout:     20,
		DialTimtout: 20,
		Data:        data,
		Cert:        "./testdata/localhost.crt",
	})
	assert.NoError(t, err)

	report, err := reqr.Run()
	assert.NoError(t, err)
	assert.NotNil(t, report)
	assert.Equal(t, 18, int(report.Count))
	assert.Len(t, report.ErrorDist, 0)

	count := gs.GetCount(callType)
	assert.Equal(t, 18, count)
}
