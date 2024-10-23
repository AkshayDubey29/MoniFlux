// backend/internal/services/authorization/service.go

package authorization

import (
	"context"
	"errors"
	"time"

	"github.com/AkshayDubey29/MoniFlux/backend/internal/common"           // Single import without alias
	mongoDB "github.com/AkshayDubey29/MoniFlux/backend/internal/db/mongo" // Aliased to avoid conflict
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	mongoDriver "go.mongodb.org/mongo-driver/mongo" // Aliased for official driver
)

// AuthorizationService provides methods for managing roles and permissions.
type AuthorizationService struct {
	config               *common.Config
	logger               *logrus.Logger
	mongoClient          *mongoDB.MongoClient
	roleCollection       *mongoDriver.Collection
	permissionCollection *mongoDriver.Collection
	userCollection       *mongoDriver.Collection
}

// NewAuthorizationService creates a new instance of AuthorizationService.
func NewAuthorizationService(cfg *common.Config, logger *logrus.Logger, mongoClient *mongoDB.MongoClient) *AuthorizationService {
	roleCol := mongoClient.Client.Database(cfg.MongoDB).Collection("roles")
	permissionCol := mongoClient.Client.Database(cfg.MongoDB).Collection("permissions")
	userCol := mongoClient.Client.Database(cfg.MongoDB).Collection("users") // Initialize user collection
	return &AuthorizationService{
		config:               cfg,
		logger:               logger,
		mongoClient:          mongoClient,
		roleCollection:       roleCol,
		permissionCollection: permissionCol,
		userCollection:       userCol,
	}
}

