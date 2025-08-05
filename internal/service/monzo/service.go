package monzo

import (
	"context"
	"github.com/Everest13/fin-aggregator-service/internal/service/category"
	"github.com/Everest13/fin-aggregator-service/internal/service/transaction"
	"github.com/Everest13/fin-aggregator-service/internal/utils/logger"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"time"

	"github.com/Everest13/fin-aggregator-service/internal/utils/random"
)

type Service struct {
	client             *client
	authStore          *authStore
	transactionService *transaction.Service
	categoryService    *category.Service
}

func NewService(monzoCfg *MonzoCfg, timeout time.Duration, transactionService *transaction.Service, categoryService *category.Service) *Service {
	return &Service{
		client:             newClient(timeout, monzoCfg),
		authStore:          &authStore{},
		transactionService: transactionService,
		categoryService:    categoryService,
	}
}

func (s *Service) GetAuthorizationURL() (string, error) {
	state, err := random.GenerateRandomString(lenState)
	if err != nil {
		logger.Error("failed to get monzo auth url", err)
		return "", status.Errorf(codes.Internal, "failed to get auth url")
	}

	s.authStore.setState(state)

	return s.client.generateAuthURL(state), nil
}

func (s *Service) AuthCallback(ctx context.Context, state, code string) error {
	if state != s.authStore.getState() {
		logger.ErrorWithFields("invalid state parameter", nil, "state", state)
		return status.Errorf(codes.InvalidArgument, "invalid state parameter")
	}

	s.authStore.deleteState()

	monzoTokens, err := s.client.getTokens(ctx, code)
	if err != nil {
		logger.Error("failed to get monzo tokens", err)
		return status.Errorf(codes.Unavailable, "failed to get monzo tokens")
	}

	s.authStore.setAuthToken(&authTokens{
		accessToken:  monzoTokens.AccessToken,
		refreshToken: monzoTokens.RefreshToken,
		expiresIn:    monzoTokens.ExpiresIn,
		issuedAt:     time.Now(),
	})

	return nil
}

func (s *Service) GetAccountID(ctx context.Context) error {
	err := s.checkTokens(ctx)
	if err != nil {
		return err
	}

	tokens := s.authStore.getAuthToken()

	accountID, err := s.client.getAccountID(ctx, tokens.accessToken)
	if err != nil {
		logger.Error("failed to get monzo account id", err)
		return status.Errorf(codes.Unavailable, "Monzo API error: %v", err)
	}

	s.authStore.setAccountID(accountID)

	return nil
}

func (s *Service) checkTokens(ctx context.Context) error {
	tokens := s.authStore.getAuthToken()
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

	s.authStore.setAuthToken(&authTokens{
		accessToken:  monzoTokens.AccessToken,
		refreshToken: monzoTokens.RefreshToken,
		expiresIn:    monzoTokens.ExpiresIn,
		issuedAt:     time.Now(),
	})

	return nil
}

func (s *Service) checkAuth(ctx context.Context) error {
	err := s.checkTokens(ctx)
	if err != nil {
		return status.Errorf(codes.Unauthenticated, "user is unauthenticated")
	}

	accountID := s.authStore.getAccountID()
	if len(accountID) == 0 {
		return status.Errorf(codes.Unauthenticated, "user is unauthenticated")
	}

	return nil
}

func (s *Service) GetMonzoTransactions(ctx context.Context, since, before time.Time, userID, bankID int64) error {
	if err := s.checkAuth(ctx); err != nil {
		return err
	}

	authToken := s.authStore.getAuthToken()
	accountID := s.authStore.getAccountID()
	if authToken == nil || accountID == "" {
		return status.Errorf(codes.Unauthenticated, "user is unauthenticated")
	}

	monzoTransaction, err := s.client.getMonzoTransactions(ctx, authToken.accessToken, accountID, since, before)
	if err != nil {
		logger.ErrorWithFields("failed to fetch Monzo transactions", err, "account_id", accountID, "since", since, "before", before)
		return status.Errorf(codes.Unavailable, "failed to get Monzo transactions")
	}

	if len(monzoTransaction) == 0 {
		logger.WarnWithFields("no Monzo transactions found", "account_id", accountID, "since", since, "before", before)
		return nil
	}

	trs, trErr, err := s.parseMonzoTransactions(ctx, monzoTransaction, since, userID, bankID)
	if err != nil {
		logger.ErrorWithFields("failed to parse Monzo transactions", err, "since", since, "user_id", userID, "bank_id", bankID)
		return status.Errorf(codes.Internal, "failed to parse Monzo transactions")
	}

	err = s.transactionService.SaveTransactions(ctx, trs)
	if err != nil {
		logger.ErrorWithFields("failed to save Monzo transactions", err, "since", since, "user_id", userID, "bank_id", bankID)
		return status.Errorf(codes.Internal, "failed to save Monzo transactions")
	}

	//todo
	if len(trErr) == 0 {
		return nil
	}

	logger.ErrorWithFields("failed to save few transactions", nil, "transactions_errors", trErr)
	return status.Errorf(codes.Internal, "failed to save some Monzo transactions")
}
