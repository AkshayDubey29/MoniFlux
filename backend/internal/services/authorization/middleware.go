// internal/services/authorization/middleware.go

package authorization

import (
	"net/http"

	"github.com/AkshayDubey29/MoniFlux/backend/internal/services/authentication"
	"github.com/sirupsen/logrus"
)

// AuthorizationMiddleware handles permission checks for protected routes.
type AuthorizationMiddleware struct {
	authService         *AuthorizationService
	logger              *logrus.Logger
	requiredPermissions []string
}

// NewAuthorizationMiddleware creates a new instance of AuthorizationMiddleware.
// Parameters:
// - authService: Instance of AuthorizationService for checking permissions.
// - logger: Logger for logging authorization events and errors.
// - requiredPermissions: Slice of permission names required to access the route.
func NewAuthorizationMiddleware(authService *AuthorizationService, logger *logrus.Logger, requiredPermissions []string) *AuthorizationMiddleware {
	return &AuthorizationMiddleware{
		authService:         authService,
		logger:              logger,
		requiredPermissions: requiredPermissions,
	}
}

// MiddlewareFunc is the HTTP middleware function that enforces permission checks.
func (am *AuthorizationMiddleware) MiddlewareFunc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Retrieve the authenticated user from the context.
		commonUser, ok := authentication.GetUserFromContext(r.Context())
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Check if the user has all required permissions.
		for _, perm := range am.requiredPermissions {
			hasPerm, err := am.authService.UserHasPermission(r.Context(), commonUser.ID.Hex(), perm)
			if err != nil {
				am.logger.Errorf("Error checking permission %s for user %s: %v", perm, commonUser.ID.Hex(), err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
			if !hasPerm {
				http.Error(w, "Forbidden: insufficient permissions", http.StatusForbidden)
				return
			}
		}

		// User has all required permissions; proceed to the next handler.
		next.ServeHTTP(w, r)
	})
}
