package storage

import (
	"encoding/json"
	"os"
	"sync"
)

type FileStorage struct {
	mu       sync.Mutex
	filePath string
	urls     map[string]string
}

func NewFileStorage(filePath string) (*FileStorage, error) {
	fs := &FileStorage{
		filePath: filePath,
		urls:     make(map[string]string),
	}

	// Загружаем данные при старте
	if err := fs.load(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	return fs, nil
}

func (s *FileStorage) load() error {
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &s.urls)
}

func (s *FileStorage) save() error {
	data, err := json.MarshalIndent(s.urls, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.filePath, data, 0644)
}

func (s *FileStorage) Save(code, longURL string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.urls[code] = longURL
	return s.save()
}

func (s *FileStorage) Get(code string) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	longURL, ok := s.urls[code]
	return longURL, ok
}
