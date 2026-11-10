package storage

import (
	"context"
	"errors"
	"sync"
)

type MemoryStorage struct {
	mu   sync.Mutex
	urls map[string]URLRecord // теперь храним записи, а не просто строки
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{urls: make(map[string]URLRecord)}
}

func (s *MemoryStorage) IsDeleted(code string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	rec, ok := s.urls[code]
	if !ok {
		return false
	}
	return rec.Deleted
}

func (s *MemoryStorage) DeleteByUser(ctx context.Context, userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for code, rec := range s.urls {
		if rec.UserID == userID {
			rec.Deleted = true
			s.urls[code] = rec
		}
	}
	return nil
}

func (s *MemoryStorage) GetByOriginalURL(ctx context.Context, originalURL string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for code, rec := range s.urls {
		if rec.OriginalURL == originalURL {
			return code, nil
		}
	}
	return "", errors.New("url not found")
}

func (s *MemoryStorage) Save(code, longURL, userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Проверяем, есть ли уже такая запись
	for _, rec := range s.urls {
		if rec.OriginalURL == longURL {
			return ErrURLExists // возвращаем ошибку, хендлер вернёт 409
		}
	}

	// Если дубликата нет — сохраняем
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
	if !ok || rec.Deleted {
		return "", false
	}
	return rec.OriginalURL, true
}

func (s *MemoryStorage) GetAllByUser(userID string) (map[string]string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	result := make(map[string]string)
	for code, rec := range s.urls {
		if rec.UserID == userID && !rec.Deleted {
			result[code] = rec.OriginalURL
		}
	}
	return result, nil
}

func (s *MemoryStorage) SaveBatch(ctx context.Context, records []URLRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, rec := range records {
		s.urls[rec.ShortURL] = rec
	}
	return nil
}
