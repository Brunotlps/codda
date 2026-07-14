# Domain Model

The domain package models one bounded context: order management. It has no
database, HTTP, or framework dependency.

## Aggregate

`Order` is the aggregate root:

```go
type Order struct {
    id        OrderID
    items     []OrderItem
    status    OrderStatus
    createdAt time.Time
}
```

Fields are private. State changes only through domain methods.

## Value Objects

`Money` stores amounts in cents. The zero value is valid and represents
`R$ 0,00`.

`OrderItem` stores a product snapshot inside an order:

- product ID
- product name
- unit price
- quantity

`OrderItem` has no identity outside its order.

## Invariants

The domain enforces these invariants:

- An order requires at least one item.
- Product ID must not be empty.
- Product name must not be empty.
- Product name must not exceed 255 characters.
- Item price must be greater than zero.
- Item quantity must be at least one.
- Money must not be negative.
- Item totals and order totals must fit in `int64` cents.
- Hydrated orders must have a non-empty ID.
- Hydrated orders must have a valid status.
- Hydrated orders must not contain duplicate product IDs.
- Status transitions must follow the state machine.

Invariants live in constructors and methods that mutate state. Use cases and
HTTP handlers do not duplicate these rules.

Domain validation sentinels are classified through `IsValidationError`. HTTP
and other adapters should use that helper instead of maintaining their own
lists of validation errors.

## Order Creation

`NewOrder(items)`:

- rejects empty item lists;
- merges duplicate product IDs by summing quantities;
- preserves the first-seen product order;
- generates the order ID in the domain;
- sets status to `pending`;
- sets `createdAt` to the current time.

The domain owns identity generation. The database does not create order IDs.

## Hydration

`HydrateOrder(id, items, status, createdAt)` rebuilds an existing order from
persistence. It does not generate a new ID, does not merge items, and does not
replace `createdAt`. Persisted duplicate product IDs are rejected instead of
merged so data corruption remains visible.

Persistence adapters must rebuild `Money` and `OrderItem` through domain
constructors before hydration so corrupted database rows surface as domain
errors.

## State Machine

Valid statuses:

- `pending`
- `paid`
- `shipped`
- `cancelled`

Valid transitions:

```text
pending -> paid
pending -> cancelled
paid    -> shipped
paid    -> cancelled
```

`shipped` and `cancelled` are terminal.

Methods:

- `MarkAsPaid()`
- `Cancel()`
- `Ship()`

There is no public status setter.

## Money Arithmetic

`Money.CheckedAdd` and `Money.CheckedMultiply` provide overflow-aware
operations. Constructors and hydration paths use these checks so persisted and
new aggregates cannot represent totals that overflow `int64` cents.
