package storage

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tele1056-droid/yaGo/internal/shortener"
)

type PostgresStorage struct {
	pool *pgxpool.Pool
}

func (s *PostgresStorage) GetMaxShortURL(ctx context.Context) (string, error) {
	// Мы не можем использовать MAX(), потому что строки сортируются лексикографически, а не по числовому значению.
	// Поэтому нужно получить все short_url и найти максимум в Go, либо использовать сортировку по длине, а затем лексикографически.
	// Простой, но не самый эффективный способ: получить все и найти максимум в Go.
	// Для учебного проекта это допустимо.
	rows, err := s.pool.Query(ctx, `SELECT short_url FROM urls`)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	var maxCode string
	var maxVal int64 = -1
	for rows.Next() {
		var code string
		if err := rows.Scan(&code); err != nil {
			return "", err
		}
		val := shortener.FromBase62(code)
		if val > maxVal {
			maxVal = val
			maxCode = code
		}
	}
	return maxCode, rows.Err()
}

func NewPostgresStorage(ctx context.Context, pool *pgxpool.Pool) (*PostgresStorage, error) {
	// 1. Создаем таблицу, если её нет
	query := `
	CREATE TABLE IF NOT EXISTS urls (
			uuid TEXT PRIMARY KEY,
			short_url TEXT NOT NULL UNIQUE,
			original_url TEXT NOT NULL,
			user_id TEXT
	);`

	if _, err := pool.Exec(ctx, query); err != nil {
		return nil, err
	}

	return &PostgresStorage{pool: pool}, nil
}

func (s *PostgresStorage) Save(code, longURL, userID string) error {
	// В задании uuid генерируется, но для простоты можно использовать code как uuid
	// или генерировать UUID через библиотеку google/uuid
	query := `INSERT INTO urls (uuid, short_url, original_url, user_id) VALUES ($1,$2, $3, $4)`
	_, err := s.pool.Exec(context.Background(), query, code, code, longURL, userID)
	return err
}

func (s *PostgresStorage) Get(code string) (string, bool) {
	var originalURL string
	query := `SELECT original_url FROM urls WHERE short_url = $1`
	err := s.pool.QueryRow(context.Background(), query, code).Scan(&originalURL)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", false
		}
		return "", false
	}
	return originalURL, true
}

func (s *PostgresStorage) GetAllByUser(userID string) (map[string]string, error) {
	query := `SELECT short_url, original_url FROM urls WHERE user_id = $1`
	rows, err := s.pool.Query(context.Background(), query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]string)
	for rows.Next() {
		var short, original string
		if err := rows.Scan(&short, &original); err != nil {
			return nil, err
		}
		result[short] = original
	}
	return result, nil
}

func (s *PostgresStorage) SaveBatch(ctx context.Context, records []URLRecord) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for _, rec := range records {
		_, err := tx.Exec(ctx,
			`INSERT INTO urls (uuid, short_url, original_url, user_id) VALUES ($1, $2, $3, $4)`,
			rec.UUID, rec.ShortURL, rec.OriginalURL, rec.UserID,
		)
		if err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}
