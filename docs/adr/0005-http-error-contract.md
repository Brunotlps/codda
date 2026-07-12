# ADR 0005: Use Stable HTTP Error Codes

## Status

Accepted

## Context

HTTP clients need machine-readable errors while the domain and application
layers should remain transport-agnostic.

## Decision

The HTTP adapter maps known domain/application errors into this response shape:

```json
{
  "error": {
    "code": "validation_error",
    "message": "human-readable message"
  }
}
```

Known codes:

- `validation_error`
- `order_not_found`
- `invalid_status_transition`
- `service_unavailable`
- `internal_error`

## Consequences

- Clients can branch on stable error codes.
- The mapping is centralized in the HTTP adapter.
- New domain validation errors must be added to the mapping or covered by a
  more general validation-error strategy.
