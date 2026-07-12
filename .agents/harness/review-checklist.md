# Review Checklist

Use this checklist before finishing an AI-assisted change.

## Architecture

- Domain remains independent from infrastructure.
- Application depends only on domain and standard library.
- New ports live with the consumer.
- Adapters translate external concerns and do not own business rules.

## Domain

- Invariants are enforced in constructors or aggregate methods.
- Invalid states are not made representable through public APIs.
- State transitions remain explicit and intention-revealing.

## Application

- Use cases orchestrate instead of duplicating domain rules.
- Errors preserve sentinel matching with `errors.Is` where needed.
- Context cancellation is propagated.

## HTTP

- DTOs stay separate from domain types.
- Error responses preserve stable machine-readable codes.
- Handlers do not duplicate domain validation unnecessarily.

## Persistence

- Aggregates are saved atomically.
- Driver-specific errors do not leak across the application boundary.
- Rehydration goes through domain constructors.

## Tests and Docs

- Focused tests were added or updated for behavior changes.
- Relevant docs in `docs/` were updated.
- The README still points to the correct entry points.
