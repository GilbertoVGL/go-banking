package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt"
)

const BEARER_SCHEMA = "Bearer "

type UserIdContextKey string

func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		splitToken := strings.Split(r.Header.Get("Authorization"), BEARER_SCHEMA)

		if len(splitToken) != 2 {
			respondWithError(w, http.StatusBadRequest, "invalid authentication token")
			return
		}

		token, err := jwt.Parse(splitToken[1], func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				respondWithError(w, http.StatusBadRequest, "invalid signing method")
			}
			return []byte(os.Getenv("JWT_SECRET")), nil
		})

		if err != nil {
			respondWithError(w, http.StatusUnauthorized, "invalid authentication token")
			return
		}

		if !token.Valid {
			respondWithError(w, http.StatusUnauthorized, "invalid authentication token")
		}

		claims := token.Claims.(jwt.MapClaims)
		ctx := r.Context()
		ctx = context.WithValue(ctx, UserIdContextKey("userClaims"), claims["userId"])
		ro := r.Clone(ctx)

		next.ServeHTTP(w, ro)
	})
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
