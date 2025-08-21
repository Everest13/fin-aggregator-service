package uploader

import "sync"

type HeaderMappingStore struct {
	mu                sync.RWMutex
	bankHeaderMapping map[int64][]HeaderMapping
}

func NewHeaderMappingStore() *HeaderMappingStore {
	return &HeaderMappingStore{
		bankHeaderMapping: map[int64][]HeaderMapping{},
	}
}

func (s *HeaderMappingStore) Set(bankHeaderMapping map[int64][]HeaderMapping) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.bankHeaderMapping = bankHeaderMapping
}

func (s *HeaderMappingStore) GetByBank(bankID int64) []HeaderMapping {
	s.mu.Lock()
	defer s.mu.Unlock()

	if res, ok := s.bankHeaderMapping[bankID]; ok {
		return res
	}

	return nil
}
