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
	"github.com/google/uuid" // или используйте crypto/rand
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

type respPair struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

func NewHandler(repo storage.Repository, baseURL string) *Handler {
	return &Handler{repo: repo, baseURL: baseURL}
}

// getUserID извлекает user_id из куки. Если куки нет, возвращает false.
func getUserID(r *http.Request) (string, bool) {
	cookie, err := r.Cookie("user_id")
	if err != nil {
		return "", false
	}
	return cookie.Value, true
}

// setUserID устанавливает куку с user_id.
func setUserID(w http.ResponseWriter, userID string) {
	http.SetCookie(w, &http.Cookie{
		Name:  "user_id",
		Value: userID,
		Path:  "/",
	})
}

func (h *Handler) GetUserURLsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	userID, ok := getUserID(r)
	if !ok {
		w.WriteHeader(http.StatusNoContent) // 204
		return
	}

	urls, err := h.repo.GetAllByUser(userID)
	if err != nil {
		http.Error(w, "Ошибка получения данных", http.StatusInternalServerError)
		return
	}
	if len(urls) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	var result []respPair
	for code, original := range urls {
		result = append(result, respPair{
			ShortURL:    fmt.Sprintf("%s/%s", h.baseURL, code),
			OriginalURL: original,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}

func (h *Handler) JSONShortenHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	contentType := r.Header.Get("Content-Type")
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil || mediaType != "application/json" {
		http.Error(w, "Ожидается Content-Type: application/json", http.StatusUnsupportedMediaType)
		return
	}

	var req ShortenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Некорректный JSON", http.StatusBadRequest)
		return
	}

	if req.URL == "" {
		http.Error(w, "URL не может быть пустым", http.StatusBadRequest)
		return
	}

	// Получаем или создаем user_id
	userID, ok := getUserID(r)
	if !ok {
		userID = uuid.New().String()
		setUserID(w, userID)
	}

	code := shortener.Shorten()
	h.repo.Save(code, req.URL, userID) // передаем userID

	resp := ShortenResponse{
		Result: fmt.Sprintf("%s/%s", h.baseURL, code),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) ShortenHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	contentType := r.Header.Get("Content-Type")
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil || mediaType != "text/plain" {
		http.Error(w, "Ожидается Content-Type: text/plain", http.StatusUnsupportedMediaType)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1024*1024)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Ошибка чтения тела", http.StatusBadRequest)
		return
	}

	if len(body) == 0 {
		http.Error(w, "Пустое тело запроса", http.StatusBadRequest)
		return
	}

	longURL := strings.TrimSpace(string(body))
	if longURL == "" {
		http.Error(w, "URL не может быть пустым", http.StatusBadRequest)
		return
	}

	if !strings.HasPrefix(longURL, "http://") && !strings.HasPrefix(longURL, "https://") {
		http.Error(w, "URL должен начинаться с http:// или https://", http.StatusBadRequest)
		return
	}

	parsedURL, err := url.ParseRequestURI(longURL)
	if err != nil || parsedURL.Host == "" {
		http.Error(w, "Некорректный URL", http.StatusBadRequest)
		return
	}

	// Получаем или создаем user_id
	userID, ok := getUserID(r)
	if !ok {
		userID = uuid.New().String()
		setUserID(w, userID)
	}

	code := shortener.Shorten()
	h.repo.Save(code, longURL, userID) // передаем userID

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
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
