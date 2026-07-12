# Local Harness

`smoke.sh` is a simple end-to-end harness for local development.

It:

1. starts PostgreSQL through Docker Compose;
2. starts `go run ./cmd/orderservice/`;
3. waits for `GET /health`;
4. creates an order;
5. reads the order;
6. pays the order;
7. ships the order;
8. lists orders;
9. stops the service process it started.

Run from the repository root:

```sh
tools/harness/smoke.sh
```

Optional environment variables:

| Variable | Default |
| --- | --- |
| `BASE_URL` | `http://localhost:8080` |
| `DATABASE_URL` | `postgres://codda:codda@localhost:5433/codda?sslmode=disable` |
| `HTTP_PORT` | `8080` |
| `LOG_FILE` | `/tmp/codda-smoke.log` |
