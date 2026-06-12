package domain

import (
	"time"

	"github.com/google/uuid"
)

// OrderID uniquely identifies an Order.
type OrderID string

// Order is the aggregate root representing a customer order: its items,
// lifecycle status, and identifying metadata.
type Order struct {
	id        OrderID
	items     []OrderItem
	status    OrderStatus
	createdAt time.Time
}

// NewOrder creates an Order from items. Items sharing the same product ID
// are merged into a single OrderItem with their quantities summed,
// preserving the order in which each product ID first appears. NewOrder
// returns ErrOrderRequiresItems if items is empty. The resulting Order is
// assigned a generated ID, status StatusPending, and the current time as
// createdAt.
func NewOrder(items []OrderItem) (*Order, error) {
	if len(items) == 0 {
		return nil, ErrOrderRequiresItems
	}

	merged, err := mergeItems(items)
	if err != nil {
		return nil, err
	}

	return &Order{
		id:        OrderID(uuid.NewString()),
		items:     merged,
		status:    StatusPending,
		createdAt: time.Now(),
	}, nil
}

// mergeItems combines items that share the same product ID into a single
// OrderItem with their quantities summed, preserving the order in which
// each product ID first appears. The returned slice does not alias items.
func mergeItems(items []OrderItem) ([]OrderItem, error) {
	var merged []OrderItem

	for _, item := range items {
		foundIdx := -1
		for i, existing := range merged {
			if existing.ProductID() == item.ProductID() {
				foundIdx = i
				break
			}
		}

		if foundIdx >= 0 {
			updated, err := merged[foundIdx].WithQuantity(merged[foundIdx].Quantity() + item.Quantity())
			if err != nil {
				return nil, err
			}
			merged[foundIdx] = updated
		} else {
			merged = append(merged, item)
		}
	}

	return merged, nil
}

// ID returns the order's unique identifier.
func (o *Order) ID() OrderID {
	return o.id
}

// Status returns the order's current status.
func (o *Order) Status() OrderStatus {
	return o.status
}

// Items returns a copy of the order's items. Mutating the returned slice
// does not affect the order.
func (o *Order) Items() []OrderItem {
	items := make([]OrderItem, len(o.items))
	copy(items, o.items)

	return items
}

// CreatedAt returns the time at which the order was created.
func (o *Order) CreatedAt() time.Time {
	return o.createdAt
}

// Total returns the sum of the totals of all items in the order.
func (o *Order) Total() Money {
	var total Money

	for _, item := range o.items {
		total = total.Add(item.Total())
	}

	return total
}

// MarkAsPaid transitions the order to StatusPaid. It returns
// ErrInvalidStatusTransition if the order's current status does not allow
// that transition.
func (o *Order) MarkAsPaid() error {
	if !o.status.CanTransitionTo(StatusPaid) {
		return ErrInvalidStatusTransition
	}

	o.status = StatusPaid

	return nil
}

// Cancel transitions the order to StatusCancelled. It returns
// ErrInvalidStatusTransition if the order's current status does not allow
// that transition.
func (o *Order) Cancel() error {
	if !o.status.CanTransitionTo(StatusCancelled) {
		return ErrInvalidStatusTransition
	}

	o.status = StatusCancelled

	return nil
}

// Ship transitions the order to StatusShipped. It returns
// ErrInvalidStatusTransition if the order's current status does not allow
// that transition.
func (o *Order) Ship() error {
	if !o.status.CanTransitionTo(StatusShipped) {
		return ErrInvalidStatusTransition
	}

	o.status = StatusShipped

	return nil
}
