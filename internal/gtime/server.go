package gtime

import (
	"context"
	"strconv"
	"time"
)

type TimeService struct {
	UnimplementedTimeServiceServer

	LastTimestamp time.Time
	LastDuration  time.Duration
}

func (s *TimeService) TestCall(ctx context.Context, req *CallRequest) (*CallReply, error) {

	s.LastTimestamp = req.GetTs().AsTime()
	s.LastDuration = req.GetDur().AsDuration()

	return &CallReply{
		Ts:      req.GetTs(),
		Dur:     req.GetDur(),
		Message: strconv.FormatUint(req.GetUserId(), 10),
	}, nil
}
