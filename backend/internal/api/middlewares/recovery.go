// backend/internal/api/middlewares/recovery.go

package middlewares

import (
	"net/http"

	"github.com/sirupsen/logrus"
)

// RecoveryMiddleware recovers from panics, logs the error, and returns a 500 response.
func RecoveryMiddleware(logger *logrus.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					logger.Errorf("Panic recovered: %v", err)
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
