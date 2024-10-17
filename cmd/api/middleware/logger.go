package middleware

import (
	"github.com/gorilla/mux"
	"github.com/grocky/go-api-starter/internal/log"
	"net/http"
)

func PopulateLogger(logger *log.Logger) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			if id := RequestIDFromContext(ctx); id != "" {
				logger = logger.With("requestId", id)
			}

			ctx = log.WithLogger(ctx, logger)
			r = r.Clone(ctx)

			next.ServeHTTP(w, r)
		})
	}
}
