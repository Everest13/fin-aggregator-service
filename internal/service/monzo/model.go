package monzo

import (
	"time"
)

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	TokenType    string `json:"token_type"`
	UserID       string `json:"user_id"`
	AccountID    string `json:"account_id,omitempty"`
}

type authTokens struct {
	accessToken  string
	refreshToken string
	expiresIn    int64
	issuedAt     time.Time
	accountID    string
}

func (a *authTokens) IsExpired() bool {
	return time.Now().After(a.issuedAt.Add(time.Duration(a.expiresIn) * time.Second))
}

type monzoTransactionsResponse struct {
	Transactions []MonzoTransaction `json:"transactions"`
}

type MonzoTransaction struct {
	ID          string `json:"id"`
	AccountID   string `json:"account_id"`
	Amount      int64  `json:"amount"`
	Description string `json:"description"`
	Created     string `json:"created"`
	Category    string `json:"category"`
	Notes       string `json:"notes"`
	Scheme      string `json:"scheme"`
}
