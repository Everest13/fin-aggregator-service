package handler

import (
	"context"
	pb "github.com/Everest13/fin-aggregator-service/pkg/api/fin-aggregate-service"
)

func (f *FinAggregatorServer) GetTransactions(ctx context.Context, req *pb.GetTransactionsRequest) (*pb.GetTransactionsResponse, error) {
	trSummary, err := f.transactionService.GetSummaryTransactions(ctx, req.GetMonth(), req.GetYear(), req.GetUserId(), req.GetBankId())
	if err != nil {
		return nil, err
	}

	return &pb.GetTransactionsResponse{
		Transactions: convertTransactionsToPb(trSummary.Transactions),
		TotalCount:   int32(trSummary.TotalCount),
		TotalIncome:  trSummary.TotalIncome,
		TotalOutcome: trSummary.TotalOutcome,
	}, nil
}
