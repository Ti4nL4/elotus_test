# Elotus Backend Test

A Go backend server with JWT authentication and file upload features.

> See [CHALLENGE.md](CHALLENGE.md) for test requirements.

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
| Upload image         | `POST /api/upload` | `server/models/upload/handler.go` | ✅     |

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
| POST   | `/api/revoke`      | Revoke tokens by time       | Yes           |
| GET    | `/api/protected`   | Test protected endpoint     | Yes           |
| POST   | `/api/upload`      | Upload image file (max 8MB) | Yes           |
| GET    | `/api/uploads`     | List user's uploads         | Yes           |
| GET    | `/api/uploads/:id` | Get specific upload         | Yes           |
| GET    | `/health`          | Health check                | No            |

## API Usage Examples

### Register

```bash
curl -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","password":"123456"}'
```

### Login

```bash
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","password":"123456"}'
```

### Upload Image

```bash
curl -X POST http://localhost:8080/api/upload \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -F "image=@/path/to/image.jpg"
```

### Revoke Tokens

```bash
curl -X POST http://localhost:8080/api/revoke \
  -H "Authorization: Bearer YOUR_TOKEN"
```

---

## Running Tests

```bash
go test ./server/tests/... -v
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
│   ├── main.go                   # Entry point
│   ├── models/
│   │   ├── auth/                 # Authentication (JWT, handlers)
│   │   ├── upload/               # File upload feature
│   │   ├── user/                 # User repository
│   │   ├── models.go             # App initialization
│   │   └── router.go             # Routes setup
│   ├── middleware/               # JWT, Rate limit, Logging
│   ├── db/
│   │   ├── migrations/           # SQL migrations
│   │   ├── database.yaml         # DB config
│   │   └── redis.yaml            # Redis config
│   ├── bredis/                   # Redis wrapper
│   ├── bsql/                     # SQL wrapper
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

---

## Design Decisions

### JWT Token Revocation

- Stores `last_revoked_token_at` timestamp in users table
- All tokens issued before this timestamp are considered revoked
- Uses Redis cache for performance (falls back to DB if Redis unavailable)

### File Upload

- Validates content type by reading file header (not just extension)
- Maximum file size: 8MB
- Stores metadata in PostgreSQL including all HTTP information
- Files saved to `tmp/images/` directory

### Rate Limiting

- IP-based rate limiting via Redis
- User-based login attempt limiting
- Graceful degradation when Redis is unavailable
