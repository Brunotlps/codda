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
shutdown.

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

`GET /health` returns static `200 OK`. It does not currently check database
readiness.

For production-like deployments, split health into:

- liveness: process is alive;
- readiness: dependencies such as PostgreSQL are reachable.

## Local Smoke Harness

Run:

```sh
tools/harness/smoke.sh
```

The script starts PostgreSQL with Docker Compose, starts the service, waits for
`/health`, creates an order, reads it, pays it, ships it, lists orders, and
then stops the service process it started.
