package main

import (
	"fmt"
	"net/http"

	"github.com/tele1056-droid/yaGo/internal/handlers"
	"github.com/tele1056-droid/yaGo/internal/storage"
	"github.com/go-chi/chi/v5"
)

func main() {
	// вариант с run() - более профессиональный
	if err := run(); err != nil {
		panic(err)
	}
}

// оставим вариант с mainHandler, который будет маршрутизировать на хендлеры
func mainHandler(h *handlers.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.URL.Path == "/" {
			h.ShortenHandler(w, r)
		} else if r.Method == http.MethodGet && r.URL.Path != "/" {
			h.RedirectHandler(w, r)
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}
}

func run() error {
	//1. Создаем хранилище
	repo := storage.NewMemoryStorage()

	//2. Создаем хендлер, внедряя в него хранилище (DI)
	h := handlers.NewHandler(repo)

	//3. Регистрируем маршруты
	r := chi.NewRouter()
	r.Post("/", h.ShortenHandler)		// только POST
	r.Get("/{id}", h.RedirectHandler) 	// только GET с параметром

	fmt.Println("Сервер запущен на :8080")
	return http.ListenAndServe(":8080", r)
}
