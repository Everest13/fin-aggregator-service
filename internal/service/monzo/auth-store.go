package monzo

import (
	"sync"
)

type authStore struct {
	mu         sync.RWMutex
	authTokens *authTokens
	accountID  string
	state      string
}

func (a *authStore) getState() string {
	a.mu.Lock()
	defer a.mu.Unlock()

	return a.state
}

func (a *authStore) setState(state string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.state = state
}

func (a *authStore) deleteState() {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.state = ""
}

func (a *authStore) setAuthToken(authTokens *authTokens) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.authTokens = authTokens
}

func (a *authStore) getAuthToken() *authTokens {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.authTokens
}

func (a *authStore) setAccountID(id string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.accountID = id
}

func (a *authStore) getAccountID() string {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.accountID
}
