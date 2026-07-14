# HTTP API

The HTTP adapter exposes the order use cases through JSON endpoints.

Base URL in local development:

```text
http://localhost:8080
```

## Health

```http
GET /health
```

Returns `200 OK`. This is currently a static liveness endpoint and does not
check PostgreSQL readiness.

## Create Order

```http
POST /orders
Content-Type: application/json
```

Request:

```json
{
  "items": [
    {
      "product_id": "p1",
      "product_name": "Widget",
      "price_cents": 1999,
      "quantity": 2
    }
  ]
}
```

Response `201 Created`:

```json
{
  "id": "order-id"
}
```

## Get Order

```http
GET /orders/{id}
```

Response `200 OK`:

```json
{
  "id": "order-id",
  "status": "pending",
  "items": [
    {
      "product_id": "p1",
      "product_name": "Widget",
      "price_cents": 1999,
      "quantity": 2
    }
  ],
  "total_cents": 3998,
  "created_at": "2026-07-12T10:00:00Z"
}
```

## List Orders

```http
GET /orders?limit=10&offset=0&status=paid
```

Supported query parameters:

- `status`
- `limit`
- `offset`

Response `200 OK`:

```json
{
  "orders": [
    {
      "id": "order-id",
      "status": "paid",
      "total_cents": 3998,
      "created_at": "2026-07-12T10:00:00Z"
    }
  ],
  "has_more": false
}
```

## Status Transitions

```http
POST /orders/{id}/pay
POST /orders/{id}/cancel
POST /orders/{id}/ship
```

Successful transition responses return `204 No Content`.

## Error Response

All handled errors use this shape:

```json
{
  "error": {
    "code": "validation_error",
    "message": "price must be greater than zero"
  }
}
```

Stable error codes:

| Code | HTTP status | Meaning |
| --- | ---: | --- |
| `validation_error` | 400 | Invalid request body, missing ID, or domain validation failure |
| `order_not_found` | 404 | No order exists for the requested ID |
| `invalid_status_transition` | 409 | The requested lifecycle transition is not allowed |
| `service_unavailable` | 503 | Request deadline exceeded |
| `internal_error` | 500 | Unhandled server error |

Domain validation errors are classified in the domain package and mapped
consistently to `400 validation_error` by the HTTP adapter.

If the request context is canceled, the adapter writes no response because the
client is assumed to have disconnected.
