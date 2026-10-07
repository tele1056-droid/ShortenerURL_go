package handlers_test

import (
    "net/http"
    "net/http/httptest"
    "strings"
    "testing"
	"io"
	"github.com/tele1056-droid/yaGo/internal/handlers"
)

func TestShortenHandler(t *testing.T) {
    // 1. Создаём фейковый запрос
    body := strings.NewReader("https://example.com")
    req := httptest.NewRequest(http.MethodPost, "/", body)
    req.Header.Set("Content-Type", "text/plain")

    // 2. Создаём фейковый ResponseWriter (записывает ответ в память)
    w := httptest.NewRecorder()

    // 3. Вызываем хендлер
    handlers.ShortenHandler(w, req)

    // 4. Проверяем результат
    resp := w.Result()
    defer resp.Body.Close()

    // Проверяем код ответа
    if resp.StatusCode != http.StatusCreated {
        t.Errorf("ожидался код 201, получен %d", resp.StatusCode)
    }

    // Проверяем Content-Type
    if resp.Header.Get("Content-Type") != "text/plain" {
        t.Errorf("ожидался Content-Type text/plain, получен %s", resp.Header.Get("Content-Type"))
    }

    // Проверяем тело ответа
    respBody, _ := io.ReadAll(resp.Body)
    if !strings.HasPrefix(string(respBody), "http://localhost:8080/") {
        t.Errorf("неожиданное тело ответа: %s", string(respBody))
    }
}
