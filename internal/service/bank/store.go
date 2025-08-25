package bank

import "sync"

type Store struct {
	mu       sync.RWMutex
	bankMap  map[int64]Bank
	bankList []Bank
}

func NewStore() *Store {
	return &Store{
		bankMap:  map[int64]Bank{},
		bankList: []Bank{},
	}
}

func (s *Store) setBankMap(bankMap map[int64]Bank) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.bankMap = bankMap
}

func (s *Store) setBankList(bankList []Bank) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.bankList = bankList
}

func (s *Store) getBankMap() map[int64]Bank {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.bankMap
}

func (s *Store) getBank(id int64) *Bank {
	s.mu.Lock()
	defer s.mu.Unlock()

	if bank, ok := s.bankMap[id]; ok {
		return &bank
	}

	return nil
}

func (s *Store) getBankList() []Bank {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.bankList
}
