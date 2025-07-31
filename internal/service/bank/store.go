package bank

import "sync"

type Store struct {
	mu    sync.RWMutex
	banks map[int64]*Bank
}

func NewStore() *Store {
	return &Store{
		banks: map[int64]*Bank{},
	}
}

func (s *Store) Reload(banks []Bank) {
	s.mu.Lock()
	defer s.mu.Unlock()

	res := make(map[int64]*Bank, len(banks))
	for _, bank := range banks {
		res[bank.ID] = &bank
	}

	s.banks = res
}

func (s *Store) GetBankMap() map[int64]*Bank {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.banks
}
