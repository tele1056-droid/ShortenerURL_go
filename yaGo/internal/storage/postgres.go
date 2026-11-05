package storage

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresStorage struct {
	pool *pgxpool.Pool
}

func (s *PostgresStorage) GetMaxUUID(ctx context.Context) (int, error) {
	var maxID int
	// Предполагаем, что uuid — это число в виде текста, или используйте MAX(short_url), если код числовой
	query := `SELECT COALESCE(MAX(CAST(uuid AS INTEGER)), 0) FROM urls`
	err := s.pool.QueryRow(ctx, query).Scan(&maxID)
	return maxID, err

	/*
		(Если uuid — это строка, а short_url — числовой код, лучше использовать MAX(CAST(short_url AS INTEGER)), если коды числовые).
	*/
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
