# ADR 0002: Generate Order IDs in the Domain

## Status

Accepted

## Context

Orders are aggregates with identity. The system needs an order ID immediately
after constructing a valid order.

## Decision

Generate order IDs in `domain.NewOrder` with UUIDs. The database persists an
already-created aggregate and does not create IDs.

## Consequences

- `OrderRepository.Save` does not return a generated ID.
- The same repository method can handle create and update.
- Persistence adapters can use UPSERT by aggregate ID.
