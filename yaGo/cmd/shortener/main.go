package main

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/tele1056-droid/yaGo/internal/config"
	"github.com/tele1056-droid/yaGo/internal/handlers"
	"github.com/tele1056-droid/yaGo/internal/middleware"
	"github.com/tele1056-droid/yaGo/internal/shortener" // добав. импорт
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

	if cfg.FileStoragePath != "" {
		// Если путь к файлу задан — используем файловое хранилище
		// 2. Сохраняем указатель на конкретный тип FileStorage
		fs, err := storage.NewFileStorage(cfg.FileStoragePath)
		if err != nil {
			return fmt.Errorf("не удалось создать файловое хранилище: %w", err)
		}

		repo = fs
		// 3. Синхронизируем счётчик shortener с максимальным uuid из файла
		shortener.SetCounter(fs.GetMaxUUID())

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

	//iter7.a. Сначала разжимаем входящий запрос
	r.Use(middleware.DecompressMiddleware)
	// Логируем
	r.Use(middleware.LoggerMiddleware(logger))

	//iter7.b. Сжимаем ответ (если клиент поддерживает)
	r.Use(middleware.GzipMiddleware)

	r.Post("/", h.ShortenHandler)
	r.Get("/{id}", h.RedirectHandler)
	r.Post("/api/shorten", h.JSONShortenHandler)
	r.Get("/api/user/urls", h.GetUserURLsHandler)

	fmt.Printf("Сервер запущен на %s\n", cfg.ServerAddress)
	return http.ListenAndServe(cfg.ServerAddress, r)
}
