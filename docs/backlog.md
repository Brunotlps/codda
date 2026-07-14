# Backlog

This backlog records known issues, risks, and improvement candidates found
during project review. Items are written in a form suitable for GitHub issues.

## High Priority

### 1. Static healthcheck does not verify database readiness

GitHub issue: [#1](https://github.com/Brunotlps/codda/issues/1)

`GET /health` always returns `200 OK`. Now that the service depends on
PostgreSQL, the API should expose readiness that fails when the database is not
reachable.

Status: addressed by PR [#14](https://github.com/Brunotlps/codda/pull/14).

Acceptance criteria:

- Add a readiness endpoint or enhance `/health` with dependency checks.
- Return `503` when PostgreSQL cannot be pinged.
- Cover the behavior with tests.

### 2. Add tests for `internal/config`

GitHub issue: [#2](https://github.com/Brunotlps/codda/issues/2)

`config.Load` has no test coverage, even though it validates startup-critical
environment variables.

Status: addressed by PR [#11](https://github.com/Brunotlps/codda/pull/11).

Acceptance criteria:

- Test default `HTTP_PORT`.
- Test valid boundary ports `1` and `65535`.
- Test non-numeric, zero, and out-of-range ports.
- Test missing `DATABASE_URL`.

### 3. Add an application smoke harness

GitHub issue: [#3](https://github.com/Brunotlps/codda/issues/3)

The local smoke harness starts the app and exercises the main API flow outside
`go test`.

Status: addressed by PR [#16](https://github.com/Brunotlps/codda/pull/16).

Acceptance criteria:

- Start PostgreSQL locally.
- Start the service.
- Wait for `/ready`.
- Create, read, pay, ship, and list an order.
- Clean up the service process.

## Medium Priority

### 4. Relative migration path depends on working directory

GitHub issue: [#4](https://github.com/Brunotlps/codda/issues/4)

The migration source is `file://internal/adapters/postgres/migrations`. A
built binary fails if it is executed from a directory where that relative path
does not exist.

Status: addressed by PR [#16](https://github.com/Brunotlps/codda/pull/16).

Acceptance criteria:

- Embed migrations or otherwise resolve them independently of the working
  directory.
- Preserve automatic startup migrations.

### 5. List query parameters are silently permissive

GitHub issue: [#5](https://github.com/Brunotlps/codda/issues/5)

Invalid values such as `status=bogus`, `limit=abc`, or `offset=abc` do not
return `400`. They are treated as empty/default filters.

Status: addressed by PR [#13](https://github.com/Brunotlps/codda/pull/13).

Acceptance criteria:

- Decide whether invalid query parameters should be rejected.
- If rejected, return `400` with `validation_error`.
- Add handler tests for invalid status, limit, and offset.

### 6. HTTP validation error mapping can drift

GitHub issue: [#6](https://github.com/Brunotlps/codda/issues/6)

`httpStatusForError` manually lists domain sentinel errors that map to
`400 validation_error`. New domain validation errors can accidentally fall
through to `500 internal_error`.

Status: addressed by PR [#12](https://github.com/Brunotlps/codda/pull/12).

Acceptance criteria:

- Add tests covering all domain validation sentinels.
- Consider grouping validation errors through a helper or typed error.

### 7. `ListOrders` loads full aggregates for summary responses

GitHub issue: [#7](https://github.com/Brunotlps/codda/issues/7)

The repository port returns `[]*domain.Order`, so the Postgres adapter loads
items even when the use case only returns `OrderResume`.

Acceptance criteria:

- Evaluate whether listing should use a dedicated read port.
- If changed, keep detail reads returning full aggregates.
- Preserve current API behavior.

## Low Priority

### 8. `Money.Multiply` can overflow `int64`

GitHub issue: [#8](https://github.com/Brunotlps/codda/issues/8)

Large prices or quantities can overflow when calculating totals.

Status: addressed by PR [#15](https://github.com/Brunotlps/codda/pull/15).

Acceptance criteria:

- Decide whether to enforce maximum money/quantity bounds or checked
  arithmetic.
- Add tests for overflow behavior.

### 9. `ErrDuplicateProductInOrder` is defined but unused

GitHub issue: [#9](https://github.com/Brunotlps/codda/issues/9)

Duplicate product validation exists as a sentinel error but is not enforced in
hydration or row mapping.

Status: addressed by PR [#15](https://github.com/Brunotlps/codda/pull/15).

Acceptance criteria:

- Decide whether hydration should reject duplicate product IDs.
- If yes, add validation and tests.

### 10. Startup failure logs look like graceful shutdown

GitHub issue: [#10](https://github.com/Brunotlps/codda/issues/10)

If `server.Start` fails, `main` cancels the signal context and logs the same
shutdown path as a normal signal-triggered shutdown.

Status: addressed by PR [#16](https://github.com/Brunotlps/codda/pull/16).

Acceptance criteria:

- Make startup failure logs distinguishable from graceful shutdown logs.
- Consider returning a non-zero exit path for startup failures after wiring.
