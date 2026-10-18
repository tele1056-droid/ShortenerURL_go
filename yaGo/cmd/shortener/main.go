package main

import (
	"fmt"
	"net/http"

	"github.com/tele1056-droid/yaGo/internal/handlers"
	"github.com/tele1056-droid/yaGo/internal/storage"
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
	mux := http.NewServeMux()
	mux.HandleFunc("/", mainHandler(h))

	fmt.Println("Сервер запущен на :8080")
	return http.ListenAndServe(":8080", mux)
}

/*
	Инекр.2
Это и есть DI: main собирает всё вместе — хранилище, хендлер, роутер. Хендлеры больше не знают, какое хранилище используется. Это делает код тестируемым и гибким.
*/

/*
При регистарции маршрута для GET (mux.HandleFunc("/{id}", handlers.RedirectHandler)) возникает ПРОБЛЕМА: http.ServeMux не поддерживает шаблоны вида /{id}. Он понимает только префиксы. Поэтому /{id} не сработает — всё уйдёт в /.

Для решения либо либо:
1. Используя стандартную библиотеку: проверять метод и путь внутри одного хендлера: для этого сделаем func mainHandler()

2. Либо использовать стороние библиотеки, например "github.com/go-chi/chi/v5"
*/
