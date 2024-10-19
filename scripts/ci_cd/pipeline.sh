#!/bin/bash

# Example CI/CD pipeline script using Jenkins or CircleCI

# Steps:
# 1. Build the application
# 2. Run tests
# 3. Build Docker image
# 4. Push Docker image to registry
# 5. Deploy to Kubernetes using Helm

echo "Starting CI/CD pipeline..."

# Build
echo "Building the application..."
go build -o bin/api cmd/api/main.go
go build -o bin/loadgen cmd/loadgen/main.go

# Test
echo "Running tests..."
go test ./...

# Docker build (Assuming Dockerfile is present)
echo "Building Docker images..."
docker build -t akshaydubey29/moniflux-api:latest -f Dockerfile.api .
docker build -t akshaydubey29/moniflux-loadgen:latest -f Dockerfile.loadgen .

# Push Docker images
echo "Pushing Docker images to registry..."
docker push akshaydubey29/moniflux-api:latest
docker push akshaydubey29/moniflux-loadgen:latest

# Deploy using Helm
echo "Deploying to Kubernetes using Helm..."
helm upgrade --install moniflux ./deployments/helm/MoniFlux

echo "CI/CD pipeline completed successfully."
