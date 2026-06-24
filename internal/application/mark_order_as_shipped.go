package application

import (
	"context"

	"github.com/Brunotlps/codda/internal/domain"
)

// MarkOrderAsShippedUseCase transitions an order to the shipped status.
type MarkOrderAsShippedUseCase struct {
	repo OrderRepository
}

// NewMarkOrderAsShippedUseCase creates a MarkOrderAsShippedUseCase backed
// by repo.
func NewMarkOrderAsShippedUseCase(repo OrderRepository) *MarkOrderAsShippedUseCase {
	return &MarkOrderAsShippedUseCase{repo: repo}
}

// Execute loads the order with the given ID, marks it as shipped, and
// persists the result. It returns ErrOrderNotFound if no order with that ID
// exists, or ErrInvalidStatusTransition if the order's current status does
// not allow the transition.
func (uc *MarkOrderAsShippedUseCase) Execute(ctx context.Context, id domain.OrderID) error {
	order, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if err := order.Ship(); err != nil {
		return err
	}

	return uc.repo.Save(ctx, order)
}
