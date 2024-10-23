// backend/internal/services/monitoring/models.go

package monitoring

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// HealthCheck represents the health status of a service or component.
type HealthCheck struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ServiceName string             `bson:"service_name" json:"service_name"`
	Status      string             `bson:"status" json:"status"` // e.g., "healthy", "unhealthy"
	CheckedAt   time.Time          `bson:"checked_at" json:"checked_at"`
	Details     string             `bson:"details,omitempty" json:"details,omitempty"`
}
