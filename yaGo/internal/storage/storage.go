package storage

import (
	"context"
	"errors"
)

var ErrURLExists = errors.New("url already exists")

type Repository interface {
	Save(code, longURL, userID string) error
	Get(code string) (string, bool)
	GetAllByUser(userID string) (map[string]string, error)
	GetByOriginalURL(ctx context.Context, originalURL string) (string, error) // <-- новый метод
	SaveBatch(ctx context.Context, records []URLRecord) error
	DeleteByUser(ctx context.Context, userID string) error
}
