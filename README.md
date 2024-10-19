# MoniFlux

## Overview
MoniFlux is a scalable load generation and observability tool designed to help users perform load tests by generating logs, metrics, and traces simultaneously or independently.

## Table of Contents
1. [Introduction](#introduction)
2. [Key Features](#key-features)
3. [Technologies](#technologies)
4. [Getting Started](#getting-started)
5. [Directory Structure](#directory-structure)
6. [Contributing](#contributing)
7. [License](#license)

## Introduction
MoniFlux enables users to dynamically configure and manage test parameters, deliver payloads to custom destinations, and run multiple tests in parallel while ensuring data isolation.

## Key Features
- Dynamic Load Generation
- Multiple Log Formats
- Real-Time Configuration Updates
- Flexible Payload Delivery
- User Isolation
- Kubernetes Deployment with Helm Charts
- CI/CD Integration
- MongoDB for Data Storage

## Technologies
- **Backend**: Golang
- **Database**: MongoDB
- **Orchestration**: Kubernetes
- **CI/CD**: Jenkins/CircleCI
- **Deployment**: Helm
- **Frontend**: React
- **Observability**: Grafana Loki

## Getting Started
Instructions to set up the project locally.

## Directory Structure
```bash
MoniFlux/
├── cmd/
│   ├── api/
│   │   └── main.go
│   └── loadgen/
│       └── main.go
├── internal/
│   ├── api/
│   │   ├── handlers/
│   │   │   └── handler.go
│   │   ├── models/
│   │   │   └── model.go
│   │   ├── routers/
│   │   │   └── router.go
│   │   └── middlewares/
│   │       └── middleware.go
│   ├── loadgen/
│   │   ├── generators/
│   │   │   └── generator.go
│   │   ├── delivery/
│   │   │   └── delivery.go
│   │   └── controllers/
│   │       └── controller.go
│   ├── config/
│   │   ├── v1/
│   │   │   └── config.go
│   │   └── utils/
│   │       └── config_utils.go
│   ├── db/
│   │   ├── mongo/
│   │   │   └── mongo.go
│   │   └── redis/
│   │       └── redis.go
│   └── services/
│       ├── authentication/
│       │   └── auth.go
│       ├── authorization/
│       │   └── authorization.go
│       └── monitoring/
│           └── monitoring.go
├── pkg/
│   ├── config/
│   │   └── config.go
│   ├── db/
│   │   └── db.go
│   ├── logger/
│   │   └── logger.go
│   └── utils/
│       └── utils.go
├── configs/
│   └── config.yaml
├── scripts/
│   └── ci_cd/
│       └── pipeline.sh
├── deployments/
│   ├── kubernetes/
│   │   ├── manifests/
│   │   │   └── deployment.yaml
│   │   └── templates/
│   │       └── service.yaml
│   └── helm/
│       └── MoniFlux/
│           ├── charts/
│           │   └── chart.yaml
│           ├── templates/
│           │   └── deployment.yaml
│           └── values/
│               └── values.yaml
├── tests/
│   ├── unit/
│   │   └── example_test.go
│   └── integration/
│       └── example_integration_test.go
├── docs/
│   └── api_documentation.md
├── logs/
│   └── .gitkeep
└── README.md
```

## Contributing
Guidelines for contributing to MoniFlux.

## License
[MIT](LICENSE)
