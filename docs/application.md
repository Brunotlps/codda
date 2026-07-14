# Application Layer

The application package contains use cases and ports. It orchestrates work but
does not own domain rules.

## Output Ports

`OrderRepository` is the persistence port:

```go
type OrderRepository interface {
    Save(ctx context.Context, order *domain.Order) error
    FindByID(ctx context.Context, id domain.OrderID) (*domain.Order, error)
}
```

`OrderReadRepository` is the read-model port used by list views:

```go
type OrderReadRepository interface {
    ListResumes(ctx context.Context, filters ListOrdersFilters, pagination Pagination) ([]OrderResume, bool, error)
}
```

Implementations:

- `internal/adapters/postgres`: production persistence.
- `internal/adapters/memory`: in-memory adapter for tests.

`ErrOrderNotFound` belongs to the application layer because it describes a
lookup result, not a domain invariant.

## Use Cases

Implemented use cases:

- `CreateOrderUseCase`
- `FindOrderByIDUseCase`
- `ListOrdersUseCase`
- `MarkOrderAsPaidUseCase`
- `MarkOrderAsCancelledUseCase`
- `MarkOrderAsShippedUseCase`

Each status transition has a dedicated use case. There is no generic
`UpdateOrderStatus` use case.

## Create Order

`CreateOrderUseCase` accepts raw input:

- product ID
- product name
- price in cents
- quantity

It builds domain value objects, creates an order, saves it, and returns the
new order ID. Item-level errors are wrapped with the item index so callers can
identify which input row failed while still using `errors.Is`.

## Find Order

`FindOrderByIDUseCase` is a direct read use case. It delegates to the
repository and returns the full aggregate.

## List Orders

`ListOrdersUseCase` returns `OrderResume`, a read model with:

- ID
- status
- total
- creation time

Pagination is sanitized in the use case:

- `Limit <= 0` becomes `20`.
- `Limit > 100` becomes `100`.
- `Offset < 0` becomes `0`.

Filters are passed to the repository:

- status
- creation date range
- price range

The use case depends on `OrderReadRepository`, not the aggregate persistence
port. Detail reads still return full `Order` aggregates through
`FindOrderByIDUseCase`.

## Transition Use Cases

Each transition follows the same shape:

1. Load the order.
2. Call the domain method.
3. Save the order.

The use case never checks `order.Status()` to decide whether the transition is
valid. That rule belongs to the aggregate.
