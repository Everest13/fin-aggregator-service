package handler

import (
	"context"

	pb "github.com/Everest13/fin-aggregator-service/pkg/api/fin-aggregate-service"
)

func (f *FinAggregatorServer) GetMonzoAccount(ctx context.Context, _ *pb.MonzoAccountRequest) (*pb.MonzoAccountResponse, error) {
	err := f.monzoService.GetAccountID(ctx)
	if err != nil {
		return nil, err
	}

	return &pb.MonzoAccountResponse{Success: true}, nil
}
