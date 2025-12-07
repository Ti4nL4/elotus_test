# Elotus Backend Test

A Go backend server with JWT authentication and file upload features.

> 📋 See [CHALLENGE.md](CHALLENGE.md) for test requirements.
> 
> 👨‍💻 See [DEVELOPER.md](DEVELOPER.md) for developer documentation.

---

## Implementation Summary

### DSA Challenges

| Challenge                           | Location                                                | Status |
| ----------------------------------- | ------------------------------------------------------- | ------ |
| Gray Code                           | `dsa/gray-code/solution.go`                           | ✅     |
| Sum of Distances in Tree            | `dsa/sum-of-distances-in-tree/solution.go`            | ✅     |
| Maximum Length of Repeated Subarray | `dsa/maximum-length-of-repeated-subarray/solution.go` | ✅     |

### Hackathon Features

| Feature              | Endpoint             | Location                            | Status |
| -------------------- | -------------------- | ----------------------------------- | ------ |
| Register             | `POST /register`   | `server/models/auth/handler.go`   | ✅     |
| Login                | `POST /login`      | `server/models/auth/handler.go`   | ✅     |
| Revoke token by time | `POST /api/revoke` | `server/models/auth/handler.go`   | ✅     |
| Upload image         | `POST /upload`     | `server/models/upload/handler.go` | ✅     |

---

## Quick Start (Docker)

```bash
# Start all services
docker-compose up -d

# Open in browser
open http://localhost:8080
```

## Local Development

### Prerequisites

- Go 1.22+
- PostgreSQL 15+
- Redis 7+ (optional)

### Setup

```bash
# 1. Configure database
cp server/db/database.sample.yaml server/db/database.yaml
# Edit with your PostgreSQL credentials

# 2. Configure Redis (optional)
cp server/db/redis.sample.yaml server/db/redis.yaml

# 3. Create environment file
cp server/.env.sample.yaml server/.env.local.yaml

# 4. Run server
cd server
go run main.go
```

---

## API Endpoints

| Method | Endpoint             | Description                 | Auth Required |
| ------ | -------------------- | --------------------------- | ------------- |
| POST   | `/register`        | Register new user           | No            |
| POST   | `/login`           | Login and get JWT token     | No            |
| POST   | `/upload`          | Upload image (field: "data")| Yes           |
| POST   | `/api/revoke`      | Revoke tokens by time       | Yes           |
| GET    | `/api/protected`   | Test protected endpoint     | Yes           |
| POST   | `/api/upload`      | Upload image (alternative)  | Yes           |
| GET    | `/api/uploads`     | List user's uploads         | Yes           |
| GET    | `/api/uploads/:id` | Get specific upload         | Yes           |
| GET    | `/health`          | Health check                | No            |

---

## API Response Format

All API responses follow a unified format:

### Success Response
```json
{
  "success": true,
  "data": { ... },
  "meta": {
    "total": 10,
    "cached": false
  }
}
```

### Error Response
```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Password must be at least 8 characters"
  }
}
```

### Error Codes
| Code | Description |
| ---- | ----------- |
| `BAD_REQUEST` | Invalid request |
| `VALIDATION_ERROR` | Input validation failed |
| `UNAUTHORIZED` | Authentication required |
| `FORBIDDEN` | Access denied |
| `NOT_FOUND` | Resource not found |
| `CONFLICT` | Resource already exists |
| `TOO_MANY_REQUESTS` | Rate limit exceeded |
| `INTERNAL_ERROR` | Server error |

---

## API Usage Examples

### Register

```bash
curl -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","password":"Password123"}'
```

**Password Requirements:**
- Minimum 8 characters
- At least one uppercase letter
- At least one lowercase letter
- At least one digit

**Username Requirements:**
- 3-50 characters
- Only letters, numbers, and underscores

### Login

```bash
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","password":"Password123"}'
```

### Upload Image

```bash
# Using field name "data" as per challenge requirements
curl -X POST http://localhost:8080/upload \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -F "data=@/path/to/image.jpg"
```

### Revoke Tokens

