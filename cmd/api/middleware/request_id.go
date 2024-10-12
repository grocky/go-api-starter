package middleware

import (
	"context"
	"go-api-starter/internal/log"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

const contextKeyRequestID = contextKey("request_id")

func PopulateRequestID() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			logger := log.FromContext(ctx)

			if existing := RequestIDFromContext(ctx); existing == "" {
				u, err := uuid.NewRandom()
				if err != nil {
					logger.Error(err.Error())
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				ctx = withRequestID(ctx, u.String())
				r = r.Clone(ctx)
			}

			next.ServeHTTP(w, r)
		})
	}
}

func RequestIDFromContext(ctx context.Context) string {
	v := ctx.Value(contextKeyRequestID)
	if v == nil {
		return ""
	}

	t, ok := v.(string)
	if !ok {
		return ""
	}

	return t
}

func withRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, contextKeyRequestID, id)
}
