package monzo

import (
	"context"
	"fmt"
	"github.com/Everest13/fin-aggregator-service/internal/service/category"
	"github.com/Everest13/fin-aggregator-service/internal/service/transaction"
	"github.com/Everest13/fin-aggregator-service/internal/utils/logger"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strconv"
	"strings"
	"time"

	"github.com/Everest13/fin-aggregator-service/internal/utils/random"
)

type Service struct {
	client *client

	stateStore         *stateStore
	tokenStore         *tokenStore
	transactionService *transaction.Service
	categoryStore      *category.Store
}

func NewService(monzoCfg *MonzoCfg, timeout time.Duration, transactionService *transaction.Service, categoryStore *category.Store) *Service {
	return &Service{
		client:             newClient(timeout, monzoCfg),
		stateStore:         &stateStore{},
		tokenStore:         &tokenStore{},
		transactionService: transactionService,
		categoryStore:      categoryStore,
	}
}

func (s *Service) GetAuthorizationURL() (string, error) {
	state, err := random.GenerateRandomString(lenState)
	if err != nil {
		return "", status.Errorf(codes.Internal, "failed to get auth url: %v", err)
	}

	s.stateStore.set(state)

	return s.client.generateAuthURL(state), nil
}

func (s *Service) AuthCallback(ctx context.Context, state, code string) error {
	if state != s.stateStore.get() {
		return status.Errorf(codes.InvalidArgument, "invalid state parameter")
	}

	s.stateStore.delete()

	monzoTokens, err := s.client.getTokens(ctx, code)
	if err != nil {
		return status.Errorf(codes.Unavailable, "failed to get monzo tokens: %v", err)
	}

	accountID, err := s.client.getAccountID(ctx, monzoTokens.AccessToken)
	if err != nil {
		return status.Errorf(codes.Unavailable, "failed to get monzo account ID: %v", err)
	}

	s.tokenStore.set(&authTokens{
		accessToken:  monzoTokens.AccessToken,
		refreshToken: monzoTokens.RefreshToken,
		expiresIn:    monzoTokens.ExpiresIn,
		accountID:    accountID,
		issuedAt:     time.Now(),
	})

	return nil
}

func (s *Service) GetMonzoTransactions(ctx context.Context, since, before time.Time) error {
	if err := s.checkAuth(ctx); err != nil {
		return err
	}
	authToken := s.tokenStore.get()

	monzoTransaction, err := s.client.getMonzoTransactions(ctx, authToken.accessToken, authToken.accountID, since, before)
	if err != nil {
		return status.Errorf(codes.Unavailable, "fetching Monzo transactions failed: %v", err)
	}

	if len(monzoTransaction) == 0 {
		//todo
	}

	err = s.saveMonzoTransactions(ctx, monzoTransaction, since)
	if err != nil {
		return status.Errorf(codes.Internal, "saving Monzo transactions failed: %v", err)
	}

	return nil
}

func (s *Service) checkAuth(ctx context.Context) error {
	tokens := s.tokenStore.get()
	if tokens == nil {
		logger.Error("user is unauthenticated: tokens undefined", nil)
		return status.Errorf(codes.Unauthenticated, "user is unauthenticated")
	}

	if !tokens.IsExpired() {
		return nil
	}

	monzoTokens, err := s.client.refreshToken(ctx, tokens.refreshToken)
	if err != nil {
		logger.Error("failed to get monzo tokens", err)
		return status.Errorf(codes.Unauthenticated, "user is unauthenticated")
	}

	s.tokenStore.set(&authTokens{
		accessToken:  monzoTokens.AccessToken,
		refreshToken: monzoTokens.RefreshToken,
		expiresIn:    monzoTokens.ExpiresIn,
		accountID:    tokens.accountID,
		issuedAt:     time.Now(),
	})

	return nil
}

func (s *Service) saveMonzoTransactions(ctx context.Context, monzoTransactions []MonzoTransaction, since time.Time) error {
	trs := make([]*transaction.Transaction, 0, len(monzoTransactions))
	for _, mTr := range monzoTransactions {
		//var err []error //todo собирать ошибки обработки транзакции

		date, err := convertDate(mTr.Created, since)
		if err != nil {
			//todo
		}

		//todo хардкод
		tr := &transaction.Transaction{
			BankID:          1,
			UserID:          1,
			ExternalID:      mTr.ID,
			Amount:          convertAmount(mTr.Amount),
			Description:     convertDescription(mTr.Description, mTr.Category, mTr.Notes, mTr.Scheme),
			TransactionDate: date,
			CategoryID:      s.convertCategory(mTr.Category, mTr.Description),
			Type:            convertType(mTr.Amount),
		}

		trs = append(trs, tr)
	}

	err := s.transactionService.SaveTransactions(ctx, trs)
	if err != nil {
		logger.Error("failed to save monzo transactions", err)
		return status.Errorf(codes.Internal, "failed to get monzo transactions")
	}

	return nil
}

func convertAmount(amount int64) string {
	pounds := amount / 100
	return strconv.FormatInt(pounds, 10)
}

func convertType(amount int64) transaction.TransactionType {
	if amount < 0 {
		return transaction.OutcomeTransactionType
	}
	return transaction.IncomeTransactionType
}

func convertDescription(desc, category, notes, scheme string) string {
	res := strings.TrimSpace(desc)

	category = strings.TrimSpace(category)
	if category != "" {
		res += ", " + category
	}

	notes = strings.TrimSpace(notes)
	if notes != "" {
		res += ", " + notes
	}

	scheme = strings.TrimSpace(scheme)
	if scheme != "" {
		res += ", " + scheme
	}

	return res
}

func convertDate(createdAt string, from time.Time) (time.Time, error) {
	var dateFormats = []string{
		"2006-01-02T15:04:05.000Z",
	}

	if createdAt == "" {
		return from, fmt.Errorf("empty date data")
	}

	for _, layout := range dateFormats {
		if t, err := time.Parse(layout, createdAt); err == nil {
			return t, nil
		}
	}

	return from, fmt.Errorf("unknown date format: %s", createdAt)
}

func (s *Service) convertCategory(mCategory, desc string) int64 {
	keywordCategory := s.categoryStore.GetKeywordCategoryMap()

	combinedText := strings.ToLower(strings.Join([]string{mCategory, desc}, " "))
	for keyword, categoryID := range keywordCategory {
		if strings.Contains(combinedText, strings.ToLower(keyword)) {
			return categoryID
		}
	}

	return category.UncategorizedID
}
