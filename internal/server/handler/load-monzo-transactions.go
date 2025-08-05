package handler

import (
	"context"
	pb "github.com/Everest13/fin-aggregator-service/pkg/api/fin-aggregate-service"
)

func (f *FinAggregatorServer) LoadMonzoTransactions(ctx context.Context, req *pb.LoadMonzoTransactionsRequest) (*pb.LoadMonzoTransactionsResponse, error) {
	err := f.monzoService.GetMonzoTransactions(
		ctx,
		req.GetSince().AsTime(),
		req.GetBefore().AsTime(),
		req.GetUserId(),
		req.GetBankId(),
	)
	if err != nil {
		return nil, err
	}

	return &pb.LoadMonzoTransactionsResponse{
		Success: true,
	}, nil
}
