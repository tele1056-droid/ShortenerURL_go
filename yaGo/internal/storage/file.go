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
	urls     map[string]URLRecord // теперь храним записи, а не просто строки
	counter  int                  // для uuid
}

func NewFileStorage(filePath string) (*FileStorage, error) {
	fs := &FileStorage{
		filePath: filePath,
		urls:     make(map[string]URLRecord),
	}
	if err := fs.load(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	return fs, nil
}

func (s *FileStorage) GetMaxUUID() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.counter // предполагается, что counter обновляется в load()
}

func (s *FileStorage) GetAllByUser(userID string) (map[string]string, error) {
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
			continue
		}
		s.urls[rec.ShortURL] = rec

		// Обновляем счётчик, чтобы он был не меньше uuid
		if id, err := strconv.Atoi(rec.UUID); err == nil {
			if id > s.counter {
				s.counter = id
			}
		}
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

func (s *FileStorage) Save(code, longURL, userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.counter++
	rec := URLRecord{
		UUID:        strconv.Itoa(s.counter),
		ShortURL:    code,
		OriginalURL: longURL,
		UserID:      userID,
	}
	s.urls[code] = rec
	return s.save(rec)
}

func (s *FileStorage) Get(code string) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	rec, ok := s.urls[code]
	if !ok {
		return "", false
	}

	return rec.OriginalURL, true
}
