package application

import (
	"context"

	"github.com/Brunotlps/codda/internal/domain"
)

// MarkOrderAsPaidUseCase transitions an order to the paid status.
type MarkOrderAsPaidUseCase struct {
	repo OrderRepository
}

// NewMarkOrderAsPaidUseCase creates a MarkOrderAsPaidUseCase backed by repo.
func NewMarkOrderAsPaidUseCase(repo OrderRepository) *MarkOrderAsPaidUseCase {
	return &MarkOrderAsPaidUseCase{repo: repo}
}

// Execute loads the order with the given ID, marks it as paid, and
// persists the result. It returns ErrOrderNotFound if no order with that ID
// exists, or ErrInvalidStatusTransition if the order's current status does
// not allow the transition.
func (uc *MarkOrderAsPaidUseCase) Execute(ctx context.Context, id domain.OrderID) error {
	order, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if err := order.MarkAsPaid(); err != nil {
		return err
	}

	return uc.repo.Save(ctx, order)
}