```bash
curl -X POST http://localhost:8080/api/revoke \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### Health Check

```bash
curl http://localhost:8080/health
```

**Response:**
```json
{
  "success": true,
  "data": {
    "status": "UP",
    "db": "OK",
    "redis": "OK",
    "timestamp": "2025-12-07T21:45:00+07:00"
  }
}
```

**Status Values:**
| Status | Description |
| ------ | ----------- |
| `UP` | All services healthy |
| `DEGRADED` | Some services have issues |
| `OK` | Component is working |
| `DOWN` | Component is not responding |
| `DISABLED` | Redis not configured (system still works) |

---

## Running Tests

```bash
# Run all tests
go test ./server/tests/... -v

# Run with summary script
./server/tests/run_tests.sh
```

---

## Project Structure

```
elotus_test/
├── dsa/                          # DSA challenges
│   ├── gray-code/
│   ├── maximum-length-of-repeated-subarray/
│   └── sum-of-distances-in-tree/
├── server/
│   ├── main.go                   # Entry point with graceful shutdown
│   ├── models/
│   │   ├── auth/                 # Authentication (JWT, handlers)
│   │   ├── upload/               # File upload feature
│   │   ├── user/                 # User repository
│   │   ├── models.go             # App initialization
│   │   └── router.go             # Routes setup
│   ├── middleware/               # JWT, Rate limit, Logging
│   ├── response/                 # Unified API response format
│   ├── validation/               # Input validation utilities
│   ├── db/
│   │   ├── migrations/           # SQL migrations
│   │   ├── database.yaml         # DB config
│   │   └── redis.yaml            # Redis config
│   ├── bredis/                   # Redis wrapper
│   ├── bsql/                     # SQL wrapper
│   ├── logger/                   # Zerolog wrapper
│   ├── tests/                    # Unit tests
│   └── html/                     # Web UI for testing
├── docker-compose.yml            # Docker deployment
├── Dockerfile                    # Multi-stage build
├── Makefile                      # Helper commands
├── CHALLENGE.md                  # Test requirements
└── README.md                     # This file
```

---

## Tech Stack

- **Language**: Go 1.22
- **Web Framework**: Echo v4
- **Database**: PostgreSQL 15
- **Cache**: Redis 7 (optional)
- **JWT**: github.com/golang-jwt/jwt/v5 with HS256
- **Password Hashing**: bcrypt
- **Logging**: zerolog

---

## Design Decisions

### Input Validation

Strong validation for user inputs:
- **Username**: 3-50 chars, alphanumeric + underscore only
- **Password**: 8+ chars, requires uppercase, lowercase, and digit
- Uses dedicated `server/validation/` package

### Unified API Response

All endpoints return consistent JSON format with:
- `success`: Boolean indicating success/failure
- `data`: Response payload (on success)
- `error`: Error details with code and message (on failure)
- `meta`: Metadata like pagination info

### JWT Token Revocation

- Stores `last_revoked_token_at` timestamp in users table
- All tokens issued before this timestamp are considered revoked
- Uses Redis cache for performance (falls back to DB if Redis unavailable)

### File Upload

- **Field name**: `data` (as per challenge requirements)
- Validates content type by reading file header (not just extension)
- Maximum file size: 8MB
- Stores metadata in PostgreSQL including all HTTP information
- Files saved to `tmp/images/` directory
- Fallback extension detection from content-type

### Rate Limiting

- IP-based rate limiting via Redis
- User-based login attempt limiting (5 attempts per 15 minutes)
- Graceful degradation when Redis is unavailable
- Rate limit headers in responses (`X-RateLimit-Limit`, `X-RateLimit-Remaining`)

### Graceful Shutdown

- Handles `SIGINT` and `SIGTERM` signals
- 30-second timeout for graceful shutdown
- Properly closes HTTP server, Redis, and database connections

---

## Docker Commands (Makefile)

```bash
make help       # Show all commands
make build      # Build Docker images
make up         # Start all services
make down       # Stop all services
make logs       # View logs
make psql       # Connect to PostgreSQL
make redis-cli  # Connect to Redis CLI
make test       # Run tests
```
