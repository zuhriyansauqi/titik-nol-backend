# Titik Nol Backend

Personal finance API built with Go, following Clean Architecture. Manages accounts, transactions, categories, and provides a dashboard summary — all behind Google SSO authentication.

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Language | Go 1.26+ |
| Framework | Gin |
| ORM | GORM + PostgreSQL |
| Auth | Google SSO + JWT |
| Config | Viper |
| Logging | `log/slog` (structured, context-aware) |
| Testing | `testing` + testify |
| Infra | Docker + Docker Compose |

## Architecture

```
cmd/api/              → Entrypoint
internal/
  domain/             → Entities, interfaces, domain errors
  usecase/            → Business logic
  repository/         → Data access (GORM)
  delivery/http/      → Handlers + middleware (Gin)
  infrastructure/     → Config, database, logger
  pkg/                → Shared packages (JWT, Google SSO, response helpers)
migrations/           → SQL migrations (golang-migrate)
```

Dependencies flow inward: `delivery → usecase → domain ← repository`.

### Component Diagram

```mermaid
graph TD
    subgraph "Delivery Layer"
        AH[AccountHandler]
        TH[TransactionHandler]
        DH[DashboardHandler]
        CH[CategoryHandler]
        OH[OnboardingHandler]
    end

    subgraph "Usecase Layer"
        AU[AccountUsecase]
        TU[TransactionUsecase]
        DU[DashboardUsecase]
        CU[CategoryUsecase]
        OU[OnboardingUsecase]
        RU[ReconciliationService]
    end

    subgraph "Domain Layer"
        DE[Entities: Account, Transaction, Category]
        DI[Interfaces: Repository & Usecase]
        DR[Domain Errors]
    end

    subgraph "Repository Layer"
        AR[AccountRepository]
        TR[TransactionRepository]
        CR[CategoryRepository]
    end

    subgraph "Infrastructure"
        DB[(PostgreSQL)]
        MW[Auth Middleware]
    end

    AH --> AU
    TH --> TU
    DH --> DU
    CH --> CU
    OH --> OU

    AU --> DI
    TU --> DI
    DU --> DI
    CU --> DI
    OU --> DI
    RU --> DI

    AR --> DB
    TR --> DB
    CR --> DB

    MW --> AH
    MW --> TH
    MW --> DH
    MW --> CH
    MW --> OH
```

### Request Flow

```mermaid
sequenceDiagram
    participant C as Client
    participant MW as AuthMiddleware
    participant H as Handler
    participant UC as Usecase
    participant R as Repository
    participant DB as PostgreSQL

    C->>MW: HTTP Request + Bearer Token
    MW->>MW: Validate JWT, extract user_id
    MW->>H: Set user_id in context
    H->>H: Parse & validate request body
    H->>UC: Call usecase method(ctx, params)
    UC->>UC: Business logic & validation
    UC->>R: Database operations
    R->>DB: SQL query (within tx if needed)
    DB-->>R: Result
    R-->>UC: Domain entities
    UC-->>H: Result / error
    H-->>C: JSON response (via response package)
```

## Getting Started

### Prerequisites

- Go 1.26+
- Docker & Docker Compose
- Make
- A Google Cloud project with OAuth 2.0 credentials (for SSO)

### Setup

```bash
# Clone the repo
git clone https://github.com/mzhryns/titik-nol-backend.git
cd titik-nol-backend

# Copy env and fill in your values
cp .env.example .env

# Start development environment (with hot-reload)
make docker-up
```

The API will be available at `http://localhost:8080`.

### Running Locally (without Docker)

```bash
# Make sure PostgreSQL is running and .env is configured
make run
```

### Production

```bash
# Build and start production containers
make docker-prod-up
```

Production uses a hardened setup: non-root user, stripped binary, resource limits, read-only filesystem, no exposed DB port, and JSON logging.

### Environment Variables

Copy [`.env.example`](.env.example) to `.env` and fill in your values. The file is self-documented with inline comments.

## API Endpoints

The interactive API documentation is available via **Scalar**:

- **UI:** `/docs/api`
- **OpenAPI Spec:** `/docs/swagger.json`

> **Note:** Run `make swagger` to regenerate the documentation after any handler changes.

Route groups overview:

- `/health` — Health check
- `/auth` — Google SSO login & current user
- `/api/v1/accounts` — Account CRUD 🔒
- `/api/v1/transactions` — Transaction CRUD 🔒
- `/api/v1/categories` — Category management 🔒
- `/api/v1/onboarding` — Initial account setup 🔒
- `/api/v1/dashboard` — Financial summary 🔒

> 🔒 = Requires `Authorization: Bearer <token>` header.

## Database Schema

```mermaid
erDiagram
    users {
        UUID id PK
        VARCHAR email UK
        VARCHAR name
        TEXT avatar_url
        VARCHAR provider
        VARCHAR provider_id UK
        TIMESTAMPTZ created_at
        TIMESTAMPTZ updated_at
    }

    accounts {
        UUID id PK
        UUID user_id FK
        VARCHAR name
        account_type_enum type "CASH | BANK | E_WALLET | CREDIT_CARD"
        BIGINT balance "smallest unit"
        TIMESTAMPTZ created_at
        TIMESTAMPTZ updated_at
        TIMESTAMPTZ deleted_at "soft delete"
    }

    categories {
        UUID id PK
        UUID user_id FK
        VARCHAR name
        category_type_enum type "INCOME | EXPENSE"
        VARCHAR icon
        TIMESTAMPTZ created_at
    }

    transactions {
        UUID id PK
        UUID user_id FK
        UUID account_id FK
        UUID category_id FK "nullable"
        tx_type_enum transaction_type "INCOME | EXPENSE | TRANSFER | ADJUSTMENT"
        BIGINT amount
        TEXT note
        TIMESTAMPTZ transaction_date
        TIMESTAMPTZ created_at
        TIMESTAMPTZ deleted_at "soft delete"
    }

    users ||--o{ accounts : has
    users ||--o{ categories : has
    users ||--o{ transactions : has
    accounts ||--o{ transactions : has
    categories ||--o{ transactions : has
```

## Make Commands

Run `make help` to list all available commands.

## Development Guidelines

- [API Response Standard](docs/api-response-standard.md) — all endpoints use the shared `response` package (RFC 7807 errors)
- [Testing Guidelines](docs/testing-guidelines.md) — top-level test functions, no nested subtests
- [Logger Guidelines](docs/logger.md) — context-aware `slog` usage
- [Git Commit Rules](docs/git-commit-rules.md) — Conventional Commits format
- [Google ID Setup Guide](docs/google-id-setup.md) — how to configure OAuth 2.0 credentials

## License

This project is open source under the [MIT License](LICENSE).

If you fork or modify this project, please give credit by linking back to the original repository and mentioning the author.

Built by [@zuhriyansauqi](https://github.com/zuhriyansauqi).
