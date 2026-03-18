# Titik Nol Backend

A GORM and Gin-based backend project for Titik Nol.

## Features

- Gin-gonic web framework
- GORM with PostgreSQL
- Dockerized setup
- Clean Architecture

## Getting Started

### Prerequisites

- Go 1.22+
- Docker & Docker Compose
- Make

### Installation

1. Clone the repository.
2. Copy `.env.example` to `.env`.
3. Run `make dev` or `docker-compose up -d`.

## Development Guidelines

### Git Commit Messages
This project follows the **Conventional Commits** specification. Please refer to the [Git Commit Rules](docs/git-commit-rules.md) for more details.

### API Standards
Consistency in API responses is crucial. Please follow the [API Response Standard](docs/api-response-standard.md) for all endpoints.

### Clean Architecture

The project follows Clean Architecture principles. Ensure that business logic is kept in the `internal/domain` (or similar) layer and that dependencies point inwards.
