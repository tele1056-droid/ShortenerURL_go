package storage

import "sync"

type MemoryStorage struct {
	mu   sync.Mutex
	urls map[string]URLRecord // теперь храним записи, а не просто строки
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{urls: make(map[string]URLRecord)}
}

func (s *MemoryStorage) Save(code, longURL, userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.urls[code] = URLRecord{
		ShortURL:    code,
		OriginalURL: longURL,
		UserID:      userID,
	}
	return nil
}

func (s *MemoryStorage) Get(code string) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	rec, ok := s.urls[code]
	if !ok {
		return "", false
	}
	return rec.OriginalURL, true
}

func (s *MemoryStorage) GetAllByUser(userID string) (map[string]string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	result := make(map[string]string)
	for code, rec := range s.urls {
		if rec.UserID == userID {
			result[code] = rec.OriginalURL
		}
	}
	return result, nil
}
