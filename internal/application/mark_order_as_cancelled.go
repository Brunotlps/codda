package application

import (
	"context"

	"github.com/Brunotlps/codda/internal/domain"
)

// MarkOrderAsCancelledUseCase transitions an order to the cancelled status.
type MarkOrderAsCancelledUseCase struct {
	repo OrderRepository
}

// NewMarkOrderAsCancelledUseCase creates a MarkOrderAsCancelledUseCase
// backed by repo.
func NewMarkOrderAsCancelledUseCase(repo OrderRepository) *MarkOrderAsCancelledUseCase {
	return &MarkOrderAsCancelledUseCase{repo: repo}
}

// Execute loads the order with the given ID, cancels it, and persists the
// result. It returns ErrOrderNotFound if no order with that ID exists, or
// ErrInvalidStatusTransition if the order's current status does not allow
// the transition.
func (uc *MarkOrderAsCancelledUseCase) Execute(ctx context.Context, id domain.OrderID) error {
	order, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if err := order.Cancel(); err != nil {
		return err
	}

	return uc.repo.Save(ctx, order)
}
