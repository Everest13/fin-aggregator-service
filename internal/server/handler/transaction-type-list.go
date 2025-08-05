package handler

import (
	"context"
	pb "github.com/Everest13/fin-aggregator-service/pkg/api/fin-aggregate-service"
)

func (f *FinAggregatorServer) ListTransactionType(_ context.Context, _ *pb.ListTransactionTypeRequest) (*pb.ListTransactionTypeResponse, error) {
	types := f.transactionService.GetTransactionTypeList()

	return &pb.ListTransactionTypeResponse{
		Type: convertTransactionTypeList(types),
	}, nil
}
