package wrapped

import (
	"context"

	wrappers "github.com/golang/protobuf/ptypes/wrappers"
)

type WrappedService struct{}

func (s *WrappedService) GetMessage(ctx context.Context, req *wrappers.StringValue) (*wrappers.StringValue, error) {
	return &wrappers.StringValue{Value: "Hello: " + req.GetValue()}, nil
}
