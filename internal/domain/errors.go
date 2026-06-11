package domain

import "errors"

// Sentinel errors describing violations of Money, OrderItem, and Order domain invariants.
var (
	// ErrNegativeMoney is returned when a Money value is constructed with a
	// negative amount of cents.
	ErrNegativeMoney = errors.New("money amount must not be negative")

	// ErrEmptyProductID is returned when an order item's product ID is empty.
	ErrEmptyProductID = errors.New("product id must not be empty")

	// ErrEmptyProductName is returned when an order item's product name is empty.
	ErrEmptyProductName = errors.New("product name must not be empty")

	// ErrProductNameTooLong is returned when an order item's product name
	// exceeds the maximum allowed length.
	ErrProductNameTooLong = errors.New("product name exceeds maximum length")

	// ErrInvalidQuantity is returned when an order item's quantity is less than 1.
	ErrInvalidQuantity = errors.New("quantity must be greater than or equal to 1")

	// ErrInvalidPrice is returned when an order item's price is not greater than zero.
	ErrInvalidPrice = errors.New("price must be greater than zero")

	// ErrEmptyOrderID is returned when an order's ID is empty.
	ErrEmptyOrderID = errors.New("order id must not be empty")

	// ErrOrderRequiresItems is returned when an order has no items.
	ErrOrderRequiresItems = errors.New("order requires at least one item")

	// ErrInvalidStatus is returned when an order is assigned a status outside
	// the set of known valid statuses.
	ErrInvalidStatus = errors.New("invalid order status")

	// ErrInvalidStatusTransition is returned when an order attempts to move
	// from its current status to a target status not allowed by the
	// transition rules in CanTransitionTo.
	ErrInvalidStatusTransition = errors.New("invalid order status transition")

	// ErrDuplicateProductInOrder is returned when an order ends up with more
	// than one item sharing the same product ID. It guards order construction
	// and validation paths that do not go through item-merging logic.
	ErrDuplicateProductInOrder = errors.New("order contains duplicate product id")
)
