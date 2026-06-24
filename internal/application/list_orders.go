package application

import (
	"context"
	"time"

	"github.com/Brunotlps/codda/internal/domain"
)

const (
	defaultLimit = 20
	maxLimit     = 100
)

// OrderResume is a lightweight projection of an Order, carrying only the
// fields needed to render a list of orders.
type OrderResume struct {
	ID        domain.OrderID
	Status    domain.OrderStatus
	Total     domain.Money
	CreatedAt time.Time
}

// ListOrdersUseCase lists orders as OrderResume projections.
type ListOrdersUseCase struct {
	repo OrderRepository
}

// NewListOrdersUseCase creates a ListOrdersUseCase backed by repo.
func NewListOrdersUseCase(repo OrderRepository) *ListOrdersUseCase {
	return &ListOrdersUseCase{repo: repo}
}

// Execute returns the page of orders selected by pagination from those
// matching filters, as OrderResume projections, along with whether further
// pages are available. Pagination is adjusted before use: Limit defaults to
// defaultLimit when zero or negative, and is capped at maxLimit; Offset is
// clamped to 0 when negative.
func (uc *ListOrdersUseCase) Execute(ctx context.Context, filters ListOrdersFilters, pagination Pagination) ([]OrderResume, bool, error) {
	if pagination.Limit <= 0 {
		pagination.Limit = defaultLimit
	}
	if pagination.Limit > maxLimit {
		pagination.Limit = maxLimit
	}
	if pagination.Offset < 0 {
		pagination.Offset = 0
	}

	orders, hasMore, err := uc.repo.List(ctx, filters, pagination)
	if err != nil {
		return nil, false, err
	}

	resumes := make([]OrderResume, len(orders))
	for i, o := range orders {
		resumes[i] = toResume(o)
	}

	return resumes, hasMore, nil
}

// toResume projects o into an OrderResume.
func toResume(o *domain.Order) OrderResume {
	return OrderResume{
		ID:        o.ID(),
		Status:    o.Status(),
		Total:     o.Total(),
		CreatedAt: o.CreatedAt(),
	}
}
