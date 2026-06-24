package application_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Brunotlps/codda/internal/adapters/memory"
	"github.com/Brunotlps/codda/internal/application"
	"github.com/Brunotlps/codda/internal/domain"
)

func TestCreateOrderUseCase_Execute(t *testing.T) {
	t.Run("single item succeeds", func(t *testing.T) {
		repo := memory.NewOrderRepository()
		uc := application.NewCreateOrderUseCase(repo)

		input := application.CreateOrderInput{
			Items: []application.CreateOrderItemInput{
				{ProductID: "p1", ProductName: "Widget", PriceCents: 1000, Quantity: 2},
			},
		}

		id, err := uc.Execute(context.Background(), input)
		if err != nil {
			t.Fatalf("Execute(...) returned unexpected error: %v", err)
		}
		if id == "" {
			t.Errorf("Execute(...) returned empty ID")
		}

		order, err := repo.FindByID(context.Background(), id)
		if err != nil {
			t.Fatalf("FindByID(...) returned unexpected error: %v", err)
		}
		if got := order.Items(); len(got) != 1 || got[0].ProductID() != "p1" {
			t.Errorf("Items() = %v, want a single item with product id %q", got, "p1")
		}
	})

	t.Run("multiple distinct items succeeds", func(t *testing.T) {
		repo := memory.NewOrderRepository()
		uc := application.NewCreateOrderUseCase(repo)

		input := application.CreateOrderInput{
			Items: []application.CreateOrderItemInput{
				{ProductID: "p1", ProductName: "Widget", PriceCents: 1000, Quantity: 1},
				{ProductID: "p2", ProductName: "Gadget", PriceCents: 500, Quantity: 3},
			},
		}

		id, err := uc.Execute(context.Background(), input)
		if err != nil {
			t.Fatalf("Execute(...) returned unexpected error: %v", err)
		}

		order, err := repo.FindByID(context.Background(), id)
		if err != nil {
			t.Fatalf("FindByID(...) returned unexpected error: %v", err)
		}
		if got := order.Items(); len(got) != 2 {
			t.Errorf("Items() returned %d items, want 2", len(got))
		}
	})

	t.Run("duplicated product ids are merged", func(t *testing.T) {
		repo := memory.NewOrderRepository()
		uc := application.NewCreateOrderUseCase(repo)

		input := application.CreateOrderInput{
			Items: []application.CreateOrderItemInput{
				{ProductID: "p1", ProductName: "Widget", PriceCents: 1000, Quantity: 2},
				{ProductID: "p1", ProductName: "Widget", PriceCents: 1000, Quantity: 3},
			},
		}

		id, err := uc.Execute(context.Background(), input)
		if err != nil {
			t.Fatalf("Execute(...) returned unexpected error: %v", err)
		}

		order, err := repo.FindByID(context.Background(), id)
		if err != nil {
			t.Fatalf("FindByID(...) returned unexpected error: %v", err)
		}
		got := order.Items()
		if len(got) != 1 || got[0].Quantity() != 5 {
			t.Errorf("Items() = %v, want a single item with quantity 5", got)
		}
	})

	t.Run("empty items list fails", func(t *testing.T) {
		repo := memory.NewOrderRepository()
		uc := application.NewCreateOrderUseCase(repo)

		_, err := uc.Execute(context.Background(), application.CreateOrderInput{})
		if !errors.Is(err, domain.ErrOrderRequiresItems) {
			t.Errorf("Execute(...) error = %v, want %v", err, domain.ErrOrderRequiresItems)
		}
	})

	t.Run("negative price fails", func(t *testing.T) {
		repo := memory.NewOrderRepository()
		uc := application.NewCreateOrderUseCase(repo)

		input := application.CreateOrderInput{
			Items: []application.CreateOrderItemInput{
				{ProductID: "p1", ProductName: "Widget", PriceCents: -1, Quantity: 1},
			},
		}

		_, err := uc.Execute(context.Background(), input)
		if !errors.Is(err, domain.ErrNegativeMoney) {
			t.Errorf("Execute(...) error = %v, want %v", err, domain.ErrNegativeMoney)
		}
	})

	t.Run("zero price fails", func(t *testing.T) {
		repo := memory.NewOrderRepository()
		uc := application.NewCreateOrderUseCase(repo)

		input := application.CreateOrderInput{
			Items: []application.CreateOrderItemInput{
				{ProductID: "p1", ProductName: "Widget", PriceCents: 0, Quantity: 1},
			},
		}

		_, err := uc.Execute(context.Background(), input)
		if !errors.Is(err, domain.ErrInvalidPrice) {
			t.Errorf("Execute(...) error = %v, want %v", err, domain.ErrInvalidPrice)
		}
	})

	t.Run("invalid quantity fails", func(t *testing.T) {
		repo := memory.NewOrderRepository()
		uc := application.NewCreateOrderUseCase(repo)

		input := application.CreateOrderInput{
			Items: []application.CreateOrderItemInput{
				{ProductID: "p1", ProductName: "Widget", PriceCents: 1000, Quantity: 0},
			},
		}

		_, err := uc.Execute(context.Background(), input)
		if !errors.Is(err, domain.ErrInvalidQuantity) {
			t.Errorf("Execute(...) error = %v, want %v", err, domain.ErrInvalidQuantity)
		}
	})

	t.Run("empty product id fails", func(t *testing.T) {
		repo := memory.NewOrderRepository()
		uc := application.NewCreateOrderUseCase(repo)

		input := application.CreateOrderInput{
			Items: []application.CreateOrderItemInput{
				{ProductID: "", ProductName: "Widget", PriceCents: 1000, Quantity: 1},
			},
		}

		_, err := uc.Execute(context.Background(), input)
		if !errors.Is(err, domain.ErrEmptyProductID) {
			t.Errorf("Execute(...) error = %v, want %v", err, domain.ErrEmptyProductID)
		}
	})

	t.Run("cancelled context fails to save", func(t *testing.T) {
		repo := memory.NewOrderRepository()
		uc := application.NewCreateOrderUseCase(repo)

		input := application.CreateOrderInput{
			Items: []application.CreateOrderItemInput{
				{ProductID: "p1", ProductName: "Widget", PriceCents: 1000, Quantity: 1},
			},
		}

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := uc.Execute(ctx, input)
		if !errors.Is(err, context.Canceled) {
			t.Errorf("Execute(...) error = %v, want %v", err, context.Canceled)
		}
	})
}
