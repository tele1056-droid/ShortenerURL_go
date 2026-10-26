package main

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/tele1056-droid/yaGo/internal/config"
	"github.com/tele1056-droid/yaGo/internal/handlers"
	"github.com/tele1056-droid/yaGo/internal/middleware"
	"github.com/tele1056-droid/yaGo/internal/storage"
	"go.uber.org/zap"
)

func main() {
	// вариант с run() - более профессиональный
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	// 1. Инициализируем конфигурацию (читаем флаги и env)
	cfg := config.NewConfig()

	// 2. Создаём логгер
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// 3. Создаём хранилище (файловое или в памяти)
	var repo storage.Repository
	var err error

	if cfg.FileStoragePath != "" {
		// Если путь к файлу задан — используем файловое хранилище
		repo, err = storage.NewFileStorage(cfg.FileStoragePath)
		if err != nil {
			return fmt.Errorf("не удалось создать файловое хранилище: %w", err)
		}
		logger.Info("используется файловое хранилище", zap.String("path", cfg.FileStoragePath))
	} else {
		// Иначе — в памяти
		repo = storage.NewMemoryStorage()
		logger.Info("используется хранилище в памяти")
	}

	// 4. Создаём хендлер, внедряя хранилище (DI)
	h := handlers.NewHandler(repo, cfg.BaseURL)

	// 5. Регистрируем маршруты
	r := chi.NewRouter()
	r.Use(middleware.LoggerMiddleware(logger))

	r.Post("/", h.ShortenHandler)
	r.Get("/{id}", h.RedirectHandler)
	r.Post("/api/shorten", h.JSONShortenHandler)

	fmt.Printf("Сервер запущен на %s\n", cfg.ServerAddress)
	return http.ListenAndServe(cfg.ServerAddress, r)
}
