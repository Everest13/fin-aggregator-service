package monzo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/Everest13/fin-aggregator-service/internal/utils/logger"
	"io"
	"net/http"
	"net/url"
	"time"
)

const (
	authURL            = "https://auth.monzo.com"
	getAccountsURL     = "https://api.monzo.com/accounts"
	getTransactionsURL = "https://api.monzo.com/transactions"
	getAuthTokenURL    = "https://api.monzo.com/oauth2/token"

	authHeader        = "Authorization"
	bearerSchema      = "Bearer"
	contentTypeHeader = "Content-Type"

	clientIDField           = "client_id"
	redirectURIField        = "redirect_uri"
	responseTypeField       = "response_type"
	stateField              = "state"
	scopeField              = "scope"
	grantTypeField          = "grant_type"
	clientSecretField       = "client_secret"
	codeField               = "code"
	accountIDField          = "account_id"
	transactionsSinceField  = "since"
	transactionsBeforeField = "before"
	refreshTokenField       = "refresh_token"
)

const lenState = 32

type MonzoCfg struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
	StateLen     int
}

type client struct {
	monzoCfg   *MonzoCfg
	httpClient *http.Client
}

func newClient(timeout time.Duration, monzoCfg *MonzoCfg) *client {
	return &client{
		httpClient: &http.Client{
			Timeout: timeout,
		},
		monzoCfg: monzoCfg,
	}
}

type requestData struct {
	method  string
	url     string
	values  url.Values
	headers map[string]string
}

func (c *client) sendResponse(ctx context.Context, reqData requestData) (io.Reader, error) {
	req, err := http.NewRequestWithContext(ctx, reqData.method, reqData.url, bytes.NewBufferString(reqData.values.Encode()))
	if err != nil {
		logger.ErrorWithFields("failed to create request", err, "req_data", reqData)
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	for key, value := range reqData.headers {
		req.Header.Set(key, value)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		logger.ErrorWithFields("failed to send request", err, "url", req.URL.String(), "method", req.Method)
		return nil, fmt.Errorf("failed to send request")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.ErrorWithFields("failed to read response body", err, "url", req.URL.String(), "status_code", resp.StatusCode)
		return nil, fmt.Errorf("failed to read response body")
	}

	if resp.StatusCode != http.StatusOK {
		logger.ErrorWithFields("monzo API error", fmt.Errorf("unexpected status code"), "url", req.URL.String(), "status_code", resp.StatusCode, "body", string(body))
		return nil, fmt.Errorf("monzo API error")
	}

	return bytes.NewReader(body), nil
}

func (c *client) generateAuthURL(state string) string {
	cfg := c.monzoCfg
	v := url.Values{}
	v.Set(clientIDField, cfg.ClientID)
	v.Set(redirectURIField, cfg.RedirectURI)
	v.Set(responseTypeField, "code")
	v.Set(stateField, state)
	v.Set(scopeField, "read:accounts read:transactions")

	return authURL + "?" + v.Encode()
}

func (c *client) getTokens(ctx context.Context, code string) (*tokenResponse, error) {
	cfg := c.monzoCfg
	values := url.Values{}
	values.Set(grantTypeField, "authorization_code")
	values.Set(clientIDField, cfg.ClientID)
	values.Set(clientSecretField, cfg.ClientSecret)
	values.Set(redirectURIField, cfg.RedirectURI)
	values.Set(codeField, code)
	values.Set(scopeField, "read:accounts read:transactions")

	reqData := requestData{
		method: http.MethodPost,
		url:    getAuthTokenURL,
		values: values,
		headers: map[string]string{
			contentTypeHeader: "application/x-www-form-urlencoded",
		},
	}

	body, err := c.sendResponse(ctx, reqData)
	if err != nil {
		logger.Error("failed to send account id request", err)
		return nil, err
	}

	var tokenResp tokenResponse
	if err = json.NewDecoder(body).Decode(&tokenResp); err != nil {
		logger.Error("failed to decode token response", err)
		return nil, fmt.Errorf("failed to decode token response: %w", err)
	}

	return &tokenResp, nil
}

func (c *client) getAccountID(ctx context.Context, accessToken string) (string, error) {
	reqData := requestData{
		method: http.MethodGet,
		url:    getAccountsURL,
		headers: map[string]string{
			authHeader: bearerSchema + " " + accessToken,
		},
	}

	body, err := c.sendResponse(ctx, reqData)
	if err != nil {
		logger.Error("failed to send account id request", err)
		return "", err
	}

	var data struct {
		Accounts []struct {
			ID string `json:"id"`
		} `json:"accounts"`
	}

	if err = json.NewDecoder(body).Decode(&data); err != nil {
		logger.Error("failed to decode account response", err)
		return "", fmt.Errorf("failed to decode account response: %w", err)
	}

	if len(data.Accounts) == 0 {
		logger.Error("no accounts found", err)
		return "", fmt.Errorf("no accounts found")
	}

	return data.Accounts[0].ID, nil
}

func (c *client) getMonzoTransactions(ctx context.Context, accessToken string, accountID string, since, before time.Time) ([]MonzoTransaction, error) {
	values := url.Values{}
	values.Set(accountIDField, accountID)
	values.Set(transactionsSinceField, since.Format(time.RFC3339))
	values.Set(transactionsBeforeField, before.Format(time.RFC3339))

	reqURL := fmt.Sprintf("%s?%s", getTransactionsURL, values.Encode())
	reqData := requestData{
		method: http.MethodGet,
		url:    reqURL,
		values: values,
		headers: map[string]string{
			authHeader: bearerSchema + " " + accessToken,
		},
	}

	body, err := c.sendResponse(ctx, reqData)
	if err != nil {
		logger.Error("failed to send transactions request", err)
		return nil, err
	}

	var monzoResp monzoTransactionsResponse
	if err = json.NewDecoder(body).Decode(&monzoResp); err != nil {
		logger.Error("failed to decode account response", err)
		return nil, fmt.Errorf("failed to decode account response: %w", err)
	}

	return monzoResp.Transactions, nil
}

func (c *client) refreshToken(ctx context.Context, refreshToken string) (*tokenResponse, error) {
	cfg := c.monzoCfg
	values := url.Values{}
	values.Set(grantTypeField, "refresh_token")
	values.Set(clientIDField, cfg.ClientID)
	values.Set(clientSecretField, cfg.ClientSecret)
	values.Set(refreshTokenField, refreshToken)

	reqData := requestData{
		method: http.MethodPost,
		url:    getAuthTokenURL,
		values: values,
		headers: map[string]string{
			contentTypeHeader: "application/x-www-form-urlencoded",
		},
	}

	body, err := c.sendResponse(ctx, reqData)
	if err != nil {
		logger.Error("failed to send token request", err)
		return nil, err
	}

	var tr tokenResponse
	if err = json.NewDecoder(body).Decode(&tr); err != nil {
		logger.Error("failed to decode token response", err)
		return nil, fmt.Errorf("failed to decode token response: %w", err)
	}

	return &tr, nil
}
