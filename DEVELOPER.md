# Developer Guide

Development documentation for **Elotus Backend Test** project.

---

## Table of Contents

1. [Architecture Overview](#architecture-overview)
2. [Directory Structure](#directory-structure)
3. [Environment Setup](#environment-setup)
4. [Code Conventions](#code-conventions)
5. [Adding New Features](#adding-new-features)
6. [Database & Migrations](#database--migrations)
7. [Testing](#testing)
8. [API Response Format](#api-response-format)
9. [Authentication Flow](#authentication-flow)
10. [Deployment](#deployment)

---

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                         Client (Browser/Mobile)                  │
└───────────────────────────────────┬─────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────┐
│                         Echo HTTP Server                         │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────────┐  │
│  │   Logging   │  │ Rate Limit  │  │    JWT Middleware       │  │
│  │  Middleware │  │ Middleware  │  │    (Protected routes)   │  │
│  └─────────────┘  └─────────────┘  └─────────────────────────┘  │
└───────────────────────────────────┬─────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────┐
│                           Handlers                               │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐  │
│  │  Auth Handler   │  │ Upload Handler  │  │  Health Handler │  │
│  │ (login/register)│  │  (file upload)  │  │  (status check) │  │
│  └─────────────────┘  └─────────────────┘  └─────────────────┘  │
└───────────────────────────────────┬─────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────┐
│                          Services                                │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐  │
│  │   JWT Service   │  │Token Revocation │  │   Validation    │  │
│  │ (generate/verify│  │     Store       │  │    Service      │  │
│  └─────────────────┘  └─────────────────┘  └─────────────────┘  │
└───────────────────────────────────┬─────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────┐
│                        Repositories                              │
│  ┌─────────────────┐  ┌─────────────────┐                       │
│  │ User Repository │  │Upload Repository│                       │
│  │   (interface)   │  │   (interface)   │                       │
│  └─────────────────┘  └─────────────────┘                       │
└───────────────────────────────────┬─────────────────────────────┘
                                    │
                    ┌───────────────┴───────────────┐
                    ▼                               ▼
┌───────────────────────────┐       ┌───────────────────────────┐
│        PostgreSQL         │       │          Redis            │
│     (Primary Storage)     │       │    (Cache/Rate Limit)     │
└───────────────────────────┘       └───────────────────────────┘
```

### Main Components

| Component   | Package         | Description              |
|-------------|-----------------|--------------------------|
| HTTP Server | `echo/v4`       | Web framework            |
| Database    | `bsql`          | PostgreSQL wrapper       |
| Cache       | `bredis`        | Redis wrapper            |
| Auth        | `models/auth`   | JWT authentication       |
| Upload      | `models/upload` | File upload handling     |
| Logging     | `logger`        | Zerolog wrapper          |
| Validation  | `validation`    | Input validation         |
| Response    | `response`      | Unified API response     |

---

## Directory Structure

```
elotus_test/
├── dsa/                              # DSA challenges (not related to server)
│   ├── gray-code/
│   ├── maximum-length-of-repeated-subarray/
│   └── sum-of-distances-in-tree/
│
├── server/
│   ├── main.go                       # Entry point
│   │
│   ├── models/                       # Business logic
│   │   ├── models.go                 # App initialization & dependency injection
│   │   ├── router.go                 # Route definitions
│   │   ├── auth/                     # Authentication module
│   │   │   ├── handler.go            # HTTP handlers
│   │   │   ├── jwt.go                # JWT service
│   │   │   └── revocation.go         # Token revocation store
│   │   ├── upload/                   # File upload module
│   │   │   ├── handler.go            # HTTP handlers
│   │   │   ├── upload.go             # Domain models & interfaces
│   │   │   └── postgres_repository.go
│   │   └── user/                     # User module
│   │       ├── user.go               # Domain models & interfaces
│   │       └── postgres_repository.go
│   │
│   ├── middleware/                   # HTTP middlewares
│   │   ├── jwt.go                    # JWT validation
│   │   ├── ratelimit.go              # Rate limiting
│   │   └── logging.go                # Request logging
│   │
│   ├── response/                     # Unified response format
│   │   └── response.go
│   │
│   ├── validation/                   # Input validation
│   │   └── validation.go
│   │
│   ├── bsql/                         # PostgreSQL wrapper
│   │   ├── bsql.go                   # DB operations
│   │   └── open.go                   # Config loading
│   │
│   ├── bredis/                       # Redis wrapper
│   │   └── bredis.go
│   │
│   ├── psql/                         # Migration tools
│   │   └── migrate.go
│   │
│   ├── logger/                       # Logging utilities
│   │   └── logger.go
│   │
│   ├── env/                          # Environment config
│   │   └── env.go
│   │
│   ├── renv/                         # Config parser
│   │   └── renv.go
│   │
│   ├── cmd/                          # CLI commands
│   │   ├── db.go                     # Database commands
│   │   └── path.go                   # Path utilities
│   │
│   ├── db/                           # Database configs & migrations
│   │   ├── migrations/               # SQL migration files
│   │   ├── database.yaml             # Local DB config
│   │   ├── database.docker.yaml      # Docker DB config
│   │   ├── redis.yaml                # Local Redis config
│   │   └── redis.docker.yaml         # Docker Redis config
│   │
│   ├── html/                         # Static web files (for testing)
│   │   ├── index.html
│   │   ├── register.html
│   │   ├── dashboard.html
│   │   └── uploads.html
│   │
│   └── tests/                        # Unit tests
│       ├── mocks.go                  # Mock repositories
│       ├── auth_handler_test.go
│       ├── auth_jwt_test.go
│       ├── middleware_test.go
│       ├── upload_test.go
│       ├── user_test.go
│       └── run_tests.sh              # Test runner script
│
├── docker-compose.yml                # Docker services
├── Dockerfile                        # Multi-stage build
├── Makefile                          # Helper commands
├── go.mod
├── go.sum
├── README.md                         # User documentation
├── CHALLENGE.md                      # Test requirements
└── DEVELOPER.md                      # This file
```

---

## Environment Setup

### Prerequisites

- Go 1.22+
- PostgreSQL 15+
- Redis 7+ (optional)
- Docker & Docker Compose (optional)

### Local Development

```bash
# 1. Clone repository
git clone <repo-url>
cd elotus_test

# 2. Configure database
cp server/db/database.sample.yaml server/db/database.yaml
# Edit PostgreSQL connection info

# 3. Configure Redis (optional)
cp server/db/redis.sample.yaml server/db/redis.yaml

# 4. Create environment file
cp server/.env.sample.yaml server/.env.local.yaml
# Edit JWT secret and other configs

# 5. Run migrations
go run ./server -db migrate

# 6. Start server
go run ./server

# Server will run at http://localhost:8080
```

### Docker Development

```bash
# Build and start all services
make up

# View logs
make logs

# Stop services
make down

# List available commands
make help
```

---

## Code Conventions

### Naming Conventions

```go
// Package names: lowercase, single word
package auth
package upload

// Interface names: verb + "er" or noun
type Repository interface { ... }
type Validator interface { ... }

// Struct names: PascalCase
type TokenClaims struct { ... }
type FileUpload struct { ... }

// Function names: PascalCase for exported, camelCase for internal
func NewHandler() { ... }      // exported
func validateFile() { ... }    // internal

// Constants: PascalCase for exported
const MaxFileSize = 8 * 1024 * 1024

// Variables: camelCase
var userRepo Repository
```

### File Organization

Each module should have this structure:

```
module/
├── model.go           # Domain models, interfaces, errors
├── handler.go         # HTTP handlers
├── service.go         # Business logic (if needed)
└── postgres_repository.go  # Database implementation
```

### Error Handling

```go
// Define errors at package level
var (
    ErrUserNotFound = errors.New("user not found")
    ErrUserExists   = errors.New("username already exists")
)

// Return errors, don't panic
func (r *Repository) GetUser(id int64) (*User, error) {
    // ...
    if err != nil {
        return nil, fmt.Errorf("failed to get user: %w", err)
    }
    return user, nil
}

// Handlers convert errors to HTTP responses
func (h *Handler) GetUser(c echo.Context) error {
    user, err := h.repo.GetUser(id)
    if err != nil {
        if errors.Is(err, ErrUserNotFound) {
            return response.NotFound(c, "User not found")
        }
        return response.InternalError(c, "Failed to get user")
    }
    return response.Success(c, user)
}
```

### Comments

Only comment complex or non-obvious code:

```go
// ✅ Good - explains complex logic
// CheckRateLimit implements sliding window rate limiting using Redis INCR + EXPIRE
func (c *Client) CheckRateLimit(...) { ... }

// ✅ Good - explains magic numbers
// pq error code 23505 = unique_violation
if pqErr.Code == "23505" { ... }

// ❌ Bad - unnecessary comment
// CreateUser creates a new user
func CreateUser() { ... }
```

---

## Adding New Features

### 1. Adding a New API Endpoint

**Example: Add GET /api/users/:id endpoint**

```go
// 1. Add method to Repository interface (server/models/user/user.go)
type Repository interface {
    // ... existing methods
    GetUserByID(id int64) (*User, bool)
}

// 2. Implement in postgres_repository.go
func (r *PostgresRepository) GetUserByID(id int64) (*User, bool) {
    var user User
    err := r.db.QueryRow(`SELECT id, username, created_at FROM users WHERE id = $1`, id).
        Scan(&user.ID, &user.Username, &user.CreatedAt)
    if err != nil {
        return nil, false
    }
    return &user, true
}

// 3. Add handler (in auth/handler.go or create new file)
func (h *Handler) GetUser(c echo.Context) error {
    idStr := c.Param("id")
    id, err := strconv.ParseInt(idStr, 10, 64)
    if err != nil {
        return response.BadRequest(c, "Invalid user ID")
    }

    user, exists := h.userRepo.GetUserByID(id)
    if !exists {
        return response.NotFound(c, "User not found")
    }

    return response.Success(c, user)
}

// 4. Register route (server/models/router.go)
func (m *Models) SetupRoutes() {
    // ... existing routes
    api.GET("/users/:id", m.authHandler.GetUser)
}

// 5. Write tests (server/tests/)
func TestGetUser_Success(t *testing.T) {
    // ...
}
```

### 2. Adding New Middleware

```go
// server/middleware/new_middleware.go
package middleware

func NewMiddleware(config Config) echo.MiddlewareFunc {
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            // Pre-processing

            err := next(c)  // Call next handler

            // Post-processing

            return err
        }
    }
}

// Register in router.go
e.Use(middleware.NewMiddleware(config))
```

### 3. Adding a New Module

```bash
# Create new directory
mkdir -p server/models/newmodule

# Create required files
touch server/models/newmodule/model.go
touch server/models/newmodule/handler.go
touch server/models/newmodule/postgres_repository.go
```

```go
// server/models/newmodule/model.go
package newmodule

type Entity struct {
    ID        int64     `json:"id"`
    Name      string    `json:"name"`
    CreatedAt time.Time `json:"created_at"`
}

type Repository interface {
    Create(entity *Entity) (*Entity, error)
    GetByID(id int64) (*Entity, bool)
    List() ([]*Entity, error)
}

var (
    ErrNotFound = errors.New("entity not found")
)
```

---

## Database & Migrations

### Creating a New Migration

```bash
# Using CLI
go run ./server -db generate -name "create_orders_table"

# Or create file manually
# server/db/migrations/20251207120000_create_orders_table.sql
```

```sql
-- Migration: Create orders table
-- Created at: 2025-12-07

-- +migrate Up
CREATE TABLE IF NOT EXISTS orders (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    total_amount DECIMAL(10, 2) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_orders_user_id ON orders(user_id);
CREATE INDEX IF NOT EXISTS idx_orders_status ON orders(status);

-- +migrate Down
DROP INDEX IF EXISTS idx_orders_status;
DROP INDEX IF EXISTS idx_orders_user_id;
DROP TABLE IF EXISTS orders;
```

### Migration Commands

```bash
# Run all pending migrations
go run ./server -db migrate

# Rollback 1 migration
go run ./server -db rollback

# Rollback n migrations
go run ./server -db rollback -steps 3

# Check status
go run ./server -db status
```

### Database Conventions

- Table names: plural, snake_case (`users`, `file_uploads`)
- Column names: snake_case (`created_at`, `user_id`)
- Primary key: `id SERIAL PRIMARY KEY`
- Foreign keys: `<table>_id` with `REFERENCES` constraint
- Timestamps: `TIMESTAMP WITH TIME ZONE`
- Always create indexes for foreign keys

---

## Testing

### Running Tests

```bash
# Run all tests
go test ./server/tests/... -v

# Run specific test
go test ./server/tests/... -v -run TestLogin

# Run with coverage
go test ./server/tests/... -coverprofile=coverage.out
go tool cover -html=coverage.out

# Use test script
./server/tests/run_tests.sh
```

### Writing New Tests

```go
// server/tests/new_test.go
package tests

import (
    "testing"
    "net/http"
    "net/http/httptest"

    "github.com/labstack/echo/v4"
)

func TestNewFeature_Success(t *testing.T) {
    // 1. Setup
    handler, mockRepo := setupHandler()
    mockRepo.AddData(...)

    // 2. Create request
    e := echo.New()
    req := httptest.NewRequest(http.MethodGet, "/endpoint", nil)
    rec := httptest.NewRecorder()
    c := e.NewContext(req, rec)

    // 3. Execute
    err := handler.Method(c)

    // 4. Assert
    if err != nil {
        t.Fatalf("Handler returned error: %v", err)
    }
    if rec.Code != http.StatusOK {
        t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
    }

    // 5. Verify response
    resp, _ := parseResponse(rec.Body.Bytes())
    if !resp.Success {
        t.Error("Expected success response")
    }
}

func TestNewFeature_Error(t *testing.T) {
    // Test error cases
}
```

### Mock Repositories

```go
// Add new mock to server/tests/mocks.go
type MockNewRepository struct {
    mu     sync.RWMutex
    data   map[int64]*Entity
    nextID int64
    // Error hooks for testing
    CreateError error
}

func NewMockNewRepository() *MockNewRepository {
    return &MockNewRepository{
        data:   make(map[int64]*Entity),
        nextID: 1,
    }
}

func (r *MockNewRepository) Create(e *Entity) (*Entity, error) {
    if r.CreateError != nil {
        return nil, r.CreateError
    }
    r.mu.Lock()
    defer r.mu.Unlock()

    e.ID = r.nextID
    r.nextID++
    r.data[e.ID] = e
    return e, nil
}
```

---

## API Response Format

### Using the Response Package

```go
import "elotus_test/server/response"

// Success responses
return response.Success(c, data)                    // 200 OK
return response.Created(c, data)                    // 201 Created
return response.SuccessWithMeta(c, data, &response.Meta{Total: 10})

// Error responses
return response.BadRequest(c, "Invalid input")      // 400
return response.ValidationError(c, "Field required") // 400
return response.Unauthorized(c, "Token required")   // 401
return response.Forbidden(c, "Access denied")       // 403
return response.NotFound(c, "Resource not found")   // 404
return response.Conflict(c, "Already exists")       // 409
return response.TooManyRequests(c, "Rate limited", 60) // 429
return response.InternalError(c, "Server error")    // 500
```

### Response Structure

```json
// Success
{
  "success": true,
  "data": { ... },
  "meta": {
    "total": 10,
    "cached": false
  }
}

// Error
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Password must be at least 8 characters"
  }
}
```

---

## Authentication Flow

### JWT Token Flow

```
1. User registers
   POST /register
   └─> Create user in DB
   └─> Return success

2. User logs in
   POST /login
   └─> Validate credentials
   └─> Check rate limit (Redis)
   └─> Generate JWT token (HS256)
   └─> Return token + expiry

3. User accesses protected endpoint
   GET /api/protected
   Header: Authorization: Bearer <token>
   └─> JWT Middleware validates token
   └─> Check if token is revoked
   └─> Set claims in context
   └─> Handler processes request

4. User revokes tokens
   POST /api/revoke
   └─> Update last_revoked_token_at in DB
   └─> Cache revocation time in Redis
   └─> All tokens issued before this time are invalid
```

### Token Revocation Logic

```go
// Token is revoked if:
// issuedAt < user.last_revoked_token_at

// Check order:
// 1. Check Redis cache for revocation time
// 2. If not in cache, check DB
// 3. Cache result in Redis for performance
```

---

## Deployment

### Docker Production

```bash
# Build and deploy
docker-compose up -d --build

# Scale app (if needed)
docker-compose up -d --scale app=3
```

### Environment Variables

```yaml
# server/.env.local.yaml (development)
environment: "development"
server_name: "elotus-dev"

# Production (set via docker-compose or K8s)
environment: "production"
server_name: "elotus-prod"
jwt_signing_key: "<strong-secret-key>"  # IMPORTANT: Change this!
```

### Health Check

```bash
# Check server status
curl http://localhost:8080/health

# Response when healthy
{
  "success": true,
  "data": {
    "status": "UP",
    "db": "OK",
    "redis": "OK",
    "timestamp": "2025-12-07T12:00:00+07:00"
  }
}
```

---

## Troubleshooting

### Common Issues

**1. Database connection failed**

```bash
# Check PostgreSQL is running
pg_isready -h localhost -p 5432

# Check connection string in database.yaml
# Ensure database exists
createdb elotus_test
```

**2. Redis connection failed**

```bash
# Check Redis is running
redis-cli ping

# Server still works without Redis (graceful degradation)
# Rate limiting and caching will be disabled
```

**3. Migration failed**

```bash
# Check migration status
go run ./server -db status

# Rollback and retry
go run ./server -db rollback
go run ./server -db migrate
```

**4. Test failures**

```bash
# Run with verbose output
go test ./server/tests/... -v

# Check specific test
go test ./server/tests/... -v -run TestFunctionName
```

---

## Resources

- [Echo Framework Documentation](https://echo.labstack.com/guide/)
- [golang-jwt Documentation](https://pkg.go.dev/github.com/golang-jwt/jwt/v5)
- [go-redis Documentation](https://redis.uptrace.dev/guide/)
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
