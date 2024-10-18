## **MoniFlux API Documentation**

### **Table of Contents**
1. [Authentication](#authentication)
2. [Load Test Management](#load-test-management)
    - [Start a New Load Test](#start-a-new-load-test)
    - [Stop a Running Load Test](#stop-a-running-load-test)
    - [Restart a Load Test](#restart-a-load-test)
    - [Pause a Load Test](#pause-a-load-test)
    - [Resume a Load Test](#resume-a-load-test)
3. [Load Test Configuration](#load-test-configuration)
    - [Update Load Test Configuration](#update-load-test-configuration)
    - [Fetch Available Log/Metric/Trace Formats](#fetch-available-logmetrictrace-formats)
4. [Load Test Scheduling](#load-test-scheduling)
    - [Schedule a Load Test](#schedule-a-load-test)
    - [Cancel a Scheduled Load Test](#cancel-a-scheduled-load-test)
5. [Load Test Results](#load-test-results)
    - [Save Load Test Results](#save-load-test-results)
    - [Retrieve Past Load Test Results](#retrieve-past-load-test-results)
6. [Load Test Retrieval and Deletion](#load-test-retrieval-and-deletion)
    - [Retrieve All Load Tests](#retrieve-all-load-tests)
    - [Fetch Specific Test Configuration](#fetch-specific-test-configuration)
    - [Delete a Load Test](#delete-a-load-test)
7. [Error Handling](#error-handling)

---

### **Authentication**

Before accessing most of the MoniFlux APIs, users must authenticate to ensure secure access. MoniFlux uses **Bearer Token** authentication.

#### **Authentication Header**

- **Header Name**: `Authorization`
- **Header Value**: `Bearer <your-auth-token>`

*Note: Replace `<your-auth-token>` with your actual authentication token.*

---

### **Load Test Management**

#### 1. **Start a New Load Test**

- **Endpoint**: `/api/start-load`
- **Method**: `POST`
- **Description**: Initiates a new load test with specified configurations for logs, metrics, traces, and events.

##### **Request Payload**

```json
{
  "testID": "abc123",
  "userID": "user_123",
  "logTypes": [
    "catalina",
    "nginx",
    {
      "type": "custom",
      "format": {
        "timestamp": "{generated:timestamp}",
        "userID": "{generated:uuid}",
        "status": "{static:200}",
        "message": "{dynamic:random-text}",
        "ip": "{generated:ip-address}"
      }
    }
  ],
  "metrics": {
    "enabled": true,
    "rate": 5000,
    "duration": 1200
  },
  "traces": {
    "enabled": true,
    "rate": 3000,
    "duration": 1200
  },
  "events": {
    "enabled": true,
    "rate": 2000,
    "duration": 1200
  },
  "duration": 1200,
  "logRate": 5000
}
```

##### **Response Example**

```json
{
  "message": "Load test started successfully",
  "testID": "abc123",
  "userID": "user_123"
}
```

---

#### 2. **Stop a Running Load Test**

- **Endpoint**: `/api/stop-load`
- **Method**: `POST`
- **Description**: Stops an ongoing load test based on the provided test ID.

##### **Request Payload**

```json
{
  "testID": "abc123"
}
```

##### **Response Example**

```json
{
  "message": "Load test stopped successfully",
  "testID": "abc123"
}
```

---

#### 3. **Restart a Load Test**

- **Endpoint**: `/api/restart-load`
- **Method**: `POST`
- **Description**: Restarts a previously stopped or paused load test with the same or updated configurations.

##### **Request Payload**

```json
{
  "testID": "abc123",
  "logTypes": ["nginx"],
  "metrics": {
    "enabled": true,
    "rate": 8000,
    "duration": 1500
  },
  "traces": {
    "enabled": false
  },
  "events": {
    "enabled": true,
    "rate": 2500,
    "duration": 1500
  }
}
```

##### **Response Example**

```json
{
  "message": "Load test restarted successfully",
  "testID": "abc123"
}
```

---

#### 4. **Pause a Load Test**

- **Endpoint**: `/api/pause-load`
- **Method**: `POST`
- **Description**: Temporarily pauses an ongoing load test without stopping it completely.

##### **Request Payload**

```json
{
  "testID": "abc123"
}
```

##### **Response Example**

```json
{
  "message": "Load test paused successfully",
  "testID": "abc123"
}
```

---

#### 5. **Resume a Load Test**

- **Endpoint**: `/api/resume-load`
- **Method**: `POST`
- **Description**: Resumes a previously paused load test.

##### **Request Payload**

```json
{
  "testID": "abc123"
}
```

##### **Response Example**

```json
{
  "message": "Load test resumed successfully",
  "testID": "abc123"
}
```

---

### **Load Test Configuration**

#### 6. **Update Load Test Configuration**

- **Endpoint**: `/api/update-config`
- **Method**: `PUT`
- **Description**: Updates the configuration of an ongoing load test, allowing dynamic changes to logs, metrics, traces, and events.

##### **Request Payload**

##### **a. Updating Log Volume**

To **update the log volume** for a specific period:

```json
{
  "testID": "abc123",
  "logTypes": [
    "catalina",
    "nginx",
    {
      "type": "custom",
      "format": {
        "timestamp": "{generated:timestamp}",
        "userID": "{generated:uuid}",
        "status": "{static:200}",
        "message": "{dynamic:random-text}",
        "ip": "{generated:ip-address}"
      }
    }
  ],
  "logRate": 8000,
  "duration": 600
}
```

##### **b. Adding More Log Types with Custom Formats**

To **add more log types** (e.g., adding Nginx and custom JSON logs):

```json
{
  "testID": "abc123",
  "logTypes": [
    "catalina",
    "nginx",
    {
      "type": "custom",
      "format": {
        "timestamp": "{generated:timestamp}",
        "userID": "{generated:uuid}",
        "status": "{static:200}",
        "message": "{dynamic:random-text}",
        "ip": "{generated:ip-address}"
      }
    },
    {
      "type": "custom",
      "format": {
        "event": "{generated:event-type}",
        "value": "{dynamic:random-value}",
        "timestamp": "{generated:timestamp}"
      }
    }
  ],
  "metrics": {
    "enabled": true,
    "rate": 5000,
    "duration": 1200
  },
  "traces": {
    "enabled": true,
    "rate": 3000,
    "duration": 1200
  },
  "events": {
    "enabled": true,
    "rate": 2000,
    "duration": 1200
  },
  "logRate": 9000,
  "duration": 1800
}
```

##### **c. Disabling Specific Load Types**

To **disable log generation** and continue with metrics and events:

```json
{
  "testID": "abc123",
  "logTypes": [],
  "metrics": {
    "enabled": true,
    "rate": 5000,
    "duration": 1200
  },
  "traces": {
    "enabled": false
  },
  "events": {
    "enabled": true,
    "rate": 2000,
    "duration": 1200
  }
}
```

##### **Response Example**

```json
{
  "message": "Load test configuration updated successfully",
  "testID": "abc123"
}
```

---

#### 7. **Fetch Available Log/Metric/Trace Formats**

- **Endpoint**: `/api/get-log-formats`
- **Method**: `GET`
- **Description**: Retrieves a list of all supported log, metric, and trace formats available in MoniFlux.

##### **Request Payload**

*No payload required.*

##### **Response Example**

```json
{
  "availableLogFormats": ["catalina", "nginx", "custom"],
  "availableMetricFormats": ["prometheus", "custom"],
  "availableTraceFormats": ["opentelemetry", "custom"]
}
```

---

### **Load Test Scheduling**

#### 8. **Schedule a Load Test**

- **Endpoint**: `/api/schedule-load-test`
- **Method**: `POST`
- **Description**: Schedules a load test to start at a specified future time.

##### **Request Payload**

```json
{
  "testID": "scheduled_test_123",
  "userID": "user_123",
  "logTypes": ["catalina"],
  "metrics": {
    "enabled": true,
    "rate": 4000,
    "duration": 1200
  },
  "traces": {
    "enabled": false
  },
  "events": {
    "enabled": false
  },
  "startTime": "2024-10-20T10:00:00Z"
}
```

##### **Response Example**

```json
{
  "message": "Load test scheduled successfully",
  "testID": "scheduled_test_123",
  "startTime": "2024-10-20T10:00:00Z"
}
```

---

#### 9. **Cancel a Scheduled Load Test**

- **Endpoint**: `/api/cancel-scheduled-test`
- **Method**: `POST`
- **Description**: Cancels a previously scheduled load test that hasn't started yet.

##### **Request Payload**

```json
{
  "testID": "scheduled_test_123"
}
```

##### **Response Example**

```json
{
  "message": "Scheduled load test canceled successfully",
  "testID": "scheduled_test_123"
}
```

---

### **Load Test Results**

#### 10. **Save Load Test Results**

- **Endpoint**: `/api/save-results`
- **Method**: `POST`
- **Description**: Persists the results of a completed load test for future reference and analysis.

##### **Request Payload**

```json
{
  "testID": "abc123",
  "userID": "user_123",
  "results": {
    "logsGenerated": 120000,
    "metricsGenerated": 60000,
    "tracesGenerated": 30000,
    "eventsGenerated": 24000,
    "errors": 5,
    "latency": "120ms",
    "status": "Completed",
    "completedAt": "2024-10-18T12:00:00Z"
  }
}
```

##### **Response Example**

```json
{
  "message": "Load test results saved successfully",
  "testID": "abc123"
}
```

---

#### 11. **Retrieve Past Load Test Results**

- **Endpoint**: `/api/get-results`
- **Method**: `GET`
- **Description**: Fetches the results of past load tests, optionally filtered by test ID or user ID.

##### **Request Parameters**

- **Query Parameters**:
    - `testID` (optional): Specific test ID to retrieve results for.
    - `userID` (optional): Specific user ID to retrieve results for.

##### **Example Request**

```http
GET /api/get-results?testID=abc123&userID=user_123 HTTP/1.1
Host: <moniflux-api-endpoint>
Authorization: Bearer <your-auth-token>
```

##### **Response Example**

```json
{
  "testID": "abc123",
  "userID": "user_123",
  "results": {
    "logsGenerated": 120000,
    "metricsGenerated": 60000,
    "tracesGenerated": 30000,
    "eventsGenerated": 24000,
    "errors": 5,
    "latency": "120ms",
    "status": "Completed",
    "completedAt": "2024-10-18T12:00:00Z"
  }
}
```

*If multiple results are found, the response can be an array of results.*

---

### **Load Test Retrieval and Deletion**

#### 12. **Retrieve All Load Tests**

- **Endpoint**: `/api/get-all-tests`
- **Method**: `GET`
- **Description**: Retrieves a list of all load tests, including both running and completed tests.

##### **Request Parameters**

- **Query Parameters**:
    - `userID` (optional): Filter tests by a specific user ID.
    - `status` (optional): Filter tests by status (e.g., running, completed, paused).

##### **Example Request**

```http
GET /api/get-all-tests?userID=user_123&status=running HTTP/1.1
Host: <moniflux-api-endpoint>
Authorization: Bearer <your-auth-token>
```

##### **Response Example**

```json
{
  "tests": [
    {
      "testID": "abc123",
      "userID": "user_123",
      "status": "Running",
      "startTime": "2024-10-18T10:00:00Z",
      "duration": 1200
    },
    {
      "testID": "def456",
      "userID": "user_456",
      "status": "Completed",
      "startTime": "2024-10-17T09:00:00Z",
      "duration": 1800,
      "completedAt": "2024-10-17T09:30:00Z"
    }
  ]
}
```

---

#### 13. **Fetch Specific Test Configuration**

- **Endpoint**: `/api/get-config`
- **Method**: `GET`
- **Description**: Retrieves the configuration details of a specific load test based on the test ID.

##### **Request Parameters**

- **Query Parameters**:
    - `testID`: The unique identifier of the load test.

##### **Example Request**

```http
GET /api/get-config?testID=abc123 HTTP/1.1
Host: <moniflux-api-endpoint>
Authorization: Bearer <your-auth-token>
```

##### **Response Example**

```json
{
  "testID": "abc123",
  "userID": "user_123",
  "logTypes": [
    "catalina",
    "nginx",
    {
      "type": "custom",
      "format": {
        "timestamp": "{generated:timestamp}",
        "userID": "{generated:uuid}",
        "status": "{static:200}",
        "message": "{dynamic:random-text}",
        "ip": "{generated:ip-address}"
      }
    }
  ],
  "metrics": {
    "enabled": true,
    "rate": 5000,
    "duration": 1200
  },
  "traces": {
    "enabled": true,
    "rate": 3000,
    "duration": 1200
  },
  "events": {
    "enabled": true,
    "rate": 2000,
    "duration": 1200
  },
  "duration": 1200,
  "logRate": 5000
}
```

---

#### 14. **Delete a Load Test**

- **Endpoint**: `/api/delete-test`
- **Method**: `DELETE`
- **Description**: Deletes a specific load test's configuration and results, freeing up resources.

##### **Request Payload**

```json
{
  "testID": "abc123"
}
```

##### **Response Example**

```json
{
  "message": "Load test deleted successfully",
  "testID": "abc123"
}
```

---

### **Advanced Features**

#### 15. **Fetch Available Log/Metric/Trace Formats**

- **Endpoint**: `/api/get-log-formats`
- **Method**: `GET`
- **Description**: Retrieves a list of all supported log, metric, and trace formats available in MoniFlux.

##### **Request Payload**

*No payload required.*

##### **Response Example**

```json
{
  "availableLogFormats": ["catalina", "nginx", "custom"],
  "availableMetricFormats": ["prometheus", "custom"],
  "availableTraceFormats": ["opentelemetry", "custom"]
}
```

---

### **Error Handling**

MoniFlux APIs provide clear and consistent error messages to help users troubleshoot issues.

#### **Common Error Responses**

1. **Invalid Request Payload**

    - **HTTP Status**: `400 Bad Request`
    - **Response Example**:
      ```json
      {
        "error": "Invalid request payload"
      }
      ```

2. **Unauthorized Access**

    - **HTTP Status**: `401 Unauthorized`
    - **Response Example**:
      ```json
      {
        "error": "Unauthorized access"
      }
      ```

3. **Test Not Found**

    - **HTTP Status**: `404 Not Found`
    - **Response Example**:
      ```json
      {
        "error": "Load test not found",
        "testID": "abc123"
      }
      ```

4. **Internal Server Error**

    - **HTTP Status**: `500 Internal Server Error`
    - **Response Example**:
      ```json
      {
        "error": "Internal server error"
      }
      ```

*Note: Each API endpoint should implement appropriate error handling to cover various failure scenarios.*

---

### **Comprehensive List of MoniFlux APIs**

Below is a summary list of all MoniFlux APIs, their endpoints, HTTP methods, and brief descriptions:

| **API Name**                           | **Endpoint**                      | **Method** | **Description**                                                         |
|----------------------------------------|-----------------------------------|------------|-------------------------------------------------------------------------|
| **Start a New Load Test**              | `/api/start-load`                 | POST       | Initiates a new load test with specified configurations.               |
| **Stop a Running Load Test**           | `/api/stop-load`                  | POST       | Stops an ongoing load test based on the test ID.                        |
| **Restart a Load Test**                | `/api/restart-load`               | POST       | Restarts a previously stopped or paused load test.                     |
| **Pause a Load Test**                  | `/api/pause-load`                 | POST       | Temporarily pauses an ongoing load test.                                |
| **Resume a Load Test**                 | `/api/resume-load`                | POST       | Resumes a previously paused load test.                                  |
| **Update Load Test Configuration**     | `/api/update-config`              | PUT        | Updates the configuration of an ongoing load test dynamically.         |
| **Fetch Available Log/Metric/Trace Formats** | `/api/get-log-formats`          | GET        | Retrieves supported log, metric, and trace formats.                    |
| **Schedule a Load Test**               | `/api/schedule-load-test`         | POST       | Schedules a load test to start at a specified future time.             |
| **Cancel a Scheduled Load Test**       | `/api/cancel-scheduled-test`      | POST       | Cancels a previously scheduled load test that hasn't started yet.       |
| **Save Load Test Results**             | `/api/save-results`               | POST       | Persists the results of a completed load test.                          |
| **Retrieve Past Load Test Results**    | `/api/get-results`                | GET        | Fetches results of past load tests, optionally filtered by parameters.  |
| **Retrieve All Load Tests**            | `/api/get-all-tests`              | GET        | Retrieves a list of all load tests, both running and completed.          |
| **Fetch Specific Test Configuration**  | `/api/get-config`                 | GET        | Retrieves configuration details of a specific load test.                |
| **Delete a Load Test**                 | `/api/delete-test`                | DELETE     | Deletes a specific load test's configuration and results.               |

---

### **Detailed API Call Examples**

#### **1. Start a New Load Test**

##### **Request**

```http
POST /api/start-load HTTP/1.1
Host: <moniflux-api-endpoint>
Authorization: Bearer <your-auth-token>
Content-Type: application/json

{
  "testID": "abc123",
  "userID": "user_123",
  "logTypes": [
    "catalina",
    "nginx",
    {
      "type": "custom",
      "format": {
        "timestamp": "{generated:timestamp}",
        "userID": "{generated:uuid}",
        "status": "{static:200}",
        "message": "{dynamic:random-text}",
        "ip": "{generated:ip-address}"
      }
    }
  ],
  "metrics": {
    "enabled": true,
    "rate": 5000,
    "duration": 1200
  },
  "traces": {
    "enabled": true,
    "rate": 3000,
    "duration": 1200
  },
  "events": {
    "enabled": true,
    "rate": 2000,
    "duration": 1200
  },
  "duration": 1200,
  "logRate": 5000
}
```

##### **Response**

```json
{
  "message": "Load test started successfully",
  "testID": "abc123",
  "userID": "user_123"
}
```

---

#### **2. Update Load Test Configuration**

##### **a. Update Log Volume**

**Scenario**: Changing log rate from 5000 to 8000 logs/sec for 600 seconds.

##### **Request**

```http
PUT /api/update-config HTTP/1.1
Host: <moniflux-api-endpoint>
Authorization: Bearer <your-auth-token>
Content-Type: application/json

{
  "testID": "abc123",
  "logTypes": [
    "catalina",
    "nginx",
    {
      "type": "custom",
      "format": {
        "timestamp": "{generated:timestamp}",
        "userID": "{generated:uuid}",
        "status": "{static:200}",
        "message": "{dynamic:random-text}",
        "ip": "{generated:ip-address}"
      }
    }
  ],
  "logRate": 8000,
  "duration": 600
}
```

##### **Response**

```json
{
  "message": "Load test configuration updated successfully",
  "testID": "abc123"
}
```

---

##### **b. Add More Log Types with Custom Formats**

**Scenario**: Adding Nginx logs and a new custom JSON log format.

##### **Request**

```http
PUT /api/update-config HTTP/1.1
Host: <moniflux-api-endpoint>
Authorization: Bearer <your-auth-token>
Content-Type: application/json

{
  "testID": "abc123",
  "logTypes": [
    "catalina",
    "nginx",
    {
      "type": "custom",
      "format": {
        "timestamp": "{generated:timestamp}",
        "userID": "{generated:uuid}",
        "status": "{static:200}",
        "message": "{dynamic:random-text}",
        "ip": "{generated:ip-address}"
      }
    },
    {
      "type": "custom",
      "format": {
        "event": "{generated:event-type}",
        "value": "{dynamic:random-value}",
        "timestamp": "{generated:timestamp}"
      }
    }
  ],
  "metrics": {
    "enabled": true,
    "rate": 5000,
    "duration": 1200
  },
  "traces": {
    "enabled": true,
    "rate": 3000,
    "duration": 1200
  },
  "events": {
    "enabled": true,
    "rate": 2000,
    "duration": 1200
  },
  "logRate": 9000,
  "duration": 1800
}
```

##### **Response**

```json
{
  "message": "Load test configuration updated successfully",
  "testID": "abc123"
}
```

---

#### **3. Add Metrics to Events**

##### **Scenario**: Starting metrics generation alongside existing event generation.

##### **Request**

```http
PUT /api/update-config HTTP/1.1
Host: <moniflux-api-endpoint>
Authorization: Bearer <your-auth-token>
Content-Type: application/json

{
  "testID": "def456",
  "metrics": {
    "enabled": true,
    "rate": 3000,
    "duration": 1800
  },
  "events": {
    "enabled": true,
    "rate": 2000,
    "duration": 1800
  }
}
```

##### **Response**

```json
{
  "message": "Load configuration updated successfully",
  "testID": "def456"
}
```

---

### **Handling Multiple Users and Test Segregation**

MoniFlux ensures that logs, metrics, traces, and events generated by different users and tests are segregated properly to prevent data conflicts and ensure data integrity.

#### **Log Storage Structure**

Logs are stored in host-mounted directories, organized by user ID and test ID:

```
/var/logs/moniflux/
├── user_123/
│   ├── test_abc123/
│   │   ├── catalina_logs.log
│   │   ├── nginx_logs.log
│   │   └── custom_logs.log
├── user_456/
│   ├── test_def456/
│   │   ├── event_logs.log
│   │   └── metrics.log
```

#### **Kubernetes Volume Mount Example**

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: moniflux-load-generator
spec:
  containers:
  - name: moniflux
    image: moniflux:latest
    volumeMounts:
    - name: log-volume
      mountPath: /mnt/logs
  volumes:
  - name: log-volume
    hostPath:
      path: /var/logs/moniflux    # Host path to store logs
      type: DirectoryOrCreate
```

### **Load Generation with Segregated Logs**

#### **Updated `load_generator.go` with Segregated Log Management**

```go
package loadgenerator

import (
    "fmt"
    "log"
    "os"
    "path/filepath"
    "time"
    "math/rand"
)

// LogPattern represents the structure for a custom log type
type LogPattern struct {
    Type   string            `json:"type"`
    Format map[string]string `json:"format"` // Key-value pairs of log structure
}

// MetricsConfig represents the metrics load generation configuration
type MetricsConfig struct {
    Enabled  bool `json:"enabled"`
    Rate     int  `json:"rate"`
    Duration int  `json:"duration"`
}

// TracesConfig represents the traces load generation configuration
type TracesConfig struct {
    Enabled  bool `json:"enabled"`
    Rate     int  `json:"rate"`
    Duration int  `json:"duration"`
}

// EventsConfig represents the events load generation configuration
type EventsConfig struct {
    Enabled  bool `json:"enabled"`
    Rate     int  `json:"rate"`
    Duration int  `json:"duration"`
}

// LoadTestConfig holds the configuration for the load test
type LoadTestConfig struct {
    TestID    string         `json:"testID"`
    UserID    string         `json:"userID"`
    LogTypes  []LogPattern   `json:"logTypes"`
    Metrics   MetricsConfig  `json:"metrics"`
    Traces    TracesConfig   `json:"traces"`
    Events    EventsConfig   `json:"events"`
    LogRate   int            `json:"logRate"`
    Duration  int            `json:"duration"`
}

// UpdateLoadTypesAndConfig updates the log types, metrics, traces, and events dynamically
func UpdateLoadTypesAndConfig(config LoadTestConfig) {
    if len(config.LogTypes) == 0 {
        fmt.Printf("Stopping log generation for test %s\n", config.TestID)
    } else {
        fmt.Printf("Generating logs for test %s with types: %v\n", config.TestID, config.LogTypes)
    }

    if config.Metrics.Enabled {
        fmt.Printf("Generating metrics for test %s at rate: %d/sec for %d seconds\n", 
            config.TestID, config.Metrics.Rate, config.Metrics.Duration)
    } else {
        fmt.Printf("Metrics generation disabled for test %s\n", config.TestID)
    }

    if config.Traces.Enabled {
        fmt.Printf("Generating traces for test %s at rate: %d/sec for %d seconds\n", 
            config.TestID, config.Traces.Rate, config.Traces.Duration)
    } else {
        fmt.Printf("Traces generation disabled for test %s\n", config.TestID)
    }

    if config.Events.Enabled {
        fmt.Printf("Generating events for test %s at rate: %d/sec for %d seconds\n", 
            config.TestID, config.Events.Rate, config.Events.Duration)
    } else {
        fmt.Printf("Events generation disabled for test %s\n", config.TestID)
    }

    // Implement logic to update the running load generators based on new configurations
}

// GenerateLogs generates logs for the specified test and user
func GenerateLogs(config LoadTestConfig) {
    // Define the base log directory for the test
    baseDir := filepath.Join("/mnt/logs", fmt.Sprintf("user_%s", config.UserID), fmt.Sprintf("test_%s", config.TestID))

    // Ensure the directory exists
    if err := os.MkdirAll(baseDir, os.ModePerm); err != nil {
        log.Fatalf("Error creating log directory: %v", err)
    }

    for _, logType := range config.LogTypes {
        logFile := filepath.Join(baseDir, fmt.Sprintf("%s_logs.log", logType.Type))

        // Open the log file for writing
        file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
        if err != nil {
            log.Fatalf("Error opening log file: %v", err)
        }

        // Start generating logs for the specified duration
        ticker := time.NewTicker(time.Second / time.Duration(config.LogRate))
        done := make(chan bool)

        go func() {
            time.Sleep(time.Duration(config.Duration) * time.Second)
            done <- true
        }()

        for {
            select {
            case <-done:
                fmt.Printf("Log generation complete for test %s, log type %s\n", config.TestID, logType.Type)
                ticker.Stop()
                file.Close()
                return
            case <-ticker.C:
                // Generate log entry based on the format
                logEntry := generateLogEntry(logType.Format)
                if _, err := file.WriteString(logEntry); err != nil {
                    log.Printf("Error writing to log file: %v", err)
                }
            }
        }
    }
}

// generateLogEntry generates a log entry based on the provided format
func generateLogEntry(format map[string]string) string {
    logEntry := ""
    for key, pattern := range format {
        value := generateDynamicValue(pattern)
        logEntry += fmt.Sprintf("%s=%s ", key, value)
    }
    logEntry += "\n"
    return logEntry
}

// Generate dynamic values based on user-defined patterns
func generateDynamicValue(pattern string) string {
    switch {
    case pattern == "{generated:timestamp}":
        return time.Now().Format(time.RFC3339)
    case pattern == "{generated:uuid}":
        return generateUUID()
    case pattern == "{generated:ip-address}":
        return generateRandomIP()
    case pattern == "{dynamic:random-text}":
        return generateRandomText()
    case pattern == "{generated:event-type}":
        return generateRandomEventType()
    case pattern == "{dynamic:random-value}":
        return generateRandomValue()
    default:
        // If pattern starts with "{static:", return the static value
        if len(pattern) > 9 && pattern[:9] == "{static:" && pattern[len(pattern)-1] == '}' {
            return pattern[9 : len(pattern)-1]
        }
        return pattern // Return as-is if no pattern is matched
    }
}

// Generate a random UUID (example implementation)
func generateUUID() string {
    return fmt.Sprintf("%08x-%04x-%04x-%04x-%12x",
        rand.Uint32(), rand.Uint16(), rand.Uint16(), rand.Uint16(), rand.Uint32())
}

// Generate a random IP address (example implementation)
func generateRandomIP() string {
    return fmt.Sprintf("%d.%d.%d.%d", rand.Intn(256), rand.Intn(256), rand.Intn(256), rand.Intn(256))
}

// Generate random text (example implementation)
func generateRandomText() string {
    words := []string{"foo", "bar", "baz", "qux", "quux"}
    return words[rand.Intn(len(words))]
}

// Generate random event type (example implementation)
func generateRandomEventType() string {
    events := []string{"login", "logout", "purchase", "click", "view"}
    return events[rand.Intn(len(events))]
}

// Generate random value for custom fields (example implementation)
func generateRandomValue() string {
    values := []string{"value1", "value2", "value3", "value4", "value5"}
    return values[rand.Intn(len(values))]
}
```

---

### **Conclusion**

This comprehensive API documentation ensures that **MoniFlux** covers all necessary scenarios for managing and performing observability load tests. By providing detailed endpoints, request payloads, and response examples, both users and developers can effectively interact with the MoniFlux system, leveraging its full capabilities for performance testing and benchmarking.

Feel free to expand upon this documentation as **MoniFlux** evolves and new features are introduced. Contributions and feedback are always welcome to enhance the tool's functionality and usability.
