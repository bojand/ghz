package sleep

import (
	"context"
	"time"
)

type SleepService struct{}

func (s *SleepService) SleepFor(ctx context.Context, req *SleepRequest) (*SleepResponse, error) {
	time.Sleep(time.Duration(req.Milliseconds) * time.Millisecond)
	return &SleepResponse{}, nil
}
