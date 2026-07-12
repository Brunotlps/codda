# Configuration

Configuration is loaded by `internal/config.Load`.

## Environment Variables

| Variable | Required | Default | Description |
| --- | --- | --- | --- |
| `DATABASE_URL` | Yes | none | PostgreSQL connection string |
| `HTTP_PORT` | No | `8080` | HTTP server port, from 1 to 65535 |

Example:

```sh
export DATABASE_URL="postgres://codda:codda@localhost:5433/codda?sslmode=disable"
export HTTP_PORT=8080
```

`HTTP_PORT` is converted into the `http.Server` address format by prefixing it
with `:`.

## Validation

`Load` rejects:

- non-numeric `HTTP_PORT` values;
- port values outside `1..65535`;
- missing `DATABASE_URL`.

`DATABASE_URL` is not parsed by the config package. The PostgreSQL driver and
migration library validate it when they connect.

## Current Gaps

- `internal/config` has no test file.
- Pool sizing and timeout configuration are not exposed through config.
- `DATABASE_URL` has no early format validation.
