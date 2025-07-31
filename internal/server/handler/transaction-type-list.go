package handler

import (
	"context"
	pb "github.com/Everest13/fin-aggregator-service/pkg/api/fin-aggregate-service"
)

func (f *FinAggregatorServer) ListTransactionType(_ context.Context, _ *pb.ListTransactionTypeRequest) (*pb.ListTransactionTypeResponse, error) {
	return &pb.ListTransactionTypeResponse{
		Type: []pb.TransactionType{
			pb.TransactionType_INCOME,
			pb.TransactionType_OUTCOME,
			pb.TransactionType_UNSPECIFIED,
		},
	}, nil
}
