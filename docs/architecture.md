# Architecture

Codda follows hexagonal architecture, also known as ports and adapters. The
business core is split into two packages:

- `internal/domain`: pure business model and invariants.
- `internal/application`: use cases and ports.

Adapters live outside the core:

- `internal/adapters/http`: primary adapter driven by HTTP requests.
- `internal/adapters/postgres`: secondary adapter implementing persistence.
- `internal/adapters/memory`: secondary adapter used by tests.

The composition root is `cmd/orderservice/main.go`.

## Dependency Rule

Dependencies point inward:

```text
HTTP adapter  -> application -> domain
Postgres      -> application -> domain
Memory        -> application -> domain
main          -> all packages for wiring only
```

The domain package does not import infrastructure packages. The application
package imports only the domain and the standard library. Adapters may import
the application and domain packages because they translate external concerns
into core operations.

## Runtime Flow

Typical create-order request:

```text
POST /orders
  -> http.Handler.CreateOrder
  -> application.CreateOrderUseCase.Execute
  -> domain.NewMoney / NewOrderItem / NewOrder
  -> application.OrderRepository.Save
  -> postgres.OrderRepository.Save
```

Typical status transition:

```text
POST /orders/{id}/pay
  -> http.Handler.MarkOrderAsPaid
  -> application.MarkOrderAsPaidUseCase.Execute
  -> repository.FindByID
  -> order.MarkAsPaid
  -> repository.Save
```

The use case does not inspect status rules directly. Transition validity stays
inside the `Order` aggregate and `OrderStatus` state machine.

## Composition Root

`cmd/orderservice/main.go` is the only package that wires the whole system:

1. Load configuration.
2. Create and ping a PostgreSQL pool.
3. Apply SQL migrations.
4. Construct the PostgreSQL repository.
5. Construct all use cases.
6. Construct the HTTP handler, router, and server.
7. Run with graceful shutdown.

No business rule should be added to `main.go`.

## Design Constraints

- Ports are owned by the consumer. `OrderRepository` lives in
  `internal/application`.
- The repository persists complete `Order` aggregates. There is no
  `OrderItemRepository`.
- Domain objects are not JSON DTOs and have no persistence tags.
- Each status transition has a dedicated use case and HTTP route.
- Commands that only mutate state return no representation; clients can issue a
  separate query if they need the updated state.
