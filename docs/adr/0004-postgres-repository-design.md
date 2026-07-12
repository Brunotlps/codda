# ADR 0004: Persist Orders as Complete Aggregates in PostgreSQL

## Status

Accepted

## Context

`Order` is the aggregate root and owns its items. Persistence must keep an
order and its items consistent.

## Decision

Implement `OrderRepository` in PostgreSQL with:

- one `orders` table;
- one `order_items` table;
- transactional `Save`;
- UPSERT for the order row;
- delete-and-reinsert strategy for items;
- rehydration through domain constructors.

## Consequences

- Save is simple and atomic.
- Item ordering is preserved by an explicit `position` column.
- The adapter loads full aggregates for list operations because the current
  port returns domain orders.
