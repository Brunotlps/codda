package http

import "time"

// CreateOrderRequest is the request body for POST /orders.
type CreateOrderRequest struct {
	Items []CreateOrderItemRequest `json:"items"`
}

// CreateOrderItemRequest is a single item within a CreateOrderRequest.
type CreateOrderItemRequest struct {
	ProductID   string `json:"product_id"`
	ProductName string `json:"product_name"`
	PriceCents  int64  `json:"price_cents"`
	Quantity    int    `json:"quantity"`
}

// CreateOrderResponse is the response body for a successful POST /orders.
type CreateOrderResponse struct {
	ID string `json:"id"`
}

// OrderResponse is the response body for GET /orders/{id}, describing an
// order in full, including its items.
type OrderResponse struct {
	ID         string              `json:"id"`
	Status     string              `json:"status"`
	Items      []OrderItemResponse `json:"items"`
	TotalCents int64               `json:"total_cents"`
	CreatedAt  time.Time           `json:"created_at"`
}

// OrderItemResponse is a single item within an OrderResponse.
type OrderItemResponse struct {
	ProductID   string `json:"product_id"`
	ProductName string `json:"product_name"`
	PriceCents  int64  `json:"price_cents"`
	Quantity    int    `json:"quantity"`
}

// OrderResumeResponse is a lightweight projection of an order, used in
// listings where item-level detail isn't needed.
type OrderResumeResponse struct {
	ID         string    `json:"id"`
	Status     string    `json:"status"`
	TotalCents int64     `json:"total_cents"`
	CreatedAt  time.Time `json:"created_at"`
}

// ListOrdersResponse is the response body for GET /orders.
type ListOrdersResponse struct {
	Orders  []OrderResumeResponse `json:"orders"`
	HasMore bool                  `json:"has_more"`
}

// ErrorResponse is the response body for any failed request.
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail describes a single error within an ErrorResponse.
type ErrorDetail struct {
	// Code is a stable, machine-readable identifier for the error, intended
	// for clients to branch on (e.g. "order_not_found"). It is part of the
	// public API contract and must remain stable across versions.
	Code string `json:"code"`

	// Message is a human-readable description of the error, intended for
	// logs and debugging.
	Message string `json:"message"`
}
