package api

import (
	"net/http"
	"os"

	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if pass := os.Getenv("TODO_PASSWORD"); pass != "" {
			cookie, err := r.Cookie("token")
			if err != nil {
				writeError(w, "Authentification required", http.StatusUnauthorized)
				return
			}

			token, err := jwt.Parse(cookie.Value, func(token *jwt.Token) (interface{}, error) {
				return []byte(pass), nil
			})

			if err != nil || !token.Valid {
				writeError(w, "Invalid token", http.StatusUnauthorized)
				return
			}
		}
		next(w, r)
	})
}
