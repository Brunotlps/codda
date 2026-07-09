# Codda

Order Service API built with hexagonal architecture in Go.

## About

Codda is a learning project focused on backend fundamentals:
hexagonal architecture (ports and adapters), domain-driven design,
Go idioms, and integration with PostgreSQL. It exposes a simple
HTTP API for managing customer orders through a lifecycle
(pending → paid → shipped, or cancelled).

## Project Structure

<img src="docs/images/file-structure.png" alt="Project Structure" width="500" />

## Requirements

- Go 1.26 or later
- Docker + Docker Compose

## Running

Start PostgreSQL:

    docker compose up -d

Set the database URL and run the service:

    export DATABASE_URL="postgres://codda:codda@localhost:5433/codda?sslmode=disable"
    go run ./cmd/orderservice/

The service starts on port 8080 (configurable via `HTTP_PORT`). Database
migrations are applied automatically on startup.

## Testing

Unit tests (no external dependencies):

    go test ./internal/domain/... ./internal/application/...

Full test suite (requires Docker for integration tests):

    go test ./...

## API

Create an order:

    curl -X POST http://localhost:8080/orders \
      -H "Content-Type: application/json" \
      -d '{
        "items": [
          {"product_id": "p1", "product_name": "Widget", "price_cents": 1999, "quantity": 2}
        ]
      }'

Get an order:

    curl http://localhost:8080/orders/{id}

List orders:

    curl "http://localhost:8080/orders?limit=10&status=paid"

Transition an order:

    curl -X POST http://localhost:8080/orders/{id}/pay
    curl -X POST http://localhost:8080/orders/{id}/cancel
    curl -X POST http://localhost:8080/orders/{id}/ship

## Configuration

| Variable       | Required | Default | Description                               |
| -------------- | -------- | ------- | ----------------------------------------- |
| `DATABASE_URL` | Yes      | -       | PostgreSQL connection string              |
| `HTTP_PORT`    | No       | 8080    | Port the HTTP server listens on (1-65535) |
