// backend/internal/services/authorization/models.go

package authorization

import (
	"github.com/AkshayDubey29/MoniFlux/backend/internal/common"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

// Permission represents a permission entity.
type Permission struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name        string             `bson:"name" json:"name"`
	Description string             `bson:"description" json:"description"`
	CreatedAt   time.Time          `bson:"created_at" json:"createdAt"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updatedAt"`
}

// Role represents a role entity.
type Role struct {
	ID          primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	Name        string               `bson:"name" json:"name"`
	Permissions []primitive.ObjectID `bson:"permissions" json:"permissions"`
	CreatedAt   time.Time            `bson:"created_at" json:"createdAt"`
	UpdatedAt   time.Time            `bson:"updated_at" json:"updatedAt"`
}

// User represents a user entity.
// Assuming the User struct has a Roles field which is a slice of ObjectIDs.
type User = common.User
