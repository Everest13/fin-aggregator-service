package handler

import (
	"github.com/Everest13/fin-aggregator-service/internal/service/bank"
	"github.com/Everest13/fin-aggregator-service/internal/service/category"
	"github.com/Everest13/fin-aggregator-service/internal/service/monzo"
	"github.com/Everest13/fin-aggregator-service/internal/service/parser"
	"github.com/Everest13/fin-aggregator-service/internal/service/transaction"
	"github.com/Everest13/fin-aggregator-service/internal/service/user"
	pb "github.com/Everest13/fin-aggregator-service/pkg/api/fin-aggregate-service"
)

type FinAggregatorServer struct {
	pb.UnimplementedFinAggregatorServiceServer
	transactionService *transaction.Service
	bankService        *bank.Service
	categoryService    *category.Service
	userService        *user.Service
	parserService      *parser.Service
	monzoService       *monzo.Service
}

func NewFinAggregatorServer(
	transactionService *transaction.Service,
	bankService *bank.Service,
	categoryService *category.Service,
	userService *user.Service,
	parserService *parser.Service,
	monzoService *monzo.Service,
) *FinAggregatorServer {
	return &FinAggregatorServer{
		transactionService: transactionService,
		bankService:        bankService,
		categoryService:    categoryService,
		userService:        userService,
		parserService:      parserService,
		monzoService:       monzoService,
	}
}
