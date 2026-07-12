# ADR 0001: Use Hexagonal Architecture

## Status

Accepted

## Context

The project is a learning-oriented order service that should keep business
rules independent from transport, persistence, and framework choices.

## Decision

Use hexagonal architecture:

- `internal/domain` contains business invariants and behavior.
- `internal/application` contains use cases and ports.
- adapters implement external concerns and depend inward.
- `cmd/orderservice` wires the application.

## Consequences

- Domain and application tests are fast and infrastructure-free.
- Replacing PostgreSQL or HTTP should not affect domain logic.
- Some simple flows require more explicit mapping code between DTOs, use
  cases, and domain types.
