package middleware

import (
	"compress/gzip"
	"net/http"
)

func DecompressMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1. Проверяем, сжат ли запрос
		if r.Header.Get("Content-Encoding") == "gzip" {
			// 2. Создаем gzip.Reader из тела запроса
			reader, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, "Ошибка распаковки", http.StatusBadRequest)
				return
			}
			defer reader.Close()

			// 3. Подменяем тело запроса на распакованное
			r.Body = reader
			// Удаляем заголовок, так как тело теперь обычное
			r.Header.Del("Content-Encoding")
		}

		// 4. Передаем управление дальше
		next.ServeHTTP(w, r)
	})
}
