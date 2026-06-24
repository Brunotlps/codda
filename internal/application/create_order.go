package application

import (
	"context"
	"fmt"

	"github.com/Brunotlps/codda/internal/domain"
)

// CreateOrderInput carries the raw data needed to create an order.
type CreateOrderInput struct {
	Items []CreateOrderItemInput
}

// CreateOrderItemInput carries the raw data needed to create a single item
// within a CreateOrderInput.
type CreateOrderItemInput struct {
	ProductID   string
	ProductName string
	PriceCents  int64
	Quantity    int
}

// CreateOrderUseCase creates a new order from raw input.
type CreateOrderUseCase struct {
	repo OrderRepository
}

// NewCreateOrderUseCase creates a CreateOrderUseCase backed by repo.
func NewCreateOrderUseCase(repo OrderRepository) *CreateOrderUseCase {
	return &CreateOrderUseCase{repo: repo}
}

// Execute builds an order from input and persists it, returning the new
// order's ID. It returns the first error encountered while building an
// item, wrapped with the item's index, or any error domain.NewOrder returns
// (e.g. ErrOrderRequiresItems).
func (uc *CreateOrderUseCase) Execute(ctx context.Context, input CreateOrderInput) (domain.OrderID, error) {
	items := make([]domain.OrderItem, 0, len(input.Items))
	for i, in := range input.Items {
		item, err := buildItem(in)
		if err != nil {
			return "", fmt.Errorf("item at index %d: %w", i, err)
		}
		items = append(items, item)
	}

	order, err := domain.NewOrder(items)
	if err != nil {
		return "", err
	}

	if err := uc.repo.Save(ctx, order); err != nil {
		return "", err
	}

	return order.ID(), nil
}

// buildItem constructs a domain.OrderItem from in.
func buildItem(in CreateOrderItemInput) (domain.OrderItem, error) {
	price, err := domain.NewMoney(in.PriceCents)
	if err != nil {
		return domain.OrderItem{}, err
	}

	return domain.NewOrderItem(in.ProductID, in.ProductName, price, in.Quantity)
}
