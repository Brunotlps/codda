package memory

import (
	"context"
	"sort"
	"sync"

	"github.com/Brunotlps/codda/internal/application"
	"github.com/Brunotlps/codda/internal/domain"
)

// OrderRepository is an in-memory implementation of
// application.OrderRepository.
type OrderRepository struct {
	mu     sync.RWMutex
	orders map[domain.OrderID]*domain.Order
}

// NewOrderRepository returns an empty OrderRepository.
func NewOrderRepository() *OrderRepository {
	return &OrderRepository{
		orders: make(map[domain.OrderID]*domain.Order),
	}
}

// Save persists order, creating it if no order with its ID exists yet or
// updating it if one does.
func (r *OrderRepository) Save(ctx context.Context, order *domain.Order) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.orders[order.ID()] = order

	return nil
}

// FindByID retrieves the order with the given ID. It returns
// application.ErrOrderNotFound if no order with that ID exists.
func (r *OrderRepository) FindByID(ctx context.Context, id domain.OrderID) (*domain.Order, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	order, ok := r.orders[id]
	if !ok {
		return nil, application.ErrOrderNotFound
	}

	return order, nil
}

// List returns the page of orders selected by pagination from those
// matching filters, along with whether further pages are available.
// Results are ordered by createdAt descending, with ID descending as a
// tiebreaker.
func (r *OrderRepository) List(ctx context.Context, filters application.ListOrdersFilters, pagination application.Pagination) ([]*domain.Order, bool, error) {
	if err := ctx.Err(); err != nil {
		return nil, false, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	var filtered []*domain.Order
	for _, order := range r.orders {
		if matchesFilters(order, filters) {
			filtered = append(filtered, order)
		}
	}

	sort.Slice(filtered, func(i, j int) bool {
		if !filtered[i].CreatedAt().Equal(filtered[j].CreatedAt()) {
			return filtered[i].CreatedAt().After(filtered[j].CreatedAt())
		}
		// Not covered by tests: time.Now() never returns equal values across
		// consecutive calls in practice, so this tiebreaker is unreachable
		// without a way to construct two Orders sharing a createdAt.
		return filtered[i].ID() > filtered[j].ID()
	})

	start := min(pagination.Offset, len(filtered))

	end := max(start+pagination.Limit, start)
	end = min(end, len(filtered))

	hasMore := len(filtered) > end

	page := make([]*domain.Order, end-start)
	copy(page, filtered[start:end])

	return page, hasMore, nil
}

// matchesFilters reports whether order satisfies every constraint set in
// filters.
func matchesFilters(order *domain.Order, filters application.ListOrdersFilters) bool {
	if filters.Status != nil && order.Status() != *filters.Status {
		return false
	}
	if filters.CreatedFrom != nil && order.CreatedAt().Before(*filters.CreatedFrom) {
		return false
	}
	if filters.CreatedTo != nil && order.CreatedAt().After(*filters.CreatedTo) {
		return false
	}
	if filters.PriceMin != nil && order.Total().Cents() < filters.PriceMin.Cents() {
		return false
	}
	if filters.PriceMax != nil && order.Total().Cents() > filters.PriceMax.Cents() {
		return false
	}

	return true
}
