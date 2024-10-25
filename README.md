# MoniFlux Backend Service

![MoniFlux Logo](https://example.com/logo.png)

MoniFlux is a comprehensive load testing and monitoring platform designed to help developers and DevOps teams assess the performance and reliability of their applications. The **Backend Service** is a core component of MoniFlux, responsible for managing load tests, scheduling, result aggregation, and interfacing with the frontend and other services.

## Table of Contents

- [Features](#features)
- [Architecture](#architecture)
- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Configuration](#configuration)
- [Running the Service](#running-the-service)
- [API Documentation](#api-documentation)
- [Testing](#testing)
- [Deployment](#deployment)
- [Logging](#logging)
- [Contributing](#contributing)
- [License](#license)

## Features

- **Load Test Management**: Create, start, schedule, cancel, and restart load tests.
- **Result Aggregation**: Collect and store logs, metrics, and traces generated during tests.
- **Scheduling**: Schedule tests to run at specific times.
- **Real-Time Monitoring**: Monitor the status and progress of ongoing tests.
- **API Access**: Comprehensive RESTful APIs for integration with frontend and other services.
- **Scalability**: Designed to handle multiple concurrent load tests efficiently.
- **Logging & Error Handling**: Robust logging with Logrus and proper error management.

## Architecture

MoniFlux Backend Service is built using Go and MongoDB. It leverages goroutines for concurrent load test execution and ensures thread-safe operations using mutexes. The service exposes RESTful APIs for interacting with load tests and managing their lifecycle.

### High-Level Flow Diagram

```plaintext
+-----------------+       +-------------------+       +------------------+
|                 |       |                   |       |                  |
|  Frontend/UI    +------->  Backend Service   +------->  MongoDB         |
|                 |       |                   |       |                  |
+-----------------+       +-------------------+       +------------------+
                                 |
                                 |
                                 v
                        +-----------------+
                        |                 |
                        |  Load Generation |
                        |     Goroutines    |
                        |                 |
                        +-----------------+
```

*Note: For a more detailed and visual diagram, refer to the `/docs/architecture.png` file.*

## Prerequisites

Before setting up the MoniFlux Backend Service, ensure you have the following installed on your system:

- **Go**: Version 1.18 or higher
- **MongoDB**: Version 4.0 or higher
- **Git**: For cloning the repository
- **Docker** (optional): For containerized deployments
- **Make** (optional): For using provided Makefile commands

## Installation

### 1. Clone the Repository

```bash
git clone https://github.com/AkshayDubey29/MoniFlux.git
cd MoniFlux/backend
```

### 2. Install Dependencies

Ensure Go modules are enabled and download all dependencies:

```bash
go mod download
```

### 3. Set Up MongoDB

If you don't have MongoDB installed locally, you can use Docker to run a MongoDB instance:

```bash
docker run --name moniflux-mongo -d -p 27017:27017 mongo:latest
```

> **Note**: The default MongoDB URI is `mongodb://localhost:27017`. Modify it if your setup differs.

### 4. Configure the Service

Create a `.env` file in the root of the backend directory to specify configuration variables:

```bash
touch .env
```

Populate the `.env` file with the following variables:

```env
# MongoDB Configuration
MONGO_URI=mongodb://localhost:27017
MONGO_DB=moniflux

# Server Configuration
PORT=8080

# Logging
LOG_LEVEL=info

# Other configurations can be added as needed
```

Alternatively, you can use a configuration file (e.g., `config.yaml`) and modify the `common.Config` struct accordingly.

## Configuration

The Backend Service uses environment variables for configuration. Ensure that all required variables are set either in the `.env` file or your environment.

### Environment Variables

- `MONGO_URI`: URI for connecting to MongoDB (e.g., `mongodb://localhost:27017`)
- `MONGO_DB`: Name of the MongoDB database (e.g., `moniflux`)
- `PORT`: Port on which the backend service will run (e.g., `8080`)
- `LOG_LEVEL`: Logging level (`debug`, `info`, `warn`, `error`)

### Configuration Struct

The `common.Config` struct holds all configuration parameters:

```go
type Config struct {
    MongoURI string
    MongoDB  string
    Port     string
    LogLevel string
    // Add other configuration fields as necessary
}
```

Ensure that these fields are populated correctly when initializing the `LoadGenController`.

## Running the Service

### 1. Build the Application

```bash
go build -o moniflux-backend cmd/api/main.go
```

### 2. Run the Application

```bash
./moniflux-backend
```

The backend service should now be running on `http://localhost:8080` (or the port specified in your `.env` file).

### 3. Using Docker (Optional)

You can also run the Backend Service using Docker for an isolated environment.

#### Build the Docker Image

```bash
docker build -t moniflux-backend .
```

#### Run the Docker Container

```bash
docker run -d --name moniflux-backend -p 8080:8080 --env-file .env moniflux-backend
```

## API Documentation

The Backend Service exposes several RESTful APIs to manage load tests. Below is a summary of the available endpoints.

### 1. Start a New Test

**Endpoint**: `POST /start-test`

**Description**: Initiates a new load test or updates an existing one.

**Request Body**:

```json
{
  "testID": "unique-test-id-123",
  "userID": "user123",
  "logType": "INFO",
  "logRate": 500,
  "logSize": 50,
  "metricsRate": 1000,
  "traceRate": 200,
  "duration": 600,
  "destination": {
    "name": "Destination1",
    "endpoint": "http://example.com/endpoint",
    "port": 8080,
    "apiKey": "test-api-key"
  }
}
```

**Response**:

- `200 OK`: Test started successfully.
- `400 Bad Request`: Invalid input or test already running.

### 2. Restart an Existing Test

**Endpoint**: `POST /restart-test`

**Description**: Restarts an existing load test with updated configurations.

**Request Body**:

```json
{
  "testID": "unique-test-id-123",
  "logRate": 1667,
  "duration": 600
}
```

**Response**:

- `200 OK`: Test restarted successfully.
- `400 Bad Request`: Invalid input or test cannot be restarted.

### 3. Schedule a Test

**Endpoint**: `POST /schedule-test`

**Description**: Schedules a load test to start at a specified future time.

**Request Body**:

```json
{
  "testID": "unique-test-id-123",
  "userID": "user123",
  "schedule": "2024-10-25T15:00:00Z"
}
```

**Response**:

- `200 OK`: Test scheduled successfully.
- `400 Bad Request`: Invalid input or test cannot be scheduled.

### 4. Cancel a Test

**Endpoint**: `POST /cancel-test`

**Description**: Cancels a running or scheduled load test.

**Request Body**:

```json
{
  "testID": "unique-test-id-123"
}
```

**Response**:

- `200 OK`: Test cancelled successfully.
- `400 Bad Request`: Invalid input or test cannot be cancelled.

### 5. Create a New Test

**Endpoint**: `POST /create-test`

**Description**: Creates a new load test with a pending status.

**Request Body**:

```json
{
  "testID": "new-test-id-12345",
  "userID": "user123",
  "logType": "INFO",
  "logRate": 300,
  "logSize": 30,
  "metricsRate": 600,
  "traceRate": 100,
  "duration": 300,
  "destination": {
    "name": "Destination2",
    "endpoint": "http://example.com/endpoint",
    "port": 9090,
    "apiKey": "another-test-key"
  }
}
```

**Response**:

- `201 Created`: Test created successfully.
- `400 Bad Request`: Test with the same ID already exists.

### 6. Retrieve All Tests

**Endpoint**: `GET /tests`

**Description**: Retrieves all active and scheduled tests.

**Response**:

- `200 OK`: List of tests.
- `500 Internal Server Error`: Failed to retrieve tests.

### 7. Retrieve Test by ID

**Endpoint**: `GET /tests/{testID}`

**Description**: Retrieves details of a specific test by its ID.

**Response**:

- `200 OK`: Test details.
- `404 Not Found`: Test not found.

## Testing

MoniFlux Backend Service includes unit and integration tests to ensure reliability and correctness.

### 1. Running Unit Tests

Navigate to the backend directory and execute:

```bash
go test ./...
```

This command runs all tests in the project recursively.

### 2. Running Integration Tests

Integration tests require a running MongoDB instance. Ensure MongoDB is running before executing:

```bash
go test -tags=integration ./...
```

> **Note**: Integration tests are tagged separately to distinguish them from unit tests.

### 3. Using Makefile (Optional)

If a `Makefile` is provided, you can use predefined commands:

```bash
make test
make integration-test
```

## Deployment

MoniFlux Backend Service can be deployed using various methods, including Docker and Kubernetes, for scalable and reliable operations.

### 1. Docker Deployment

#### Build the Docker Image

```bash
docker build -t moniflux-backend .
```

#### Run the Docker Container

```bash
docker run -d \
  --name moniflux-backend \
  -p 8080:8080 \
  --env-file .env \
  moniflux-backend
```

### 2. Kubernetes Deployment

Create a Kubernetes deployment YAML (`deployment.yaml`):

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: moniflux-backend
spec:
  replicas: 3
  selector:
    matchLabels:
      app: moniflux-backend
  template:
    metadata:
      labels:
        app: moniflux-backend
    spec:
      containers:
        - name: backend
          image: moniflux-backend:latest
          ports:
            - containerPort: 8080
          envFrom:
            - configMapRef:
                name: moniflux-backend-config
---
apiVersion: v1
kind: Service
metadata:
  name: moniflux-backend-service
spec:
  type: LoadBalancer
  ports:
    - port: 80
      targetPort: 8080
  selector:
    app: moniflux-backend
```

Apply the deployment:

```bash
kubectl apply -f deployment.yaml
```

### 3. Continuous Integration/Continuous Deployment (CI/CD)

Integrate with CI/CD tools like GitHub Actions, Jenkins, or GitLab CI to automate building, testing, and deploying the Backend Service upon code commits and merges.

## Logging

MoniFlux Backend Service uses **Logrus** for structured logging. Logs include information about test lifecycle events, errors, and debug information.

### Log Levels

- `debug`: Detailed information, typically of interest only when diagnosing problems.
- `info`: Confirmation that things are working as expected.
- `warn`: An indication that something unexpected happened, or indicative of some problem in the near future.
- `error`: Due to a more serious problem, the software has not been able to perform some function.

### Configuring Log Levels

Set the `LOG_LEVEL` environment variable in the `.env` file:

```env
LOG_LEVEL=info
```

## Contributing

Contributions are welcome! Follow these steps to contribute to MoniFlux Backend Service:

1. **Fork the Repository**

   Click the "Fork" button at the top-right of the repository page.

2. **Clone Your Fork**

   ```bash
   git clone https://github.com/your-username/MoniFlux.git
   cd MoniFlux/backend
   ```

3. **Create a New Branch**

   ```bash
   git checkout -b feature/your-feature-name
   ```

4. **Make Changes and Commit**

   ```bash
   git add .
   git commit -m "Add your descriptive commit message"
   ```

5. **Push to Your Fork**

   ```bash
   git push origin feature/your-feature-name
   ```

6. **Create a Pull Request**

   Navigate to the original repository and click "Compare & pull request."

### Code Standards

- Follow Go's [Effective Go](https://golang.org/doc/effective_go) guidelines.
- Ensure code is formatted using `gofmt`.
- Write clear and concise commit messages.
- Include unit and integration tests for new features.

## License

This project is licensed under the [MIT License](LICENSE).

---

## Contact

For questions, issues, or feature requests, please open an issue on the [GitHub repository](https://github.com/AkshayDubey29/MoniFlux/issues) or contact the maintainer at [email@example.com](mailto:email@example.com).

---

## Acknowledgements

- [Logrus](https://github.com/sirupsen/logrus) for structured logging.
- [MongoDB](https://www.mongodb.com/) for the NoSQL database solution.
- [UUID](https://github.com/google/uuid) for unique identifier generation.

---

## Additional Resources

- [MoniFlux Frontend Documentation](https://github.com/AkshayDubey29/MoniFlux/frontend)
- [API Specifications](./docs/api-specs.md)
- [Architecture Diagrams](./docs/architecture.png)

---

**Happy Testing with MoniFlux!**