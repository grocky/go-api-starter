package middleware

import (
	"github.com/gorilla/mux"
	"github.com/grocky/go-api-starter/internal/log"
	"net/http"
)

func Recovery() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			logger := log.FromContext(ctx).Named("middleware.Recovery")

			defer func() {
				if err := recover(); err != nil {
					logger.Error("http handler panic", "panic", err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
			}()

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
