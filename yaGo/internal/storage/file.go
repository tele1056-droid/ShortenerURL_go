package storage

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
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

func (s *FileStorage) DeleteByUser(ctx context.Context, userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 1. Удаляем из карты в памяти
	for code, rec := range s.urls {
		if rec.UserID == userID {
			delete(s.urls, code)
		}
	}

	// 2. Перезаписываем файл только оставшимися записями
	file, err := os.Create(s.filePath) // создает/перезаписывает файл
	if err != nil {
		return err
	}
	defer file.Close()

	// Проходим по оставшимся записям и пишем их в файл
	for _, rec := range s.urls {
		data, err := json.Marshal(rec)
		if err != nil {
			return err
		}
		data = append(data, '\n')
		if _, err := file.Write(data); err != nil {
			return err
		}
	}

	return nil
}

func (s *FileStorage) GetByOriginalURL(ctx context.Context, originalURL string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for code, rec := range s.urls {
		if rec.OriginalURL == originalURL {
			return code, nil
		}
	}
	return "", errors.New("url not found")
}

func (s *FileStorage) SaveBatch(ctx context.Context, records []URLRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, rec := range records {
		s.urls[rec.ShortURL] = rec
		if err := s.save(rec); err != nil {
			return err
		}
	}
	return nil
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

	// Проверяем, есть ли уже такая запись
	for _, rec := range s.urls {
		if rec.OriginalURL == longURL {
			return ErrURLExists
		}
	}

	// Если дубликата нет — сохраняем в карту и в файл
	s.counter++
	rec := URLRecord{
		UUID:        strconv.Itoa(s.counter),
		ShortURL:    code,
		OriginalURL: longURL,
		UserID:      userID,
	}
	s.urls[code] = rec
	return s.save(rec) // запись в файл
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
