# Persistence

The production repository is `internal/adapters/postgres.OrderRepository`.
It implements the application `OrderRepository` port with `pgxpool`.

## Schema

Migrations live in `internal/adapters/postgres/migrations`.

Tables:

```sql
CREATE TABLE orders (
    id          UUID PRIMARY KEY,
    status      TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL
);

CREATE TABLE order_items (
    order_id      UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    position      INTEGER NOT NULL,
    product_id    TEXT NOT NULL,
    product_name  TEXT NOT NULL,
    price_cents   BIGINT NOT NULL,
    quantity      INTEGER NOT NULL,
    PRIMARY KEY (order_id, position)
);
```

Indexes:

- `idx_orders_created_at`
- `idx_orders_status`

## Save

`Save` runs in a transaction:

1. UPSERT the `orders` row by ID.
2. Delete existing `order_items` for the order.
3. Insert all current items with stable `position` values.
4. Commit.

The repository does not distinguish create from update at the port boundary.
The adapter decides through PostgreSQL `ON CONFLICT`.

`created_at` is inserted on first save and not updated on conflict.

## FindByID

`FindByID`:

1. Queries the `orders` row.
2. Converts `pgx.ErrNoRows` to `application.ErrOrderNotFound`.
3. Queries `order_items` ordered by position.
4. Reconstructs the domain aggregate through `rowsToOrder`.

Infrastructure-specific not-found errors do not leak into use cases.

## ListResumes

`ListResumes` builds a dynamic query based on optional filters:

- status
- created from
- created to
- minimum total
- maximum total

The query uses a derived totals subquery because the domain does not store
`total_cents` as a column. The same total is returned in the `OrderResume`
projection.

Pagination uses `limit + 1` to detect whether another page exists without
running a separate count query.

List responses use the application `OrderReadRepository` port and return
`OrderResume` values directly. Detail reads still use `FindByID` and
reconstruct full aggregates with item rows.

## Migrations

`cmd/orderservice` applies migrations during startup with `golang-migrate`.

Migration files are embedded in the binary from
`internal/adapters/postgres/migrations` and loaded through the `iofs` source
driver. Startup migrations do not depend on the process working directory.
