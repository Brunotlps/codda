package application

import (
	"context"
	"errors"
	"time"

	"github.com/Brunotlps/codda/internal/domain"
)

// ErrOrderNotFound is returned by OrderRepository methods when no order
// matches the requested identifier.
var ErrOrderNotFound = errors.New("order not found")

// ListOrdersFilters narrows the set of orders returned by
// OrderRepository.List. Each field is optional: a nil value means that
// field does not constrain the results.
type ListOrdersFilters struct {
	// Status, if set, restricts results to orders with this status.
	Status *domain.OrderStatus

	// CreatedFrom, if set, restricts results to orders created at or
	// after this time.
	CreatedFrom *time.Time

	// CreatedTo, if set, restricts results to orders created at or
	// before this time.
	CreatedTo *time.Time

	// PriceMin, if set, restricts results to orders whose total is at
	// least this amount.
	PriceMin *domain.Money

	// PriceMax, if set, restricts results to orders whose total is at
	// most this amount.
	PriceMax *domain.Money
}

// Pagination selects a page of a result set returned by
// OrderRepository.List. The repository does not validate these values;
// validation is the use case's responsibility.
type Pagination struct {
	// Offset is the number of matching orders to skip from the start of
	// the result set.
	Offset int

	// Limit is the maximum number of orders to return.
	Limit int
}

// OrderRepository persists and retrieves Order aggregates. It is the
// output port implemented by secondary adapters, such as a Postgres
// adapter.
type OrderRepository interface {
	// Save persists order, creating it if no order with its ID exists yet
	// or updating it if one does. The repository decides which based on
	// the order's ID.
	Save(ctx context.Context, order *domain.Order) error

	// FindByID retrieves the order with the given ID. It returns
	// ErrOrderNotFound if no order with that ID exists.
	FindByID(ctx context.Context, id domain.OrderID) (*domain.Order, error)

	// List returns the page of orders selected by pagination from those
	// matching filters, along with whether further pages are available.
	// Orders are returned by createdAt descending (most recent first),
	// with ID descending as a tiebreaker, so that pagination is stable.
	List(ctx context.Context, filters ListOrdersFilters, pagination Pagination) ([]*domain.Order, bool, error)
}
