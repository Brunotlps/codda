package application_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Brunotlps/codda/internal/adapters/memory"
	"github.com/Brunotlps/codda/internal/application"
	"github.com/Brunotlps/codda/internal/domain"
)

func TestMarkOrderAsCancelledUseCase_Execute(t *testing.T) {
	t.Run("order is pending", func(t *testing.T) {
		repo := memory.NewOrderRepository()
		order := makeOrder(t)
		if err := repo.Save(context.Background(), order); err != nil {
			t.Fatalf("Save(...) returned unexpected error: %v", err)
		}

		uc := application.NewMarkOrderAsCancelledUseCase(repo)

		if err := uc.Execute(context.Background(), order.ID()); err != nil {
			t.Fatalf("Execute(...) returned unexpected error: %v", err)
		}

		got, err := repo.FindByID(context.Background(), order.ID())
		if err != nil {
			t.Fatalf("FindByID(...) returned unexpected error: %v", err)
		}
		if got.Status() != domain.StatusCancelled {
			t.Errorf("Status() = %v, want %v", got.Status(), domain.StatusCancelled)
		}
	})

	t.Run("order is paid", func(t *testing.T) {
		repo := memory.NewOrderRepository()
		order := makeOrder(t)
		if err := order.MarkAsPaid(); err != nil {
			t.Fatalf("MarkAsPaid() setup failed: %v", err)
		}
		if err := repo.Save(context.Background(), order); err != nil {
			t.Fatalf("Save(...) returned unexpected error: %v", err)
		}

		uc := application.NewMarkOrderAsCancelledUseCase(repo)

		if err := uc.Execute(context.Background(), order.ID()); err != nil {
			t.Fatalf("Execute(...) returned unexpected error: %v", err)
		}

		got, err := repo.FindByID(context.Background(), order.ID())
		if err != nil {
			t.Fatalf("FindByID(...) returned unexpected error: %v", err)
		}
		if got.Status() != domain.StatusCancelled {
			t.Errorf("Status() = %v, want %v", got.Status(), domain.StatusCancelled)
		}
	})

	t.Run("order does not exist", func(t *testing.T) {
		repo := memory.NewOrderRepository()
		uc := application.NewMarkOrderAsCancelledUseCase(repo)

		err := uc.Execute(context.Background(), domain.OrderID("missing"))
		if !errors.Is(err, application.ErrOrderNotFound) {
			t.Errorf("Execute(...) error = %v, want %v", err, application.ErrOrderNotFound)
		}
	})

	t.Run("order cannot transition to cancelled", func(t *testing.T) {
		repo := memory.NewOrderRepository()
		order := makeOrder(t)
		if err := order.MarkAsPaid(); err != nil {
			t.Fatalf("MarkAsPaid() setup failed: %v", err)
		}
		if err := order.Ship(); err != nil {
			t.Fatalf("Ship() setup failed: %v", err)
		}
		if err := repo.Save(context.Background(), order); err != nil {
			t.Fatalf("Save(...) returned unexpected error: %v", err)
		}

		uc := application.NewMarkOrderAsCancelledUseCase(repo)

		err := uc.Execute(context.Background(), order.ID())
		if !errors.Is(err, domain.ErrInvalidStatusTransition) {
			t.Errorf("Execute(...) error = %v, want %v", err, domain.ErrInvalidStatusTransition)
		}
	})
}
