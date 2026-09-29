package domain

import (
	"math"
	"unicode/utf8"
)

// MaxProductNameLength is the maximum number of characters allowed in an
// OrderItem's product name. 255 aligns with the conventional VARCHAR(255)
// limit and comfortably accommodates real product names.
const MaxProductNameLength = 255

// MaxOrderItemQuantity is the largest quantity representable by the
// PostgreSQL INTEGER column used for persisted order items.
const MaxOrderItemQuantity = math.MaxInt32

// OrderItem represents a single line item within an Order: a product,
// its unit price, and the quantity ordered.
type OrderItem struct {
	productID   string
	productName string
	price       Money
	quantity    int
}

// NewOrderItem creates an OrderItem after validating its invariants:
// productID and productName must not be empty, productName must not exceed
// MaxProductNameLength characters, price must be greater than zero, and
// quantity must be between 1 and MaxOrderItemQuantity. On the first violated
// invariant it returns the zero value and the corresponding sentinel error.
func NewOrderItem(productID, productName string, price Money, quantity int) (OrderItem, error) {
	if productID == "" {
		return OrderItem{}, ErrEmptyProductID
	}
	if productName == "" {
		return OrderItem{}, ErrEmptyProductName
	}
	if utf8.RuneCountInString(productName) > MaxProductNameLength {
		return OrderItem{}, ErrProductNameTooLong
	}
	if price.Cents() <= 0 {
		return OrderItem{}, ErrInvalidPrice
	}
	if quantity < 1 {
		return OrderItem{}, ErrInvalidQuantity
	}
	if quantity > MaxOrderItemQuantity {
		return OrderItem{}, ErrQuantityTooLarge
	}
	if _, err := price.CheckedMultiply(quantity); err != nil {
		return OrderItem{}, err
	}

	return OrderItem{
		productID:   productID,
		productName: productName,
		price:       price,
		quantity:    quantity,
	}, nil
}

// ProductID returns the item's product ID.
func (i OrderItem) ProductID() string {
	return i.productID
}

// ProductName returns the item's product name.
func (i OrderItem) ProductName() string {
	return i.productName
}

// Price returns the item's unit price.
func (i OrderItem) Price() Money {
	return i.price
}

// Quantity returns the item's quantity.
func (i OrderItem) Quantity() int {
	return i.quantity
}

// Total returns the item's price multiplied by its quantity.
func (i OrderItem) Total() Money {
	return i.price.Multiply(i.quantity)
}

// WithQuantity returns a copy of i with its quantity replaced by quantity.
// It returns ErrInvalidQuantity if quantity is less than 1, or
// ErrQuantityTooLarge if it exceeds MaxOrderItemQuantity, or ErrMoneyOverflow
// if the resulting line total cannot be represented.
func (i OrderItem) WithQuantity(quantity int) (OrderItem, error) {
	if quantity < 1 {
		return OrderItem{}, ErrInvalidQuantity
	}
	if quantity > MaxOrderItemQuantity {
		return OrderItem{}, ErrQuantityTooLarge
	}
	if _, err := i.price.CheckedMultiply(quantity); err != nil {
		return OrderItem{}, err
	}

	i.quantity = quantity

	return i, nil
}
