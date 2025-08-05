package parser

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"github.com/Everest13/fin-aggregator-service/internal/service/bank"
	"github.com/Everest13/fin-aggregator-service/internal/service/category"
	"github.com/Everest13/fin-aggregator-service/internal/service/transaction"
	"github.com/Everest13/fin-aggregator-service/internal/utils/logger"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Service struct {
	parserFactory      *parserFactory
	bankService        *bank.Service
	transactionService *transaction.Service
	categoryService    *category.Service
}

func NewService(
	bankService *bank.Service,
	transactionService *transaction.Service,
	categoryService *category.Service,
) *Service {

	service := &Service{
		parserFactory:      newParserFactory(categoryService),
		bankService:        bankService,
		transactionService: transactionService,
		categoryService:    categoryService,
	}

	return service
}

func (s *Service) UploadCSV(ctx context.Context, bankID, userID int64, csvData []byte) (map[int64][]error, error) {
	bankInfo, err := s.bankService.GetBank(ctx, bankID)
	if err != nil {
		logger.ErrorWithFields("failed to get bank", err, "bank_id", bankID)
		return nil, status.Errorf(codes.Internal, "failed to get bank")
	}

	bankParser := s.parserFactory.GetParser(bank.BankName(bankInfo.Name))

	reader := csv.NewReader(bytes.NewReader(csvData))
	records, err := reader.ReadAll()
	if err != nil {
		logger.ErrorWithFields("CSV parsing error", err, "bank_id", bankID, "user_id", userID)
		return nil, status.Errorf(codes.InvalidArgument, "invalid CSV format")
	}

	if len(records) == 0 || len(records[0]) == 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid CSV format")
	}

	targetIDsMap, err := s.validateHeaders(ctx, records[0], bankID)
	if err != nil {
		logger.ErrorWithFields("CSV headers validation failed", err, "bank_id", bankID, "user_id", userID)
		return nil, status.Errorf(codes.InvalidArgument, "invalid CSV format")
	}

	transactions, recordErrs := bankParser.ParseRecords(ctx, records, targetIDsMap, bankID, userID)
	err = s.transactionService.SaveTransactions(ctx, transactions)
	if err != nil {
		logger.ErrorWithFields("transaction persistence error", err, "bank_id", bankID, "user_id", userID)
		return nil, status.Errorf(codes.Internal, "transaction persistence error")
	}

	//todo
	return recordErrs, nil
}

var requiredFields = []bank.TargetField{
	bank.DateTargetField,
	bank.AmountTargetField,
}

func (s *Service) validateHeaders(ctx context.Context, csvHeaders []string, bankID int64) (map[bank.TargetField][]int, error) {
	headers, err := s.bankService.GetBankHeaders(ctx, bankID)
	if err != nil {
		return nil, fmt.Errorf("failed to get bank headers: %w", err)
	}

	headerMap := make(map[string][]bank.TargetField, len(headers))
	for _, h := range headers {
		headerMap[h.Name] = append(headerMap[h.Name], h.TargetField...)
	}

	targetField := map[bank.TargetField][]int{}
	for i, h := range csvHeaders {
		if header, ok := headerMap[h]; ok {
			for _, t := range header {
				targetField[t] = append(targetField[t], i)
			}
		}
	}

	missing := []bank.TargetField{}
	for _, f := range requiredFields {
		if _, ok := targetField[f]; !ok {
			missing = append(missing, f)
		}
	}

	if len(missing) > 0 {
		return nil, fmt.Errorf("missing required fields in CSV: %v", missing)
	}

	return targetField, nil
}
