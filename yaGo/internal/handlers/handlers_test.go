package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/tele1056-droid/yaGo/internal/storage"
)

func TestGetUserURLsHandler(t *testing.T) {
	repo := storage.NewMemoryStorage()
	repo.Save("1", "https://example.com", "test-user")
	h := NewHandler(repo, "http://localhost:8080")

	// Без куки → 204
	req := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)
	w := httptest.NewRecorder()
	h.GetUserURLsHandler(w, req)
	if w.Code != http.StatusNoContent {
		t.Errorf("ожидался код 204, получен %d", w.Code)
	}

	// С кукой → 200
	req = httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)
	req.AddCookie(&http.Cookie{Name: "user_id", Value: "test-user"})
	w = httptest.NewRecorder()
	h.GetUserURLsHandler(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("ожидался код 200, получен %d", w.Code)
	}
}

func TestJSONShortenHandler(t *testing.T) {
	tests := []struct {
		name        string
		method      string
		contentType string
		body        string
		wantCode    int
		wantResult  string // ожидаемый результат (если 201)
	}{
		{
			"успех",
			http.MethodPost,
			"application/json",
			`{"url":"https://practicum.yandex.ru"}`,
			http.StatusCreated,
			"http://localhost:8080/1", // счётчик начинается с 1
		},
		{
			"неверный метод",
			http.MethodGet,
			"application/json",
			`{"url":"https://practicum.yandex.ru"}`,
			http.StatusMethodNotAllowed,
			"",
		},
		{
			"неверный content-type",
			http.MethodPost,
			"text/plain",
			`{"url":"https://practicum.yandex.ru"}`,
			http.StatusUnsupportedMediaType,
			"",
		},
		{
			"пустой URL",
			http.MethodPost,
			"application/json",
			`{"url":""}`,
			http.StatusBadRequest,
			"",
		},
		{
			"некорректный JSON",
			http.MethodPost,
			"application/json",
			`{"url":`,
			http.StatusBadRequest,
			"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := storage.NewMemoryStorage()
			h := NewHandler(repo, "http://localhost:8080")

			req := httptest.NewRequest(tt.method, "/api/shorten", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", tt.contentType)
			req.AddCookie(&http.Cookie{Name: "user_id", Value: "test-user"})

			w := httptest.NewRecorder()
			h.JSONShortenHandler(w, req)

			if w.Code != tt.wantCode {
				t.Errorf("ожидался код %d, получен %d", tt.wantCode, w.Code)
			}

			// Проверяем только префикс, а не точный код
			if tt.wantCode == http.StatusCreated {
				var resp ShortenResponse
				if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
					t.Fatalf("не удалось распарсить ответ: %v", err)
				}
				if !strings.HasPrefix(resp.Result, "http://localhost:8080/") {
					t.Errorf("неожиданный результат: %s", resp.Result)
				}
			}
		})
	}
}

func TestShortenHandler(t *testing.T) {
	tests := []struct {
		name        string
		method      string
		contentType string
		body        string
		wantCode    int
	}{
		{"успех", http.MethodPost, "text/plain", "https://example.com", http.StatusCreated},
		{"неверный метод", http.MethodGet, "text/plain", "https://example.com", http.StatusMethodNotAllowed},
		{"неверный content-type", http.MethodPost, "application/json", "{}", http.StatusUnsupportedMediaType},
		{"пустое тело", http.MethodPost, "text/plain", "", http.StatusBadRequest},
		{"невалидный URL", http.MethodPost, "text/plain", "абракадабра", http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаём мок-хранилище
			repo := storage.NewMemoryStorage()
			h := NewHandler(repo, "http://localhost:8080")

			// Создаём фейковый запрос
			req := httptest.NewRequest(tt.method, "/", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", tt.contentType)
			req.AddCookie(&http.Cookie{Name: "user_id", Value: "test-user"})
			// Создаём фейковый ответ
			w := httptest.NewRecorder()

			// Вызываем хендлер
			h.ShortenHandler(w, req)

			// Проверяем код ответа
			if w.Code != tt.wantCode {
				t.Errorf("ожидался код %d, получен %d", tt.wantCode, w.Code)
			}
		})
	}
}

func TestRedirectHandler(t *testing.T) {
	// Подготовка: создаём хранилище с одной ссылкой
	repo := storage.NewMemoryStorage()
	repo.Save("abc123", "https://example.com", "test-user")
	h := NewHandler(repo, "http://localhost:8080")

	tests := []struct {
		name         string
		method       string
		path         string
		wantCode     int
		wantLocation string
	}{
		{"успех", http.MethodGet, "/abc123", http.StatusTemporaryRedirect, "https://example.com"},
		{"неверный метод", http.MethodPost, "/abc123", http.StatusMethodNotAllowed, ""},
		{"несуществующий код", http.MethodGet, "/nonexistent", http.StatusNotFound, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаём роутер chi и регистрируем маршрут
			r := chi.NewRouter()
			r.Get("/{id}", h.RedirectHandler)

			// Создаём запрос и прогоняем его через роутер
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != tt.wantCode {
				t.Errorf("ожидался код %d, получен %d", tt.wantCode, w.Code)
			}

			if tt.wantLocation != "" {
				location := w.Header().Get("Location")
				if location != tt.wantLocation {
					t.Errorf("ожидался Location %s, получен %s", tt.wantLocation, location)
				}
			}
		})
	}
}
