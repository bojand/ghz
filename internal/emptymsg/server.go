package emptymsg

import (
	"context"
	fmt "fmt"

	"github.com/golang/protobuf/ptypes/empty"
)

// EmptyMessageService is the gRPC service implementation
type EmptyMessageService struct{}

// GetEmpty is the GetEmpty call
func (s *EmptyMessageService) GetEmpty(ctx context.Context, req *empty.Empty) (*empty.Empty, error) {
	fmt.Println("GetEmpty")
	return &empty.Empty{}, nil
}

// GetEmptyMessage is the GetEmptyMessage call
func (s *EmptyMessageService) GetEmptyMessage(ctx context.Context, req *EmptyMessage) (*EmptyMessage, error) {
	fmt.Println("EmptyMessageService")
	return &EmptyMessage{}, nil
}
