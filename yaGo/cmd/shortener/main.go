package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tele1056-droid/yaGo/internal/config"
	"github.com/tele1056-droid/yaGo/internal/handlers"
	"github.com/tele1056-droid/yaGo/internal/middleware"
	"github.com/tele1056-droid/yaGo/internal/shortener"
	"github.com/tele1056-droid/yaGo/internal/storage"
	"go.uber.org/zap"
)

func main() {
	if err := run(); err != nil {
		fmt.Printf("Ошибка: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// 1. Конфигурация
	cfg := config.NewConfig()

	// 2. Логгер
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// 3. Хранилище (Postgres > File > Memory)
	var repo storage.Repository
	var pool *pgxpool.Pool

	if cfg.DatabaseDSN != "" {
		var err error
		pool, err = pgxpool.New(context.Background(), cfg.DatabaseDSN)
		if err != nil {
			return fmt.Errorf("не удалось создать пул соединений: %w", err)
		}
		defer pool.Close()

		if err := pool.Ping(context.Background()); err != nil {
			return fmt.Errorf("БД недоступна: %w", err)
		}

		pgRepo, err := storage.NewPostgresStorage(context.Background(), pool)
		if err != nil {
			return fmt.Errorf("не удалось инициализировать Postgres: %w", err)
		}
		repo = pgRepo
		logger.Info("используется PostgreSQL")

		maxID, err := pgRepo.GetMaxUUID(context.Background())
		if err != nil {
			return fmt.Errorf("не удалось получить максимальный ID из БД: %w", err)
		}
		shortener.SetCounter(maxID) // Устанавливаем счётчик shortener на актуальное значение
		logger.Info("счётчик синхронизирован с БД", zap.Int("maxID", maxID))

	} else if cfg.FileStoragePath != "" {
		fs, err := storage.NewFileStorage(cfg.FileStoragePath)
		if err != nil {
			return fmt.Errorf("не удалось создать файловое хранилище: %w", err)
		}
		repo = fs
		shortener.SetCounter(fs.GetMaxUUID())
		logger.Info("используется файловое хранилище", zap.String("path", cfg.FileStoragePath))

	} else {
		repo = storage.NewMemoryStorage()
		logger.Info("используется хранилище в памяти")
	}

	// 4. Хендлер
	h := handlers.NewHandler(repo, cfg.BaseURL, pool)

	// 5. Роутер
	r := chi.NewRouter()
	r.Use(middleware.DecompressMiddleware)
	r.Use(middleware.LoggerMiddleware(logger))
	r.Use(middleware.GzipMiddleware)

	r.Post("/", h.ShortenHandler)
	r.Get("/{id}", h.RedirectHandler)
	r.Post("/api/shorten", h.JSONShortenHandler)
	r.Get("/api/user/urls", h.GetUserURLsHandler)
	r.Get("/ping", h.PingHandler)

	// 6. HTTP-сервер
	server := &http.Server{
		Addr:    cfg.ServerAddress,
		Handler: r,
	}

	// 7. Graceful shutdown
	go func() {
		fmt.Printf("Сервер запущен на %s\n", cfg.ServerAddress)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Ошибка сервера", zap.Error(err))
		}
	}()

	// Ждем сигнал
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Завершение работы сервера...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return server.Shutdown(ctx)
}
