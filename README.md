# Codda — Order Service

A simplified order management API, built in Go following a hexagonal architecture.

## Status

Under development

## Stack

- **Language:** Go 1.26+
- **HTTP Router:**
- **Database:**
- **Integration testing:** testcontainers-go

## Architecture

The project is organized around a hexagonal architecture (Ports and Adapters):

- `internal/domain/` — entities, value objects, and invariants (pure core).
- `internal/application/` — use cases and output ports.
- `internal/adapters/` — concrete implementations (HTTP, PostgreSQL, in-memory).
- `internal/config/` — configuration loading and validation.
- `cmd/orderservice/` — composition root.

## Getting Started

## Testing
