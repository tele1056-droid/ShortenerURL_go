package handlers

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/tele1056-droid/yaGo/internal/shortener"
	"github.com/tele1056-droid/yaGo/internal/storage"
)

type Handler struct {
	repo storage.Repository
}

func NewHandler(repo storage.Repository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) ShortenHandler(w http.ResponseWriter, r *http.Request) {
	// 1. проверка метода
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	// 2. Проверка Content-Type
	contentType := r.Header.Get("Content-Type")
	if contentType != "text/plain" {
		http.Error(w, "Ожидается Content-Type: text/plain", http.StatusUnsupportedMediaType)
		return
	}

	// 3. Ограничение размера тела (защита от гигантских запросов)
	r.Body = http.MaxBytesReader(w, r.Body, 1024*1024) // 1 Мб

	// 4. Чтение тела, по условию url будет в теле запроса (r.Body)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Ошибка чтения тела", http.StatusBadRequest)
		return
	}

	// 5. Проверка, что тело не пустое
	if len(body) == 0 {
		http.Error(w, "Пустое тело запроса", http.StatusBadRequest)
		return
	}

	// 6. Обрезка пробелов и переводов строк
	longURL := strings.TrimSpace(string(body))

	// 7. Повторная проверка после обрезки
	if longURL == "" {
		http.Error(w, "URL не может быть пустым", http.StatusBadRequest)
		return
	}

	// 8. Проверка, что URL начинается с http:// или https://
	if !strings.HasPrefix(longURL, "http://") && !strings.HasPrefix(longURL, "https://") {
		http.Error(w, "URL должен начинаться с http:// или https://", http.StatusBadRequest)
		return
	}

	// 9. Проверка, что URL валидный
	parsedURL, err := url.ParseRequestURI(longURL)
	if err != nil || parsedURL.Host == "" {
		http.Error(w, "Некорректный URL", http.StatusBadRequest)
		return
	}

	// 10. Генерация кода
	code := shortener.Shorten()

	// Сохраняем через интерфейс
	h.repo.Save(code, longURL)

	// 11. Установка заголовка и кода ответа
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)

	// 12. Запись ответа
	fmt.Fprintf(w, "http://localhost:8080/%s", code)
}

func (h *Handler) RedirectHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Проверка метода
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// 2. Извлечение кода из пути
	//для начала нужно проверить не пустой ли r.URL.Path (""), если он пустой то при дальнейшей проверке будет вызвана паника и программа упадёт
	if len(r.URL.Path) < 2 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	code := r.URL.Path[1:] // тут мы из пути, т.е. то что в URL начиная с "/", и мы обрезаем с символа под индексом 1 и до конца, те символ "/" с индексом 0 не попадет в наш code

	// 3. Поиск URL в хранилище
	longURL, ok := h.repo.Get(code)
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// 4. Установка заголовка Location и кода 307
	w.Header().Set("Location", longURL)
	w.WriteHeader(http.StatusTemporaryRedirect)

	/* Как работает редирект:
	1. Клиент(браузер) отправляет GET запрос http://localhost:8080/EwHXdJfB
	2. Обработчик срабатывает на этот запрос и находит какой URL соответствует коду который передали в запросе
	3. Сервер отвечает с кодом 307 Temporary Redirect и Location: с URL
	4. Клиент (браузер) получает этот ответ, видит код 307 и заголовок Location, и автоматически отправляет новый GET-запрос уже на URL из Location
	5. Сервер по URL (из Location) отвечает, и браузер (клиент) показывает страницу
	*/

}
