package main

import (
	"fmt"
	"net/http"

	"github.com/tele1056-droid/yaGo/internal/handlers"
)

func main() {
	// вариант с run() - более профессиональный
	if err := run(); err != nil {
		panic(err)
	}
}

func mainHandler(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.URL.Path == "/" {
			handlers.ShortenHandler(w, r)
		} else if r.Method == http.MethodGet && r.URL.Path != "/" {
			// сюда мы попадем если будет и метод GET и путь не равен корню "/" (т.е. "/EwHXdfg" - это не "/")
			handlers.RedirectHandler(w, r)
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}

func run() error {
	// mux := http.NewServeMux()
	// mux.HandleFunc("/", handlers.ShortenHandler)

	mux := http.NewServeMux()
	mux.HandleFunc("/", mainHandler)

	fmt.Println("Сервер запущен на :8080")
	return http.ListenAndServe(":8080", mux)
}


/*
При регистарции маршрута для GET (mux.HandleFunc("/{id}", handlers.RedirectHandler)) возникает ПРОБЛЕМА: http.ServeMux не поддерживает шаблоны вида /{id}. Он понимает только префиксы. Поэтому /{id} не сработает — всё уйдёт в /.

Для решения либо либо:
1. Используя стандартную библиотеку: проверять метод и путь внутри одного хендлера: для этого сделаем func mainHandler()

2. Либо использовать стороние библиотеки, например "github.com/go-chi/chi/v5"
*/
