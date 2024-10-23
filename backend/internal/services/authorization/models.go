// backend/internal/services/authorization/models.go

package authorization

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Permission represents a specific action or access right within the system.
type Permission struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name        string             `bson:"name" json:"name"` // e.g., "create_user", "delete_logs"
	Description string             `bson:"description" json:"description"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
}

// Role represents a collection of permissions.
type Role struct {
	ID          primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	Name        string               `bson:"name" json:"name"`               // e.g., "admin", "editor", "viewer"
	Permissions []primitive.ObjectID `bson:"permissions" json:"permissions"` // References to Permission IDs
	CreatedAt   time.Time            `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time            `bson:"updated_at" json:"updated_at"`
}
