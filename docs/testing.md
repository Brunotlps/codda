# Testing

The test strategy mirrors the architecture:

- domain tests are pure unit tests;
- application tests use the in-memory repository;
- HTTP tests wire the real router and handler against the in-memory repository;
- Postgres tests run against a real PostgreSQL container through
  Testcontainers.

## Commands

Fast tests without Docker:

```sh
go test ./internal/domain/... ./internal/application/... ./internal/adapters/memory/... ./internal/adapters/http/... ./internal/config/...
```

Full suite:

```sh
go test ./...
```

Coverage:

```sh
go test -cover ./...
```

## Current Coverage Snapshot

Last measured locally:

| Package | Coverage |
| --- | ---: |
| `internal/domain` | 100.0% |
| `internal/application` | 100.0% |
| `internal/adapters/memory` | 97.8% |
| `internal/adapters/postgres` | 78.4% |
| `internal/adapters/http` | 66.7% |
| `internal/config` | 0.0% |
| `cmd/orderservice` | 0.0% |

## Sandbox Notes

Some environments restrict local sockets or Docker access. In those
environments:

- `httptest.NewServer` can fail if opening a local listener is not permitted;
- Testcontainers fails if it cannot access the Docker socket;
- coverage tooling can fail if the Go toolchain cannot write required cache
  files.

Run the full suite outside the restricted sandbox when validating release
readiness.

## Test Design

The in-memory repository is a first-class adapter, not a mock. It keeps use
case tests fast while preserving the same application port used by production.

The Postgres package uses one test container per package test run and truncates
tables between tests.
