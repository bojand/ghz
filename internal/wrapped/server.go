package wrapped

import (
	"context"

	wrappers "github.com/golang/protobuf/ptypes/wrappers"
)

type WrappedService struct{}

func (s *WrappedService) GetMessage(ctx context.Context, req *wrappers.StringValue) (*wrappers.StringValue, error) {
	return &wrappers.StringValue{Value: "Hello: " + req.GetValue()}, nil
}

func (s *WrappedService) GetBytesMessage(ctx context.Context, req *wrappers.BytesValue) (*wrappers.BytesValue, error) {
	return &wrappers.BytesValue{Value: req.GetValue()}, nil
}

func (s *WrappedService) mustEmbedUnimplementedWrappedServiceServer() {}
