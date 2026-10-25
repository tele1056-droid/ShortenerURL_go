package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/tele1056-droid/yaGo/internal/shortener"
	"github.com/tele1056-droid/yaGo/internal/storage"
)

type ShortenRequest struct {
	URL string `json:"url"`
}

type ShortenResponse struct {
	Result string `json:"result"`
}

type Handler struct {
	repo    storage.Repository
	baseURL string
}

func NewHandler(repo storage.Repository, baseURL string) *Handler {
	return &Handler{repo: repo, baseURL: baseURL}
}

// 1.a.
func (h *Handler) JSONShortenHandler(w http.ResponseWriter, r *http.Request) {
	// 1.a. Проверка метода (POST) и Content-Type (application/json)
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// 1.b Проверка Content-Type (с поддержкой charset)
	contentType := r.Header.Get("Content-Type")
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil || mediaType != "application/json" {
        http.Error(w, "Ожидается Content-Type: application/json", http.StatusUnsupportedMediaType)
        return
    }

	// 2.a. Чтение и парсинг JSON из тела
	var req ShortenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Некорректный JSON", http.StatusBadRequest)
		return
	}

	// 3.a. Валидация URL (можно переиспользовать логику из текстовго хендлера)
	if req.URL == "" {
		http.Error(w, "URL не может быть пустым", http.StatusBadRequest)
		return
	}

	// 4.a. Генерация кода и сохранение
	code := shortener.Shorten()
	h.repo.Save(code, req.URL)

	// 5.a. Формирование JSON-ответа
	resp := ShortenResponse{
		Result: fmt.Sprintf("%s/%s", h.baseURL, code),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp) // вот здесь используется JSON-энкодер

}

func (h *Handler) ShortenHandler(w http.ResponseWriter, r *http.Request) {
	// 1. проверка метода
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	// 2. Проверка Content-Type
	contentType := r.Header.Get("Content-Type")
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil || mediaType != "text/plain" {
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
	fmt.Fprintf(w, "%s/%s", h.baseURL, code)
}

func (h *Handler) RedirectHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Проверка метода
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// 2. Извлечение кода из пути, в данной версии теперь chi вырезает код из пути
	code := chi.URLParam(r, "id")

	if code == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

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
