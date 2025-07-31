package handler

import (
	"context"

	pb "github.com/Everest13/fin-aggregator-service/pkg/api/fin-aggregate-service"
)

func (f *FinAggregatorServer) MonzoCallback(ctx context.Context, req *pb.MonzoCallbackRequest) (*pb.MonzoCallbackResponse, error) {
	err := f.monzoService.AuthCallback(ctx, req.GetState(), req.GetCode())
	if err != nil {
		return nil, err
	}

	return &pb.MonzoCallbackResponse{
		Success: true,
	}, nil
}
