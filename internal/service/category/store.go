package category

import "sync"

type Store struct {
	mu                sync.RWMutex
	keywordCategoryID map[string]int64
	categories        map[int64]Category
}

func NewStore() *Store {
	return &Store{
		keywordCategoryID: map[string]int64{},
	}
}

func (s *Store) ReloadKeywordCategoryIDMap(keywords []Keyword) {
	s.mu.Lock()
	defer s.mu.Unlock()

	keywordCategoryMap := map[string]int64{}
	for _, keyword := range keywords {
		keywordCategoryMap[keyword.Name] = keyword.CategoryID
	}

	s.keywordCategoryID = keywordCategoryMap
}

func (s *Store) ReloadCategoriesMap(categories []Category) {
	s.mu.Lock()
	defer s.mu.Unlock()

	categoriesMap := make(map[int64]Category, len(categories))
	for _, category := range categories {
		categoriesMap[category.ID] = category
	}

	s.categories = categoriesMap
}

func (s *Store) GetKeywordCategoryIDMap() map[string]int64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.keywordCategoryID
}

func (s *Store) GetCategory(id int64) *Category {
	s.mu.Lock()
	defer s.mu.Unlock()

	category := s.categories[id]

	return &category
}
