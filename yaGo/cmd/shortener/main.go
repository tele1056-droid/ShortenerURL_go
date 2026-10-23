package main

import (
	"fmt"
	"net/http"

	"github.com/tele1056-droid/yaGo/internal/handlers"
	"github.com/tele1056-droid/yaGo/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/tele1056-droid/yaGo/internal/config"
)

func main() {
	// вариант с run() - более профессиональный
	if err := run(); err != nil {
		panic(err)
	}
}


func run() error {
	//1.a. Инициализируем конфигурацию (читаем флаги)
	cfg := config.NewConfig()

	//1. Создаем хранилище
	repo := storage.NewMemoryStorage()

	//2. Создаем хендлер, внедряя в него хранилище (DI)
	h := handlers.NewHandler(repo, cfg.BaseURL)

	//3. Регистрируем маршруты
	r := chi.NewRouter()
	r.Post("/", h.ShortenHandler)		// только POST
	r.Get("/{id}", h.RedirectHandler) 	// только GET с параметром
	r.Post("/api/shorten", h.JSONShortenHandler) // новый маршрут

	fmt.Printf("Сервер запущен на %s\n", cfg.ServerAddress)
	return http.ListenAndServe(cfg.ServerAddress, r)
}
