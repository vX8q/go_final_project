package api

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type AuthRequest struct {
	Password string `json:"password"`
}

type AuthResponse struct {
	Token string `json:"token,omitempty"`
	Error string `json:"error,omitempty"`
}

func SigninHandler(w http.ResponseWriter, r *http.Request) {
	var req AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	expectedPass := os.Getenv("TODO_PASSWORD")
	if expectedPass == "" {
		writeJSON(w, AuthResponse{Error: "Authentication not configured"})
		return
	}

	if req.Password != expectedPass {
		writeError(w, "Неверный пароль", http.StatusUnauthorized)
		return
	}

	// Генерация JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"exp": time.Now().Add(8 * time.Hour).Unix(),
	})

	tokenString, err := token.SignedString([]byte(expectedPass))
	if err != nil {
		writeError(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:    "token",
		Value:   tokenString,
		Expires: time.Now().Add(8 * time.Hour),
	})

	writeJSON(w, AuthResponse{Token: tokenString})
}
