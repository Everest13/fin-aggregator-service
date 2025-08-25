package uploader

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"github.com/Everest13/fin-aggregator-service/internal/service/bank"
	"github.com/Everest13/fin-aggregator-service/internal/service/category"
	"github.com/Everest13/fin-aggregator-service/internal/service/transaction"
	csvParser "github.com/Everest13/fin-aggregator-service/internal/service/uploader/csv-parser"
	"github.com/Everest13/fin-aggregator-service/internal/utils/logger"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"sync"
)

var chunkSize = 100

type Service struct {
	repo               *repository
	headerMappingStore *HeaderMappingStore
	csvParserFactory   *csvParser.Factory
	bankService        *bank.Service
	transactionService *transaction.Service
	categoryService    *category.Service
}

func NewService(
	dbPool *pgxpool.Pool,
	bankService *bank.Service,
	transactionService *transaction.Service,
	categoryService *category.Service,
) *Service {
	service := &Service{
		repo:               newRepository(dbPool),
		headerMappingStore: NewHeaderMappingStore(),
		csvParserFactory:   csvParser.NewFactory(categoryService),
		bankService:        bankService,
		transactionService: transactionService,
		categoryService:    categoryService,
	}

	return service
}

func (s *Service) Initialize(ctx context.Context) error {
	headerMappings, err := s.repo.getBankHeaderMappings(ctx)
	if err != nil {
		logger.Error("failed to get header_mapping", err)
		return fmt.Errorf("service initialization failed: %w", err)
	}

	bankHeaderMapping := map[int64][]HeaderMapping{}
	for _, headerMapping := range headerMappings {
		bankHeaderMapping[headerMapping.BankID] = append(bankHeaderMapping[headerMapping.BankID], headerMapping)
	}

	s.headerMappingStore.Set(bankHeaderMapping)

	return nil
}

func (s *Service) UploadCSV(ctx context.Context, bankID, userID int64, csvData []byte) (map[int64][]error, error) {
	reader := csv.NewReader(bytes.NewReader(csvData))
	records, err := reader.ReadAll()
	if err != nil {
		logger.ErrorWithFields("CSV parsing error", err, "bank_id", bankID, "user_id", userID)
		return nil, status.Errorf(codes.InvalidArgument, "invalid CSV format")
	}

	if len(records) == 0 || len(records[0]) == 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid CSV format")
	}

	return s.processTransactionBatches(ctx, records, bankID, userID)
}

func (s *Service) processTransactionBatches(ctx context.Context, records [][]string, bankID, userID int64) (map[int64][]error, error) {
	bankParser, err := s.getBankParser(ctx, bankID)
	if err != nil {
		return nil, err
	}

	transactionFieldColumns, err := s.getTransactionFieldColumns(ctx, records[0], bankID)
	if err != nil {
		logger.ErrorWithFields("transaction fields columns getting failed", err, "bank_id", bankID, "user_id", userID)
		return nil, status.Errorf(codes.InvalidArgument, "invalid CSV format")
	}

	var wg sync.WaitGroup

	records = records[1:]
	errCh := make(chan map[int64][]error, (len(records)/chunkSize)+1)
	processChunk := func(chunk [][]string, startRow int) {
		defer wg.Done()

		transactions, recordsErrs := bankParser.ParseRecords(ctx, chunk, transactionFieldColumns, bankID, userID)
		if len(recordsErrs) > 0 {
			mappedErrs := make(map[int64][]error)
			for recordErrs, errs := range recordsErrs {
				mappedErrs[recordErrs+int64(startRow)] = errs
			}
			errCh <- mappedErrs
		}

		saveErr := s.transactionService.SaveTransactions(ctx, transactions)
		if saveErr != nil {
			//todo handling err
			logger.ErrorWithFields("transaction persistence error", err, "bank_id", bankID, "user_id", userID)
		}
	}

	for i := 0; i < len(records); i += chunkSize {
		end := getChunkEnd(i+chunkSize, len(records))
		chunk := records[i:end]

		wg.Add(1)
		go processChunk(chunk, i+1)
	}

	wg.Wait()
	close(errCh)

	allRecordErrs := make(map[int64][]error)
	for recErrs := range errCh {
		for row, errs := range recErrs {
			allRecordErrs[row] = append(allRecordErrs[row], errs...)
		}
	}

	return allRecordErrs, nil
}

func getChunkEnd(a, b int) int {
	if a > b {
		return b
	}
	return a
}

func (s *Service) getBankParser(ctx context.Context, bankID int64) (csvParser.Parser, error) {
	bankInfo, err := s.bankService.GetBank(ctx, bankID)
	if err != nil {
		logger.ErrorWithFields("failed to get bank", err, "bank_id", bankID)
		return nil, status.Errorf(codes.Internal, "failed to get bank")
	}

	for _, importMethod := range bankInfo.ImportMethods {
		if importMethod == bank.CSVImportMethod {
			return s.csvParserFactory.GetParser(bank.BankName(bankInfo.Name)), nil
		}
	}

	logger.ErrorWithFields("inappropriate bank import method", nil, "bank_id", bankID)
	return nil, status.Errorf(codes.Internal, "failed to get bank")
}

func (s *Service) getHeaderMapping(ctx context.Context, bankID int64) ([]HeaderMapping, error) {
	headerMapping := s.headerMappingStore.GetByBank(bankID)
	if headerMapping != nil {
		return headerMapping, nil
	}

	logger.ErrorWithFields("failed to get header mapping from cache", nil, "bank_id", bankID)

	headerMapping, err := s.repo.getBankHeaderMappingsByBank(ctx, bankID)
	if err != nil {
		logger.Error("failed to get header_mapping", err)
		return nil, fmt.Errorf("service initialization failed: %w", err)
	}

	return headerMapping, nil
}

func (s *Service) getTransactionFieldColumns(ctx context.Context, csvHeaders []string, bankID int64) (map[transaction.TransactionField][]int, error) {
	headerMapping, err := s.getHeaderMapping(ctx, bankID)
	if err != nil {
		return nil, fmt.Errorf("failed to get bank headers: %w", err)
	}

	csvHeadersMap := make(map[string]int, len(csvHeaders))
	for i, h := range csvHeaders {
		csvHeadersMap[h] = i
	}

	transactionFieldColumns := map[transaction.TransactionField][]int{}
	for _, hP := range headerMapping {
		id, ok := csvHeadersMap[hP.Name]
		if ok {
			for _, trField := range hP.TrFields {
				transactionFieldColumns[trField] = append(transactionFieldColumns[trField], id)
			}
			continue
		}

		if hP.Required {
			return nil, fmt.Errorf("missing required fields in CSV: %v", hP.Name)
		}
	}

	return transactionFieldColumns, err
}
