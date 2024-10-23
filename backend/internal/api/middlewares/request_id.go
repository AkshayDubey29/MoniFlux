// backend/internal/api/middlewares/request_id.go

package middlewares

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

// ContextKey is a type for context keys used in the middleware
type ContextKey string

const (
	ContextRequestIDKey ContextKey = "requestID"
)

// RequestIDMiddleware generates a unique request ID for each HTTP request
func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Generate a new UUID
		requestID := uuid.New().String()

		// Add the request ID to the response headers
		w.Header().Set("X-Request-ID", requestID)

		// Add the request ID to the request context
		ctx := r.Context()
		ctx = contextWithRequestID(ctx, requestID)

		// Pass the request with the new context to the next handler
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// contextWithRequestID adds the request ID to the context
func contextWithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, ContextRequestIDKey, requestID)
}

// GetRequestID retrieves the request ID from the context
func GetRequestID(ctx context.Context) (string, bool) {
	requestID, ok := ctx.Value(ContextRequestIDKey).(string)
	return requestID, ok
}
