package storage

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tele1056-droid/yaGo/internal/shortener"
)

type PostgresStorage struct {
	pool *pgxpool.Pool
}

func (s *PostgresStorage) GetByOriginalURL(ctx context.Context, originalURL string) (string, error) {
	var shortURL string
	err := s.pool.QueryRow(ctx, `SELECT short_url FROM urls WHERE original_url = $1 AND deleted = FALSE`, originalURL).Scan(&shortURL)
	if err != nil {
		return "", err
	}
	return shortURL, nil
}

func (s *PostgresStorage) DeleteByUser(ctx context.Context, userID string) error {
	_, err := s.pool.Exec(ctx, `UPDATE urls SET deleted = TRUE WHERE user_id = $1`, userID)
	return err
}

func (s *PostgresStorage) GetMaxShortURL(ctx context.Context) (string, error) {
	rows, err := s.pool.Query(ctx, `SELECT short_url FROM urls WHERE deleted = FALSE`)
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
	query := `
	CREATE TABLE IF NOT EXISTS urls (
		id SERIAL PRIMARY KEY,
		short_url TEXT NOT NULL UNIQUE,
		original_url TEXT NOT NULL UNIQUE,
		user_id TEXT,
		deleted BOOLEAN DEFAULT FALSE
	);`

	if _, err := pool.Exec(ctx, query); err != nil {
		return nil, err
	}

	return &PostgresStorage{pool: pool}, nil
}

func (s *PostgresStorage) Save(code, longURL, userID string) error {
	_, err := s.pool.Exec(context.Background(),
		`INSERT INTO urls (short_url, original_url, user_id) VALUES ($1, $2, $3)`,
		code, longURL, userID,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			if strings.Contains(pgErr.ConstraintName, "original_url") {
				return ErrURLExists
			}
			return fmt.Errorf("конфликт short_url: %w", err)
		}
		return err
	}
	return nil
}

// Get возвращает URL и флаг «найдено». Если URL удалён, возвращает false.
func (s *PostgresStorage) Get(code string) (string, bool) {
	var originalURL string
	var deleted bool
	query := `SELECT original_url, deleted FROM urls WHERE short_url = $1`
	err := s.pool.QueryRow(context.Background(), query, code).Scan(&originalURL, &deleted)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", false // никогда не существовал
		}
		return "", false
	}
	if deleted {
		return "", false // существовал, но удалён
	}
	return originalURL, true
}

// IsDeleted проверяет, был ли URL удалён (для возврата 410 Gone).
func (s *PostgresStorage) IsDeleted(code string) bool {
	var deleted bool
	query := `SELECT deleted FROM urls WHERE short_url = $1`
	err := s.pool.QueryRow(context.Background(), query, code).Scan(&deleted)
	if err != nil {
		return false // не найден
	}
	return deleted
}

// GetAllByUser возвращает только НЕудалённые URL пользователя.
func (s *PostgresStorage) GetAllByUser(userID string) (map[string]string, error) {
	query := `SELECT short_url, original_url FROM urls WHERE user_id = $1 AND deleted = FALSE`
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
			`INSERT INTO urls (short_url, original_url, user_id) VALUES ($1, $2, $3)`,
			rec.ShortURL, rec.OriginalURL, rec.UserID,
		)
		if err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}
