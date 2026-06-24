package application_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Brunotlps/codda/internal/adapters/memory"
	"github.com/Brunotlps/codda/internal/application"
	"github.com/Brunotlps/codda/internal/domain"
)

func TestListOrdersUseCase_Execute(t *testing.T) {
	t.Run("no filters returns all resumes", func(t *testing.T) {
		repo := memory.NewOrderRepository()
		for range 3 {
			if err := repo.Save(context.Background(), makeOrder(t)); err != nil {
				t.Fatalf("Save(...) returned unexpected error: %v", err)
			}
		}

		uc := application.NewListOrdersUseCase(repo)

		got, hasMore, err := uc.Execute(context.Background(), application.ListOrdersFilters{}, application.Pagination{})
		if err != nil {
			t.Fatalf("Execute(...) returned unexpected error: %v", err)
		}
		if len(got) != 3 {
			t.Errorf("Execute(...) returned %d resumes, want 3", len(got))
		}
		if hasMore {
			t.Errorf("Execute(...) hasMore = true, want false")
		}
	})

	t.Run("resume fields match order", func(t *testing.T) {
		repo := memory.NewOrderRepository()
		order := makeOrder(t)
		if err := order.MarkAsPaid(); err != nil {
			t.Fatalf("MarkAsPaid() setup failed: %v", err)
		}
		if err := repo.Save(context.Background(), order); err != nil {
			t.Fatalf("Save(...) returned unexpected error: %v", err)
		}

		uc := application.NewListOrdersUseCase(repo)

		got, _, err := uc.Execute(context.Background(), application.ListOrdersFilters{}, application.Pagination{})
		if err != nil {
			t.Fatalf("Execute(...) returned unexpected error: %v", err)
		}
		if len(got) != 1 {
			t.Fatalf("Execute(...) returned %d resumes, want 1", len(got))
		}

		resume := got[0]
		if resume.ID != order.ID() {
			t.Errorf("ID = %v, want %v", resume.ID, order.ID())
		}
		if resume.Status != order.Status() {
			t.Errorf("Status = %v, want %v", resume.Status, order.Status())
		}
		if resume.Total.Cents() != order.Total().Cents() {
			t.Errorf("Total = %v, want %v", resume.Total, order.Total())
		}
		if !resume.CreatedAt.Equal(order.CreatedAt()) {
			t.Errorf("CreatedAt = %v, want %v", resume.CreatedAt, order.CreatedAt())
		}
	})

	t.Run("pagination reports hasMore", func(t *testing.T) {
		repo := memory.NewOrderRepository()
		for range 5 {
			if err := repo.Save(context.Background(), makeOrder(t)); err != nil {
				t.Fatalf("Save(...) returned unexpected error: %v", err)
			}
		}

		uc := application.NewListOrdersUseCase(repo)

		got, hasMore, err := uc.Execute(context.Background(), application.ListOrdersFilters{}, application.Pagination{Limit: 2})
		if err != nil {
			t.Fatalf("Execute(...) returned unexpected error: %v", err)
		}
		if len(got) != 2 {
			t.Errorf("Execute(...) returned %d resumes, want 2", len(got))
		}
		if !hasMore {
			t.Errorf("Execute(...) hasMore = false, want true")
		}
	})

	t.Run("zero limit applies default", func(t *testing.T) {
		repo := memory.NewOrderRepository()
		for range 25 {
			if err := repo.Save(context.Background(), makeOrder(t)); err != nil {
				t.Fatalf("Save(...) returned unexpected error: %v", err)
			}
		}

		uc := application.NewListOrdersUseCase(repo)

		got, hasMore, err := uc.Execute(context.Background(), application.ListOrdersFilters{}, application.Pagination{Limit: 0})
		if err != nil {
			t.Fatalf("Execute(...) returned unexpected error: %v", err)
		}
		if len(got) != 20 {
			t.Errorf("Execute(...) returned %d resumes, want 20 (default limit)", len(got))
		}
		if !hasMore {
			t.Errorf("Execute(...) hasMore = false, want true")
		}
	})

	t.Run("limit above max is clamped", func(t *testing.T) {
		repo := memory.NewOrderRepository()
		for range 110 {
			if err := repo.Save(context.Background(), makeOrder(t)); err != nil {
				t.Fatalf("Save(...) returned unexpected error: %v", err)
			}
		}

		uc := application.NewListOrdersUseCase(repo)

		got, hasMore, err := uc.Execute(context.Background(), application.ListOrdersFilters{}, application.Pagination{Limit: 1000})
		if err != nil {
			t.Fatalf("Execute(...) returned unexpected error: %v", err)
		}
		if len(got) != 100 {
			t.Errorf("Execute(...) returned %d resumes, want 100 (max limit)", len(got))
		}
		if !hasMore {
			t.Errorf("Execute(...) hasMore = false, want true")
		}
	})

	t.Run("negative offset is clamped to zero", func(t *testing.T) {
		repo := memory.NewOrderRepository()
		for range 3 {
			if err := repo.Save(context.Background(), makeOrder(t)); err != nil {
				t.Fatalf("Save(...) returned unexpected error: %v", err)
			}
		}

		uc := application.NewListOrdersUseCase(repo)

		got, hasMore, err := uc.Execute(context.Background(), application.ListOrdersFilters{}, application.Pagination{Offset: -5, Limit: 10})
		if err != nil {
			t.Fatalf("Execute(...) returned unexpected error: %v", err)
		}
		if len(got) != 3 {
			t.Errorf("Execute(...) returned %d resumes, want 3", len(got))
		}
		if hasMore {
			t.Errorf("Execute(...) hasMore = true, want false")
		}
	})

	t.Run("filters by status", func(t *testing.T) {
		repo := memory.NewOrderRepository()

		pending := makeOrder(t)
		if err := repo.Save(context.Background(), pending); err != nil {
			t.Fatalf("Save(...) returned unexpected error: %v", err)
		}

		paid := makeOrder(t)
		if err := paid.MarkAsPaid(); err != nil {
			t.Fatalf("MarkAsPaid() setup failed: %v", err)
		}
		if err := repo.Save(context.Background(), paid); err != nil {
			t.Fatalf("Save(...) returned unexpected error: %v", err)
		}

		uc := application.NewListOrdersUseCase(repo)

		status := domain.StatusPaid
		got, _, err := uc.Execute(context.Background(), application.ListOrdersFilters{Status: &status}, application.Pagination{})
		if err != nil {
			t.Fatalf("Execute(...) returned unexpected error: %v", err)
		}
		if len(got) != 1 || got[0].ID != paid.ID() {
			t.Errorf("Execute(Status=paid) = %v, want only %v", got, paid.ID())
		}
	})

	t.Run("cancelled context returns error", func(t *testing.T) {
		repo := memory.NewOrderRepository()
		uc := application.NewListOrdersUseCase(repo)

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, _, err := uc.Execute(ctx, application.ListOrdersFilters{}, application.Pagination{})
		if !errors.Is(err, context.Canceled) {
			t.Errorf("Execute(...) error = %v, want %v", err, context.Canceled)
		}
	})
}
