// backend/internal/services/authentication/middleware.go

package authentication

import (
	"context"
	"net/http"
	"strings"

	"github.com/AkshayDubey29/MoniFlux/backend/internal/api/models"
	"github.com/sirupsen/logrus"
)

// AuthMiddleware handles JWT authentication for protected routes.
type AuthMiddleware struct {
	authService *AuthenticationService
	logger      *logrus.Logger
}

// NewAuthMiddleware creates a new instance of AuthMiddleware.
func NewAuthMiddleware(authService *AuthenticationService, logger *logrus.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		authService: authService,
		logger:      logger,
	}
}

// MiddlewareFunc is the HTTP middleware function that enforces authentication.
func (am *AuthMiddleware) MiddlewareFunc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Retrieve the Authorization header.
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
			return
		}

		// Expected format: "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			http.Error(w, "Invalid Authorization header format", http.StatusUnauthorized)
			return
		}

		tokenString := parts[1]

		// Validate the JWT token using the AuthenticationService.
		claims, err := am.authService.ValidateJWT(tokenString)
		if err != nil {
			am.logger.Errorf("Invalid JWT token: %v", err)
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Retrieve the user associated with the token.
		user, err := am.authService.GetUserByID(r.Context(), claims.UserID)
		if err != nil {
			am.logger.Errorf("Failed to retrieve user from token: %v", err)
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Inject the user into the request context.
		ctx := context.WithValue(r.Context(), "user", user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserFromContext retrieves the authenticated user from the request context.
// Returns the User and a boolean indicating whether the user was found.
func GetUserFromContext(ctx context.Context) (*models.User, bool) {
	user, ok := ctx.Value("user").(*models.User)
	return user, ok
}
