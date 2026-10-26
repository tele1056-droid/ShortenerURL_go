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
	// a.1 создаем логгер
	logger, _ := zap.NewProduction()
	defer logger.Sync() // Важно: сбрасывает буфер логов при выходе

	//1.a. Инициализируем конфигурацию (читаем флаги)
	cfg := config.NewConfig()

	//1. Создаем хранилище
	repo := storage.NewMemoryStorage()

	//2. Создаем хендлер, внедряя в него хранилище (DI)
	h := handlers.NewHandler(repo, cfg.BaseURL)

	//3. Регистрируем маршруты
	r := chi.NewRouter()

	//Подключаем middlware глобально для всех запросов
	r.Use(middleware.LoggerMiddleware(logger))

	r.Post("/", h.ShortenHandler)                // только POST
	r.Get("/{id}", h.RedirectHandler)            // только GET с параметром
	r.Post("/api/shorten", h.JSONShortenHandler) // новый маршрут

	fmt.Printf("Сервер запущен на %s\n", cfg.ServerAddress)
	return http.ListenAndServe(cfg.ServerAddress, r)
}