// CreatePermission creates a new permission.
func (as *AuthorizationService) CreatePermission(ctx context.Context, name, description string) (*Permission, error) {
	// Check if permission with the same name already exists.
	var existing Permission
	err := as.permissionCollection.FindOne(ctx, bson.M{"name": name}).Decode(&existing)
	if err == nil {
		return nil, errors.New("permission already exists")
	}
	if err != mongoDriver.ErrNoDocuments {
		as.logger.Errorf("Error checking existing permission: %v", err)
		return nil, errors.New("internal server error")
	}

	// Create new permission.
	permission := &Permission{
		Name:        name,
		Description: description,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	result, err := as.permissionCollection.InsertOne(ctx, permission)
	if err != nil {
		as.logger.Errorf("Error inserting permission: %v", err)
		return nil, errors.New("internal server error")
	}

	// Type assertion with error handling
	insertedID, ok := result.InsertedID.(primitive.ObjectID)
	if !ok {
		as.logger.Errorf("Failed to assert InsertedID to primitive.ObjectID")
		return nil, errors.New("internal server error")
	}
	permission.ID = insertedID
	as.logger.Infof("Permission created: %s", name)
	return permission, nil
}

// GetPermission retrieves a permission by its name.
func (as *AuthorizationService) GetPermission(ctx context.Context, name string) (*Permission, error) {
	var permission Permission
	err := as.permissionCollection.FindOne(ctx, bson.M{"name": name}).Decode(&permission)
	if err != nil {
		if errors.Is(err, mongoDriver.ErrNoDocuments) {
			return nil, errors.New("permission not found")
		}
		as.logger.Errorf("Error retrieving permission: %v", err)
		return nil, errors.New("internal server error")
	}
	return &permission, nil
}

// CreateRole creates a new role with the specified permissions.
func (as *AuthorizationService) CreateRole(ctx context.Context, name string, permissionNames []string) (*Role, error) {
	// Check if role with the same name already exists.
	var existing Role
	err := as.roleCollection.FindOne(ctx, bson.M{"name": name}).Decode(&existing)
	if err == nil {
		return nil, errors.New("role already exists")
	}
	if err != mongoDriver.ErrNoDocuments {
		as.logger.Errorf("Error checking existing role: %v", err)
		return nil, errors.New("internal server error")
	}

	// Fetch permission IDs.
	permissionIDs := []primitive.ObjectID{}
	for _, pname := range permissionNames {
		perm, err := as.GetPermission(ctx, pname)
		if err != nil {
			return nil, err // Permission not found or internal error.
		}
		permissionIDs = append(permissionIDs, perm.ID)
	}

	// Create new role.
	role := &Role{
		Name:        name,
		Permissions: permissionIDs,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	result, err := as.roleCollection.InsertOne(ctx, role)
	if err != nil {
		as.logger.Errorf("Error inserting role: %v", err)
		return nil, errors.New("internal server error")
	}

	// Type assertion with error handling
	insertedID, ok := result.InsertedID.(primitive.ObjectID)
	if !ok {
		as.logger.Errorf("Failed to assert InsertedID to primitive.ObjectID for role %s", name)
		return nil, errors.New("internal server error")
	}
	role.ID = insertedID
	as.logger.Infof("Role created: %s", name)
	return role, nil
}

// GetRole retrieves a role by its name.
func (as *AuthorizationService) GetRole(ctx context.Context, name string) (*Role, error) {
	var role Role
	err := as.roleCollection.FindOne(ctx, bson.M{"name": name}).Decode(&role)
	if err != nil {
		if errors.Is(err, mongoDriver.ErrNoDocuments) {
			return nil, errors.New("role not found")
		}
		as.logger.Errorf("Error retrieving role: %v", err)
		return nil, errors.New("internal server error")
	}
	return &role, nil
}

// AssignRoleToUser assigns a role to a user.
// Assumes that the User model has a 'Roles' field which is a slice of ObjectIDs referencing roles.
func (as *AuthorizationService) AssignRoleToUser(ctx context.Context, userID string, roleName string) error {
	// Convert userID to ObjectID
	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		as.logger.Errorf("Invalid userID format: %v", err)
		return errors.New("invalid user ID format")
	}

	// Fetch the role by name.
	role, err := as.GetRole(ctx, roleName)
	if err != nil {
		return err
	}

	// Update the user's roles.
	filter := bson.M{"_id": userObjectID}
	update := bson.M{
		"$addToSet": bson.M{
			"roles": role.ID,
		},
		"$set": bson.M{
			"updated_at": time.Now(),
		},
	}

	result, err := as.userCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		as.logger.Errorf("Error assigning role to user: %v", err)
		return errors.New("internal server error")
	}

	if result.MatchedCount == 0 {
		return errors.New("user not found")
	}

	as.logger.Infof("Role %s assigned to user %s", roleName, userID)
	return nil
}

// RemoveRoleFromUser removes a role from a user.
func (as *AuthorizationService) RemoveRoleFromUser(ctx context.Context, userID string, roleName string) error {
	// Convert userID to ObjectID
	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		as.logger.Errorf("Invalid userID format: %v", err)
		return errors.New("invalid user ID format")
	}

	// Fetch the role by name.
	role, err := as.GetRole(ctx, roleName)
	if err != nil {
		return err
	}

	// Update the user's roles.
	filter := bson.M{"_id": userObjectID}
	update := bson.M{
		"$pull": bson.M{
			"roles": role.ID,
		},
		"$set": bson.M{
			"updated_at": time.Now(),
		},
	}

	result, err := as.userCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		as.logger.Errorf("Error removing role from user: %v", err)
		return errors.New("internal server error")
	}

	if result.MatchedCount == 0 {
		return errors.New("user not found")
	}

	as.logger.Infof("Role %s removed from user %s", roleName, userID)
	return nil
}

