package handler

import (
	"context"
	pb "github.com/Everest13/fin-aggregator-service/pkg/api/fin-aggregate-service"
)

func (f *FinAggregatorServer) UploadCSV(ctx context.Context, req *pb.UploadCSVRequest) (*pb.UploadCSVResponse, error) {
	recordErrs, err := f.parserService.UploadCSV(ctx, req.GetBankId(), req.GetUserId(), req.GetCsvData())
	if err != nil {
		return nil, err
	}

	return &pb.UploadCSVResponse{
		Success:     true,
		RecordError: convertRecordErrorsPb(recordErrs),
	}, nil
}
