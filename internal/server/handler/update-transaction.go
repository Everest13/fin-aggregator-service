package handler

import (
	"context"
	"github.com/Everest13/fin-aggregator-service/internal/service/transaction"
	pb "github.com/Everest13/fin-aggregator-service/pkg/api/fin-aggregate-service"
)

func (f *FinAggregatorServer) UpdateTransaction(ctx context.Context, req *pb.UpdateTransactionRequest) (*pb.UpdateTransactionResponse, error) {
	updateData := &transaction.TransactionUpdateData{
		ID: req.GetTransactionId(),
	}
	if req.Type != nil {
		trType := mapPbToTransactionType(req.GetType())
		updateData.Type = &trType
	}

	if req.CategoryId != nil {
		updateData.CategoryID = req.CategoryId
	}

	tr, err := f.transactionService.UpdateTransaction(ctx, updateData)
	if err != nil {
		return nil, err
	}

	return &pb.UpdateTransactionResponse{
		Transaction: convertTransactionToPb(tr),
	}, nil
}
