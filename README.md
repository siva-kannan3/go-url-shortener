# Go URL Shortener

A simple URL shortener backend built with Go and PostgreSQL.

This project is primarily a learning project for understanding Go backend development, HTTP APIs, PostgreSQL, database migrations, repository patterns, testing, Docker, and deployment.

## Tech Stack

- Go
- PostgreSQL
- pgx
- `database/sql`
- golang-migrate
- Docker
- Docker Compose

## Project Structure

```text
.
├── cmd/
│   └── server/
│       └── main.go
│
├── internal/
│   ├── config/
│   │   └── config.go
│   │
│   ├── database/
│   │   └── postgres.go
│   │
│   ├── domain/
│   │   └── url.go
│   │
│   ├── handler/
│   │   ├── router.go
│   │   └── router_test.go
│   │
│   ├── repository/
│   │   ├── postgres.go
│   │   └── postgres_test.go
│   │
│   └── service/
│       ├── id.go
│       ├── service.go
│       └── service_test.go
│
├── migrations/
│   ├── 000001_create_urls_table.up.sql
│   └── 000001_create_urls_table.down.sql
│
├── docker-compose.yml
├── go.mod
└── go.sum
```

## Architecture

The application follows a layered architecture:

```text
HTTP Handler
     ↓
   Service
     ↓
Repository Interface
     ↓
Postgres Repository
     ↓
PostgreSQL
```

### Handler

Responsible for HTTP-specific concerns such as:

- Parsing HTTP requests
- Validating request format
- Returning HTTP responses
- Handling redirects

### Service

Contains application/business logic such as:

- URL validation
- Short ID generation
- Retrying when a generated ID already exists

The service depends on the `URLRepository` interface rather than directly depending on PostgreSQL.

### Repository

Responsible for persistence.

The current production implementation is `PostgresRepository`, which uses PostgreSQL through `database/sql` and the pgx driver.

### Database

PostgreSQL stores shortened URLs.

Current schema:

```sql
CREATE TABLE urls (
    short_id VARCHAR(6) PRIMARY KEY,
    original_url TEXT NOT NULL
);
```

`short_id` is the primary key because it is the application's natural identifier for a shortened URL.

## Prerequisites

Install the following:

- Go
- Docker
- Docker Compose
- PostgreSQL client (`psql`)
- golang-migrate CLI

## Running PostgreSQL

PostgreSQL runs through Docker Compose.

Start the database:

```bash
docker compose up -d
```

The PostgreSQL container is exposed on port `5433` on the host to avoid conflicts with a local PostgreSQL installation.

Check the container:

```bash
docker compose ps
```

Connect to the database:

```bash
psql -h localhost -p 5433 -U postgres -d url_shortener
```

The local development database configuration is:

```text
postgres://postgres:postgres@localhost:5433/url_shortener?sslmode=disable
```

Do not commit database credentials or connection strings containing credentials to the repository.

## Database Migrations

The project uses [golang-migrate](https://github.com/golang-migrate/migrate) to manage database schema changes.

Apply all pending migrations:

```bash
migrate \
  -path migrations \
  -database "postgres://postgres:postgres@localhost:5433/url_shortener?sslmode=disable" \
  up
```

Check the migration state:

```bash
migrate \
  -path migrations \
  -database "postgres://postgres:postgres@localhost:5433/url_shortener?sslmode=disable" \
  version
```

Rollback the latest migration:

```bash
migrate \
  -path migrations \
  -database "postgres://postgres:postgres@localhost:5433/url_shortener?sslmode=disable" \
  down 1
```

Migration files follow this naming convention:

```text
<number>_<description>.up.sql
<number>_<description>.down.sql
```

## Configuration

The application expects the following environment variable:

```text
DATABASE_URL
```

Example for local development:

```bash
export DATABASE_URL="postgres://postgres:postgres@localhost:5433/url_shortener?sslmode=disable"
```

Start the application:

```bash
go run ./cmd/server
```

The server listens on:

```text
http://localhost:8080
```

## API

### Create a Short URL

```http
POST /shorten
Content-Type: application/json
```

Request:

```json
{
  "url": "https://www.google.com"
}
```

Example response:

```json
{
  "id": "abc123",
  "url": "https://www.google.com"
}
```

### Redirect

```http
GET /{id}
```

Example:

```text
GET /abc123
```

The server responds with an HTTP redirect to the original URL.

### Not Found

If the short ID does not exist:

```text
404 Not Found
```

## Running Tests

Run all tests:

```bash
go test ./...
```

Run tests with the race detector:

```bash
go test -race ./...
```

### Test Structure

The project uses different types of tests for different layers.

#### Service tests

Service tests use a mock repository so that business logic can be tested independently of PostgreSQL.

```text
Service
   ↓
Mock Repository
```

#### Repository tests

Repository tests use a real PostgreSQL database.

```text
PostgresRepository
       ↓
   PostgreSQL
```

These tests require `DATABASE_URL` to be set.

#### Handler tests

Handler tests exercise the HTTP layer together with the real PostgreSQL repository.

```text
HTTP
 ↓
Handler
 ↓
Service
 ↓
PostgresRepository
 ↓
PostgreSQL
```

## Connection Pooling

The application uses `database/sql`, which manages a connection pool rather than a single database connection.

Current pool configuration:

```text
Maximum open connections: 10
Maximum idle connections: 5
Connection maximum lifetime: 30 minutes
```

These values are currently intended for local/MVP usage and may need to be adjusted for a production workload.

## Development

Start PostgreSQL:

```bash
docker compose up -d
```

Set the database URL:

```bash
export DATABASE_URL="postgres://postgres:postgres@localhost:5433/url_shortener?sslmode=disable"
```

Run migrations:

```bash
migrate \
  -path migrations \
  -database "$DATABASE_URL" \
  up
```

Run the application:

```bash
go run ./cmd/server
```

Run tests:

```bash
go test ./...
```

## Current Scope

The current MVP provides:

- URL shortening
- Short ID generation
- PostgreSQL persistence
- URL redirection
- Repository abstraction
- Database migrations
- Connection pooling
- Automated tests
- Dockerized PostgreSQL

## Future Improvements

Possible future improvements include:

- Stronger URL validation
- Graceful HTTP server shutdown
- HTTP server timeouts
- Health/readiness endpoints
- Improved configuration management
- Structured logging
- Authentication
- Rate limiting
- URL expiration
- Analytics
- Redis caching
- Load testing
- Observability
- Horizontal scaling

## Learning Goals

This project is being developed incrementally to understand:

- Go project structure
- Go interfaces
- Dependency injection
- HTTP servers and handlers
- Error handling
- Concurrency
- PostgreSQL
- SQL and database constraints
- Database migrations
- Connection pooling
- Integration testing
- Unit testing
- Docker
- Production deployment
- Backend system design
