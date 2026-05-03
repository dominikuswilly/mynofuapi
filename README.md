# MyNofu API

A robust backend service for MyNofu, built with Go and adhering to Clean Architecture principles.

## Features

- **SOLID, DRY, KISS, and Clean Architecture**: Highly maintainable and testable structure.
- **Authentication**: Initial `[POST] /public/auth` endpoint implemented.
- **Mock Data**: Pre-configured with an admin user for immediate testing.

## Prerequisites

- [Go](https://golang.org/doc/install) (latest version recommended)
- [Make](https://www.gnu.org/software/make/) (optional, but recommended)

## How to Run

### Using Makefile (Recommended)

1. **Install dependencies**:
   ```bash
   go mod tidy
   ```

2. **Start the server**:
   ```bash
   make run
   ```

### Using Go CLI

1. **Start the server**:
   ```bash
   go run cmd/api/main.go
   ```

The server will be available at `http://localhost:8080`.

## API Documentation (Swagger)

The API includes interactive documentation via Swagger UI.

### Access
Once the server is running, you can access the documentation at:
[http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html)

### Regenerating Documentation
To update the documentation after changing annotations or API structure, run:
```bash
go run github.com/swaggo/swag/cmd/swag@latest init -g cmd/api/main.go
```

## Testing the API

You can test the authentication endpoint using `curl`:

```bash
curl -X POST http://localhost:8080/public/auth \
     -H "Content-Type: application/json" \
     -d '{
           "username": "admin",
           "password": "password123"
         }'
```

**Success Response:**
```json
{
  "access_token": "mock-jwt-token",
  "token_type": "Bearer",
  "expires_in": 3600
}
```

## Project Structure

- `cmd/api/`: Entry point of the application.
- `internal/domain/`: Business entities and repository/usecase interfaces.
- `internal/usecase/`: Business logic implementations.
- `internal/repository/`: Data access implementations (PostgreSQL, Mock, etc.).
- `internal/delivery/http/`: HTTP handlers and routing.
- `pkg/`: Shared utility packages.
