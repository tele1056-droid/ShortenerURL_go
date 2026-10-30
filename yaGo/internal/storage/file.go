package storage

import (
	"bufio"
	"encoding/json"
	"os"
	"strconv"
	"sync"
)

type FileStorage struct {
	mu       sync.Mutex
	filePath string
	urls     map[string]string // short_url -> original_url
	counter  int               // для uuid
}

func NewFileStorage(filePath string) (*FileStorage, error) {
	fs := &FileStorage{
		filePath: filePath,
		urls:     make(map[string]string),
	}
	if err := fs.load(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	return fs, nil
}

func (s *FileStorage) load() error {
	file, err := os.Open(s.filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var rec URLRecord
		if err := json.Unmarshal(scanner.Bytes(), &rec); err != nil {
			continue // пропускаем битые строки
		}
		s.urls[rec.ShortURL] = rec.OriginalURL
		s.counter++ // восстанавливаем счётчик
	}
	return scanner.Err()
}

func (s *FileStorage) save(rec URLRecord) error {
	file, err := os.OpenFile(s.filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	data, err := json.Marshal(rec)
	if err != nil {
		return err
	}
	data = append(data, '\n')
	_, err = file.Write(data)
	return err
}

func (s *FileStorage) Save(code, longURL string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.urls[code] = longURL
	s.counter++
	rec := URLRecord{
		UUID:        strconv.Itoa(s.counter),
		ShortURL:    code,
		OriginalURL: longURL,
	}
	return s.save(rec)
}

func (s *FileStorage) Get(code string) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	longURL, ok := s.urls[code]
	return longURL, ok
}
