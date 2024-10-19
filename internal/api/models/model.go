package models

import "time"

// Test represents a load test configuration and status
type Test struct {
    TestID        string    `json:"testID" bson:"testID"`
    UserID        string    `json:"userID" bson:"userID"`
    LogType       string    `json:"logType" bson:"logType"`
    LogRate       int       `json:"logRate" bson:"logRate"`
    LogSize       int       `json:"logSize" bson:"logSize"`
    MetricsRate   int       `json:"metricsRate" bson:"metricsRate"`
    TraceRate     int       `json:"traceRate" bson:"traceRate"`
    Duration      int       `json:"duration" bson:"duration"`
    Destination   Destination `json:"destination" bson:"destination"`
    Status        string    `json:"status" bson:"status"`
    ScheduledTime time.Time `json:"scheduledTime,omitempty" bson:"scheduledTime,omitempty"`
    CreatedAt     time.Time `json:"createdAt" bson:"createdAt"`
    UpdatedAt     time.Time `json:"updatedAt" bson:"updatedAt"`
}

// Destination represents where the payloads are delivered
type Destination struct {
    Port     int    `json:"port" bson:"port"`
    Endpoint string `json:"endpoint" bson:"endpoint"`
}