// UserHasPermission checks if a user has a specific permission.
func (as *AuthorizationService) UserHasPermission(ctx context.Context, userID string, permissionName string) (bool, error) {
	// Convert userID to ObjectID
	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		as.logger.Errorf("Invalid userID format: %v", err)
		return false, errors.New("invalid user ID format")
	}

	// Fetch the permission by name.
	permission, err := as.GetPermission(ctx, permissionName)
	if err != nil {
		return false, err
	}

	// Fetch the user and populate roles.
	var user User
	err = as.userCollection.FindOne(ctx, bson.M{"_id": userObjectID}).Decode(&user)
	if err != nil {
		if errors.Is(err, mongoDriver.ErrNoDocuments) {
			return false, errors.New("user not found")
		}
		as.logger.Errorf("Error retrieving user: %v", err)
		return false, errors.New("internal server error")
	}

	// If user has no roles, deny access.
	if len(user.Roles) == 0 {
		return false, nil
	}

	// Fetch roles and check for the permission.
	cursor, err := as.roleCollection.Find(ctx, bson.M{"_id": bson.M{"$in": user.Roles}})
	if err != nil {
		as.logger.Errorf("Error fetching user roles: %v", err)
		return false, errors.New("internal server error")
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var role Role
		if err := cursor.Decode(&role); err != nil {
			as.logger.Errorf("Error decoding role: %v", err)
			continue
		}
		for _, pid := range role.Permissions {
			if pid == permission.ID {
				return true, nil
			}
		}
	}

	if err := cursor.Err(); err != nil {
		as.logger.Errorf("Cursor error: %v", err)
		return false, errors.New("internal server error")
	}

	return false, nil
}

// CreateDefaultRoles initializes default roles and permissions if they do not exist.
func (as *AuthorizationService) CreateDefaultRoles(ctx context.Context) error {
	// Define default permissions.
	defaultPermissions := []Permission{
		{
			Name:        "create_user",
			Description: "Ability to create new users",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			Name:        "delete_user",
			Description: "Ability to delete existing users",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			Name:        "view_logs",
			Description: "Ability to view system logs",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		// Add more default permissions as needed.
	}

	for _, perm := range defaultPermissions {
		// Check if permission exists.
		var existing Permission
		err := as.permissionCollection.FindOne(ctx, bson.M{"name": perm.Name}).Decode(&existing)
		if err == mongoDriver.ErrNoDocuments {
			// Insert the permission.
			result, err := as.permissionCollection.InsertOne(ctx, perm)
			if err != nil {
				as.logger.Errorf("Error inserting default permission %s: %v", perm.Name, err)
				return err
			}
			insertedID, ok := result.InsertedID.(primitive.ObjectID)
			if !ok {
				as.logger.Errorf("Failed to assert InsertedID to primitive.ObjectID for permission %s", perm.Name)
				return errors.New("internal server error")
			}
			perm.ID = insertedID
			as.logger.Infof("Default permission created: %s", perm.Name)
		} else if err != nil {
			as.logger.Errorf("Error checking default permission %s: %v", perm.Name, err)
			return err
		} else {
			as.logger.Infof("Default permission already exists: %s", perm.Name)
		}
	}

	// Define default roles.
	defaultRoles := []struct {
		Name        string
		Permissions []string
	}{
		{
			Name: "admin",
			Permissions: []string{
				"create_user",
				"delete_user",
				"view_logs",
				// Add more permissions as needed.
			},
		},
		{
			Name: "editor",
			Permissions: []string{
				"create_user",
				"view_logs",
				// Add more permissions as needed.
			},
		},
		{
			Name: "viewer",
			Permissions: []string{
				"view_logs",
				// Add more permissions as needed.
			},
		},
	}

	for _, roleDef := range defaultRoles {
		// Check if role exists.
		var existing Role
		err := as.roleCollection.FindOne(ctx, bson.M{"name": roleDef.Name}).Decode(&existing)
		if err == mongoDriver.ErrNoDocuments {
			// Create the role.
			role, err := as.CreateRole(ctx, roleDef.Name, roleDef.Permissions)
			if err != nil {
				as.logger.Errorf("Error creating default role %s: %v", roleDef.Name, err)
				return err
			}
			as.logger.Infof("Default role created: %s", role.Name)
		} else if err != nil {
			as.logger.Errorf("Error checking default role %s: %v", roleDef.Name, err)
			return err
		} else {
			as.logger.Infof("Default role already exists: %s", roleDef.Name)
		}
	}

	return nil
}
