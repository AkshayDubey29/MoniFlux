// backend/internal/services/authentication/service.go

package authentication

import (
	"context"
	"errors"
	"time"

	"github.com/AkshayDubey29/MoniFlux/backend/internal/api/models"
	"github.com/AkshayDubey29/MoniFlux/backend/internal/common"
	jwt "github.com/golang-jwt/jwt/v4"
	"github.com/sirupsen/logrus" // Added logrus import
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive" // Added primitive import for ObjectID handling
	"go.mongodb.org/mongo-driver/mongo"
)

// AuthenticationService provides methods for JWT operations and user retrieval.
type AuthenticationService struct {
	config         *common.Config
	logger         *logrus.Logger
	userCollection *mongo.Collection
	jwtSecret      string
}

// NewAuthenticationService creates a new instance of AuthenticationService.
func NewAuthenticationService(cfg *common.Config, logger *logrus.Logger, mongoClient *mongo.Client) (*AuthenticationService, error) {
	userCol := mongoClient.Database(cfg.MongoDB).Collection("users")
	if userCol == nil {
		return nil, errors.New("failed to get users collection")
	}

	return &AuthenticationService{
		config:         cfg,
		logger:         logger,
		userCollection: userCol,
		jwtSecret:      cfg.JWTSecret,
	}, nil
}

// ValidateJWT validates the JWT token and returns the claims.
func (as *AuthenticationService) ValidateJWT(tokenString string) (*models.Claims, error) {
	claims := &models.Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// Ensure the signing method is HMAC
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
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

// GetUserByID retrieves a user by their ID.
func (as *AuthenticationService) GetUserByID(ctx context.Context, userID string) (*models.User, error) {
	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		as.logger.Errorf("Invalid userID format: %v", err)
		return nil, errors.New("invalid user ID format")
	}

	var user models.User
	err = as.userCollection.FindOne(ctx, bson.M{"_id": objID}).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("user not found")
		}
		as.logger.Errorf("Error retrieving user: %v", err)
		return nil, errors.New("internal server error")
	}

	return &user, nil
}

// GenerateJWT generates a JWT token for a given user.
func (as *AuthenticationService) GenerateJWT(userID string) (string, error) {
	claims := &models.Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24)), // Token valid for 24 hours
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "MoniFlux",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(as.jwtSecret))
}
