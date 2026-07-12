# Codda

Codda is an Order Service API built in Go with hexagonal architecture,
domain-driven design, and PostgreSQL persistence.

The service manages customer orders through this lifecycle:

```text
pending -> paid -> shipped
pending -> cancelled
paid    -> cancelled
```

## Project Structure

<img src="docs/images/file-structure.png" alt="Project Structure" width="500" />

## Requirements

- Go 1.26 or later
- Docker and Docker Compose

## Running Locally

Start PostgreSQL:

```sh
docker compose up -d
```

Run the service:

```sh
export DATABASE_URL="postgres://codda:codda@localhost:5433/codda?sslmode=disable"
go run ./cmd/orderservice/
```

The HTTP server listens on `:8080` by default. Set `HTTP_PORT` to override it.
Database migrations are applied automatically on startup.

## Testing

Run fast tests without external infrastructure:

```sh
go test ./internal/domain/... ./internal/application/... ./internal/adapters/memory/... ./internal/adapters/http/... ./internal/config/...
```

Run the full suite, including PostgreSQL integration tests with Testcontainers:

```sh
go test ./...
```

Run the local smoke harness:

```sh
tools/harness/smoke.sh
```

## API Quick Start

Create an order:

```sh
curl -X POST http://localhost:8080/orders \
  -H "Content-Type: application/json" \
  -d '{
    "items": [
      {"product_id": "p1", "product_name": "Widget", "price_cents": 1999, "quantity": 2}
    ]
  }'
```

Get an order:

```sh
curl http://localhost:8080/orders/{id}
```

List orders:

```sh
curl "http://localhost:8080/orders?limit=10&status=paid"
```

Transition an order:

```sh
curl -X POST http://localhost:8080/orders/{id}/pay
curl -X POST http://localhost:8080/orders/{id}/cancel
curl -X POST http://localhost:8080/orders/{id}/ship
```

## Official Documentation

- [Architecture](docs/architecture.md)
- [Domain Model](docs/domain.md)
- [Application Layer](docs/application.md)
- [HTTP API](docs/http-api.md)
- [Persistence](docs/persistence.md)
- [Configuration](docs/configuration.md)
- [Testing](docs/testing.md)
- [Operations](docs/operations.md)
- [Backlog](docs/backlog.md)
- [AI-Assisted Development](docs/ai-assisted-development.md)

Architecture decision records live in [docs/adr](docs/adr/).
