package handler

import (
	"github.com/Everest13/fin-aggregator-service/internal/service/bank"
	"github.com/Everest13/fin-aggregator-service/internal/service/category"
	"github.com/Everest13/fin-aggregator-service/internal/service/transaction"
	"github.com/Everest13/fin-aggregator-service/internal/service/user"
	pb "github.com/Everest13/fin-aggregator-service/pkg/api/fin-aggregate-service"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func convertBankListToPb(banks []bank.Bank) []*pb.Bank {
	res := make([]*pb.Bank, len(banks))
	for i, b := range banks {
		res[i] = &pb.Bank{
			Id:           b.ID,
			Name:         b.Name,
			ImportMethod: convertBankImportTypeToPb(b.ImportMethods),
		}
	}

	return res
}

func convertBankImportTypeToPb(importTypes []bank.ImportMethod) []pb.BankImportMethod {
	res := make([]pb.BankImportMethod, 0, len(importTypes))
	for _, importType := range importTypes {
		switch importType {
		case bank.CSVImportMethod:
			res = append(res, pb.BankImportMethod_CSV)
		case bank.APIImportMethod:
			res = append(res, pb.BankImportMethod_API)
		default:
			res = append(res, pb.BankImportMethod_UNDEFINED)
		}
	}

	return res
}

func convertUserListToPb(users []user.User) []*pb.User {
	res := make([]*pb.User, len(users))
	for i, u := range users {
		res[i] = &pb.User{
			Id:    u.ID,
			Name:  u.Name,
			Banks: u.Banks,
		}
	}

	return res
}

func convertCategoryListToPb(categories []category.Category) []*pb.Category {
	res := make([]*pb.Category, len(categories))
	for i, c := range categories {
		res[i] = &pb.Category{
			Id:   c.ID,
			Name: c.Name,
		}
	}

	return res
}

func convertRecordErrorsPb(recordErrs map[int64][]error) []*pb.RecordError {
	pbRecordError := make([]*pb.RecordError, 0, len(recordErrs))
	for rowID, errs := range recordErrs {
		recErrs := make([]string, 0, len(errs))
		for _, err := range errs {
			recErrs = append(recErrs, err.Error())
		}
		pbRecordError = append(pbRecordError, &pb.RecordError{
			RowId:  rowID,
			Errors: recErrs,
		})
	}

	return pbRecordError
}

func convertTransactionsToPb(transactions []transaction.EnrichedTransaction) []*pb.Transaction {
	res := make([]*pb.Transaction, len(transactions))
	for i, tr := range transactions {
		res[i] = &pb.Transaction{
			Id:              tr.ID,
			BankId:          tr.BankID,
			ExternalId:      tr.ExternalID,
			UserId:          tr.UserID,
			Amount:          tr.Amount,
			CategoryId:      tr.CategoryID,
			Description:     tr.Description,
			Type:            mapTransactionTypeToPb(tr.Type),
			TransactionDate: timestamppb.New(tr.TransactionDate),
			CreatedAt:       timestamppb.New(tr.CreatedAt),
			BankName:        tr.BankName,
			CategoryName:    tr.CategoryName,
			UserName:        tr.UserName,
		}
	}

	return res
}

func convertTransactionToPb(tr *transaction.EnrichedTransaction) *pb.Transaction {
	return &pb.Transaction{
		Id:              tr.ID,
		BankId:          tr.BankID,
		ExternalId:      tr.ExternalID,
		UserId:          tr.UserID,
		Amount:          tr.Amount,
		CategoryId:      tr.CategoryID,
		Description:     tr.Description,
		Type:            mapTransactionTypeToPb(tr.Type),
		TransactionDate: timestamppb.New(tr.TransactionDate),
		CreatedAt:       timestamppb.New(tr.CreatedAt),
		BankName:        tr.BankName,
		CategoryName:    tr.CategoryName,
	}
}

func mapTransactionTypeToPb(t transaction.TransactionType) pb.TransactionType {
	switch t {
	case transaction.IncomeTransactionType:
		return pb.TransactionType_INCOME
	case transaction.OutcomeTransactionType:
		return pb.TransactionType_OUTCOME
	default:
		return pb.TransactionType_UNSPECIFIED
	}
}

func mapPbToTransactionType(t pb.TransactionType) transaction.TransactionType {
	switch t {
	case pb.TransactionType_INCOME:
		return transaction.IncomeTransactionType
	case pb.TransactionType_OUTCOME:
		return transaction.OutcomeTransactionType
	default:
		return transaction.UnspecifiedTransactionType
	}
}

func convertTransactionTypeList(types []transaction.TransactionType) []pb.TransactionType {
	res := make([]pb.TransactionType, 0, len(types))
	for _, t := range types {
		res = append(res, mapTransactionTypeToPb(t))
	}

	return res
}
