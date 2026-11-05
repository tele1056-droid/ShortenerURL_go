package storage

import "context"

type Repository interface {
	Save(code, longURL, userID string) error
	Get(code string) (string, bool)
	GetAllByUser(userID string) (map[string]string, error)
	SaveBatch(ctx context.Context, records []URLRecord) error
}
