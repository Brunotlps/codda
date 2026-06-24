package application

import (
	"context"

	"github.com/Brunotlps/codda/internal/domain"
)

// FindOrderByIDUseCase retrieves a single order by its ID.
type FindOrderByIDUseCase struct {
	repo OrderRepository
}

// NewFindOrderByIDUseCase creates a FindOrderByIDUseCase backed by repo.
func NewFindOrderByIDUseCase(repo OrderRepository) *FindOrderByIDUseCase {
	return &FindOrderByIDUseCase{repo: repo}
}

// Execute retrieves the order with the given ID. It returns
// ErrOrderNotFound if no order with that ID exists.
func (uc *FindOrderByIDUseCase) Execute(ctx context.Context, id domain.OrderID) (*domain.Order, error) {
	return uc.repo.FindByID(ctx, id)
}
