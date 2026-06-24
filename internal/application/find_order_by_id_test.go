package application_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Brunotlps/codda/internal/adapters/memory"
	"github.com/Brunotlps/codda/internal/application"
	"github.com/Brunotlps/codda/internal/domain"
)

func TestFindOrderByIDUseCase_Execute(t *testing.T) {
	t.Run("order exists", func(t *testing.T) {
		repo := memory.NewOrderRepository()
		order := makeOrder(t)
		if err := repo.Save(context.Background(), order); err != nil {
			t.Fatalf("Save(...) returned unexpected error: %v", err)
		}

		uc := application.NewFindOrderByIDUseCase(repo)

		got, err := uc.Execute(context.Background(), order.ID())
		if err != nil {
			t.Fatalf("Execute(...) returned unexpected error: %v", err)
		}
		if got.ID() != order.ID() {
			t.Errorf("Execute(...).ID() = %v, want %v", got.ID(), order.ID())
		}
	})

	t.Run("order does not exist", func(t *testing.T) {
		repo := memory.NewOrderRepository()
		uc := application.NewFindOrderByIDUseCase(repo)

		_, err := uc.Execute(context.Background(), domain.OrderID("missing"))
		if !errors.Is(err, application.ErrOrderNotFound) {
			t.Errorf("Execute(...) error = %v, want %v", err, application.ErrOrderNotFound)
		}
	})
}
