package handlers

import (
    "net/http"
    "net/http/httptest"
    "strings"
    "testing"

    "github.com/tele1056-droid/yaGo/internal/storage"
)

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
            h := NewHandler(repo)

            // Создаём фейковый запрос
            req := httptest.NewRequest(tt.method, "/", strings.NewReader(tt.body))
            req.Header.Set("Content-Type", tt.contentType)

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
    repo.Save("abc123", "https://example.com")
    h := NewHandler(repo)

    tests := []struct {
        name        string
        method      string
        path        string
        wantCode    int
        wantLocation string
    }{
        {"успех", http.MethodGet, "/abc123", http.StatusTemporaryRedirect, "https://example.com"},
        {"неверный метод", http.MethodPost, "/abc123", http.StatusMethodNotAllowed, ""},
        {"пустой код", http.MethodGet, "/", http.StatusBadRequest, ""},
        {"несуществующий код", http.MethodGet, "/nonexistent", http.StatusNotFound, ""},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            req := httptest.NewRequest(tt.method, tt.path, nil)
            w := httptest.NewRecorder()

            h.RedirectHandler(w, req)

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