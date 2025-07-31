package monzo

import (
	"sync"
)

type tokenStore struct {
	mu         sync.RWMutex
	authTokens *authTokens
}

func (t *tokenStore) set(authTokens *authTokens) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.authTokens = authTokens
}

func (t *tokenStore) get() *authTokens {
	t.mu.RLock()
	defer t.mu.RUnlock()

	return t.authTokens
}

func (t *tokenStore) delete() {
	t.mu.RLock()
	defer t.mu.RUnlock()

	t.authTokens = nil
}
