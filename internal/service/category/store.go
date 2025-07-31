package category

import "sync"

type Store struct {
	mu              sync.RWMutex
	keywordCategory map[string]int64
	categories      []*Category //todo
}

func NewStore() *Store {
	return &Store{
		keywordCategory: map[string]int64{},
	}
}

func (s *Store) Reload(keywords []Keyword) {
	s.mu.Lock()
	defer s.mu.Unlock()

	keywordCategoryMap := map[string]int64{}
	for _, keyword := range keywords {
		keywordCategoryMap[keyword.Name] = keyword.CategoryID
	}

	s.keywordCategory = keywordCategoryMap
}

func (s *Store) GetKeywordCategoryMap() map[string]int64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.keywordCategory
}
