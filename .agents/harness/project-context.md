# Codda Project Context

Codda is a Go order service built with hexagonal architecture.

## Source of Truth

1. Production code and tests.
2. Official docs in `docs/`.
3. Historical notes in `local-docs/`, when available.

If docs and code disagree, trust the code and update docs as part of the task.

## Package Map

- `cmd/orderservice`: composition root.
- `internal/domain`: aggregate, value objects, state machine, domain errors.
- `internal/application`: use cases and ports.
- `internal/adapters/http`: HTTP DTOs, handlers, router, server wrapper.
- `internal/adapters/postgres`: PostgreSQL repository and migrations.
- `internal/adapters/memory`: in-memory repository for tests.
- `internal/config`: environment loading.
- `tools/harness`: local smoke harness.

## Architectural Rules

- Domain must not import adapters, HTTP, PostgreSQL, or application.
- Application must not import adapters or infrastructure.
- Ports belong to application, not adapters.
- Adapters may import application and domain.
- `main.go` wires components and should not contain business rules.

## Main Business Flow

- Create order: HTTP request -> create use case -> domain constructors ->
  repository save.
- Read order: HTTP request -> find use case -> repository find.
- List orders: HTTP request -> list use case -> repository list -> resume
  projection.
- Transitions: HTTP action route -> transition use case -> domain transition
  method -> repository save.

## Validation Commands

Fast tests:

```sh
go test ./internal/domain/... ./internal/application/... ./internal/adapters/memory/... ./internal/adapters/http/... ./internal/config/...
```

Full tests:

```sh
go test ./...
```

Smoke harness:

```sh
tools/harness/smoke.sh
```
