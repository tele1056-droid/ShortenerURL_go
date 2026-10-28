package middleware

import (
	"compress/gzip"
	"fmt"
	"net/http"
)

func DecompressMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("DEBUG: Content-Encoding = %q, Content-Type = %q\n",
			r.Header.Get("Content-Encoding"), r.Header.Get("Content-Type"))

		// 1. Проверяем, сжат ли запрос
		if r.Header.Get("Content-Encoding") == "gzip" {
			// 2. Создаем gzip.Reader из тела запроса, распаковка
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
			r.Header.Set("Content-Type", "text/plain")

			fmt.Println("DEBUG: распаковали, установили Content-Type: text/plain")
		}

		// 4. Передаем управление дальше
		next.ServeHTTP(w, r)
	})
}
