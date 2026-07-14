# Operations

## Local Startup

Start PostgreSQL:

```sh
docker compose up -d
```

Run the service:

```sh
export DATABASE_URL="postgres://codda:codda@localhost:5433/codda?sslmode=disable"
go run ./cmd/orderservice/
```

The service logs database connection, migration progress, server startup, and
shutdown. Startup failures from `server.Start` are logged separately from
signal-triggered graceful shutdown and return a non-zero process exit.

## Database Lifecycle

At startup, the service:

1. Creates a `pgxpool.Pool`.
2. Pings the database with a five-second timeout.
3. Applies pending migrations.
4. Starts accepting HTTP requests.

If pool creation, ping, or migration fails, startup fails fast.

## Shutdown

The process listens for `SIGINT` and `SIGTERM`. On shutdown it:

1. Stops accepting new HTTP requests.
2. Waits up to ten seconds for in-flight requests.
3. Closes the PostgreSQL pool.

## Healthcheck

`GET /health` returns static `200 OK` when the process is alive. It is a
liveness endpoint and does not check database readiness.

`GET /ready` checks required dependencies. In production wiring it pings the
PostgreSQL pool and returns:

- `200 OK` when PostgreSQL is reachable;
- `503 service_unavailable` when PostgreSQL cannot be pinged.

## Local Smoke Harness

Run:

```sh
tools/harness/smoke.sh
```

The script starts PostgreSQL with Docker Compose, starts the service, waits for
`/ready`, checks `/health`, creates an order, reads it, pays it, ships it,
lists orders, and then stops the service process it started.
