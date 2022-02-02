package middleware

import (
	"context"
	"net/http"

	"github.com/GilbertoVGL/go-banking/pkg/config"
)

func ReqTimeout(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ctx, cancel := context.WithTimeout(ctx, config.RequestTimeout)
		defer cancel()
		r = r.Clone(ctx)

		next.ServeHTTP(w, r)
	})
}
