package storage

import "sync"

type MemoryStorage struct {
	mu   sync.Mutex
	urls map[string]string
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{urls: make(map[string]string)}
}

func (s *MemoryStorage) Save(code, longURL string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.urls[code] = longURL
	return nil
}

func (s *MemoryStorage) Get(code string) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	longURL, ok := s.urls[code]
	return longURL, ok
}

func (s *MemoryStorage) GetAll() map[string]string {
	s.mu.Lock()
	defer s.mu.Unlock()
	// Возвращаем копию, чтобы не было гонок
	result := make(map[string]string, len(s.urls))
	for k, v := range s.urls {
		result[k] = v
	}
	return result
}
