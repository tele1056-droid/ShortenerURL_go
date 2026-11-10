package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

const CookieName = "user_id"

// sign создает HMAC-подпись для userID
func sign(userID, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(userID))
	return hex.EncodeToString(h.Sum(nil))
}

// CreateCookie создает и устанавливает подписанную куку
func CreateCookie(w http.ResponseWriter, secret string) string {
	userID := uuid.New().String()
	signature := sign(userID, secret)
	cookieValue := userID + "." + signature

	http.SetCookie(w, &http.Cookie{
		Name:  CookieName,
		Value: cookieValue,
		Path:  "/",
	})
	return userID
}

// GetUserID проверяет подпись и возвращает userID. Если кука невалидна — возвращает пустую строку.
func GetUserID(r *http.Request, secret string) (string, bool) {
	cookie, err := r.Cookie(CookieName)
	if err != nil {
		return "", false
	}

	parts := strings.Split(cookie.Value, ".")
	if len(parts) != 2 {
		return "", false // неверный формат
	}

	userID, signature := parts[0], parts[1]
	expectedSignature := sign(userID, secret)

	if !hmac.Equal([]byte(signature), []byte(expectedSignature)) {
		return "", false // подпись не совпадает
	}

	return userID, true
}
