# ADR 0003: Use Explicit Status Transition Use Cases

## Status

Accepted

## Context

Order status transitions are business intents, not generic field updates.
Different transitions can evolve with different requirements.

## Decision

Expose one use case per transition:

- `MarkOrderAsPaidUseCase`
- `MarkOrderAsCancelledUseCase`
- `MarkOrderAsShippedUseCase`

The HTTP API mirrors these intents with action routes:

- `POST /orders/{id}/pay`
- `POST /orders/{id}/cancel`
- `POST /orders/{id}/ship`

## Consequences

- Use cases stay small and intention-revealing.
- No generic status update switch is needed.
- Adding transition-specific behavior later has a clear place.
