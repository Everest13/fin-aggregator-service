package handler

import (
	"context"
	pb "github.com/Everest13/fin-aggregator-service/pkg/api/fin-aggregate-service"
)

func (f *FinAggregatorServer) ListBank(ctx context.Context, _ *pb.ListBankRequest) (*pb.ListBankResponse, error) {
	banks, err := f.bankService.BankList(ctx)
	if err != nil {
		return nil, err
	}

	return &pb.ListBankResponse{
		Banks: convertBankListToPb(banks),
	}, nil
}
