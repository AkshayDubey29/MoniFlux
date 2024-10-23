// backend/internal/services/authentication/service.go

package authentication

import (
	"context"
	"errors"

	"github.com/AkshayDubey29/MoniFlux/backend/internal/api/models"
	"github.com/AkshayDubey29/MoniFlux/backend/internal/common"
	jwt "github.com/golang-jwt/jwt/v4"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// AuthenticationService handles user authentication and JWT operations.
type AuthenticationService struct {
	logger      *logrus.Logger
	mongoClient *mongo.Client
	jwtSecret   string
	config      *common.Config
}

// NewAuthenticationService creates a new AuthenticationService instance.
func NewAuthenticationService(logger *logrus.Logger, mongoClient *mongo.Client, jwtSecret string, cfg *common.Config) *AuthenticationService {
	return &AuthenticationService{
		logger:      logger,
		mongoClient: mongoClient,
		jwtSecret:   jwtSecret,
		config:      cfg,
	}
}

// ValidateJWT validates the JWT token and returns the claims.
func (as *AuthenticationService) ValidateJWT(tokenString string) (*models.Claims, error) {
	claims := &models.Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(as.jwtSecret), nil
	})
	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

// GetUserByID retrieves a user by their ID from the database.
func (as *AuthenticationService) GetUserByID(ctx context.Context, userID string) (*models.User, error) {
	var user models.User
	collection := as.mongoClient.Database("your_database_name").Collection("users")
	err := collection.FindOne(ctx, bson.M{"userID": userID}).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, mongo.ErrNoDocuments
		}
		as.logger.Errorf("Failed to retrieve user: %v", err)
		return nil, err
	}
	return &user, nil
}
