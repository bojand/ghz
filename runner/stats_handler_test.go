package runner

import (
	"context"
	"testing"
	"time"

	"github.com/bojand/ghz/internal"
	"github.com/bojand/ghz/internal/helloworld"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

func TestStatsHandler(t *testing.T) {
	_, s, err := internal.StartServer(false)

	if err != nil {
		assert.FailNow(t, err.Error())
	}

	defer s.Stop()

	rChan := make(chan *callResult, 1)
	done := make(chan bool, 1)
	results := make([]*callResult, 0, 2)

	go func() {
		for res := range rChan {
			results = append(results, res)
		}
		done <- true
	}()

	conn, err := grpc.Dial(
		internal.TestLocalhost,
		grpc.WithInsecure(),
		grpc.WithStatsHandler(&statsHandler{results: rChan}))

	if err != nil {
		assert.FailNow(t, err.Error())
	}

	c := helloworld.NewGreeterClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err = c.SayHello(ctx, &helloworld.HelloRequest{Name: "Bob"})
	assert.NoError(t, err)

	_, err = c.SayHello(ctx, &helloworld.HelloRequest{Name: "Kate"})
	assert.NoError(t, err)

	close(rChan)

	<-done

	assert.Equal(t, 2, len(results))
	assert.NotNil(t, results[0])
	assert.NotNil(t, results[1])
}
