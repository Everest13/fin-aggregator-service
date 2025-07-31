package parser

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"github.com/Everest13/fin-aggregator-service/internal/service/bank"
	"github.com/Everest13/fin-aggregator-service/internal/service/category"
	"github.com/Everest13/fin-aggregator-service/internal/service/transaction"
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
		parserFactory:      newParserFactory(categoryService.Store()),
		bankService:        bankService,
		transactionService: transactionService,
		categoryService:    categoryService,
	}

	return service
}

// todo add logs
func (s *Service) UploadCSV(ctx context.Context, bankID, userID int64, csvData []byte) (map[int64][]error, error) {
	bankInfo, err := s.bankService.GetBank(ctx, bankID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get bank info: %v", err)
	}

	bankParser := s.parserFactory.GetParser(bank.BankName(bankInfo.Name))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get parser: %v", err)
	}

	reader := csv.NewReader(bytes.NewReader(csvData))
	records, err := reader.ReadAll()
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to read CSV: %v", err)
	}

	if len(records) == 0 || len(records[0]) == 0 {
		return nil, status.Error(codes.InvalidArgument, "CSV file is empty or invalid")
	}

	targetIDsMap, err := s.ValidateHeaders(ctx, records[0], bankID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "CSV headers validation failed: %v", err)
	}

	transactions, recordErrs := bankParser.ParseRecords(records, targetIDsMap, bankID, userID)
	err = s.transactionService.SaveTransactions(ctx, transactions)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to save transactions: %v", err)
	}

	return recordErrs, nil
}

var requiredFields = []bank.TargetField{
	bank.DateTargetField,
	bank.AmountTargetField,
}

func (s *Service) ValidateHeaders(ctx context.Context, csvHeaders []string, bankID int64) (map[bank.TargetField][]int, error) {
	headers, err := s.bankService.GetBankHeaders(ctx, bankID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get bank headers: %v", err)
	}

	headerMap := make(map[string]bank.Header, len(headers))
	for _, h := range headers {
		headerMap[h.Name] = h
	}

	targetField := map[bank.TargetField][]int{}
	for i, h := range csvHeaders {
		if header, ok := headerMap[h]; ok {
			for _, t := range header.TargetField {
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
