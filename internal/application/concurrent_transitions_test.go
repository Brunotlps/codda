package application_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Brunotlps/codda/internal/adapters/memory"
	"github.com/Brunotlps/codda/internal/application"
	"github.com/Brunotlps/codda/internal/domain"
)

// simultaneousReadsRepository lets both transition use cases read the same
// persisted status before either one attempts to write.
type simultaneousReadsRepository struct {
	*memory.OrderRepository
	reads    atomic.Int32
	bothRead chan struct{}
}

func (r *simultaneousReadsRepository) FindByID(ctx context.Context, id domain.OrderID) (*domain.Order, error) {
	order, err := r.OrderRepository.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if r.reads.Add(1) == 2 {
		close(r.bothRead)
	}
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-r.bothRead:
		return order, nil
	}
}

func TestConcurrentTransitionsRejectStaleStatus(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	repo := &simultaneousReadsRepository{
		OrderRepository: memory.NewOrderRepository(),
		bothRead:        make(chan struct{}),
	}
	order := makeOrder(t)
	if err := repo.Save(ctx, order); err != nil {
		t.Fatalf("Save(...) returned unexpected error: %v", err)
	}

	results := make(chan error, 2)
	go func() { results <- application.NewMarkOrderAsPaidUseCase(repo).Execute(ctx, order.ID()) }()
	go func() { results <- application.NewMarkOrderAsCancelledUseCase(repo).Execute(ctx, order.ID()) }()

	successes, conflicts := 0, 0
	for range 2 {
		switch err := <-results; {
		case err == nil:
			successes++
		case errors.Is(err, domain.ErrInvalidStatusTransition):
			conflicts++
		default:
			t.Fatalf("transition error = %v, want nil or %v", err, domain.ErrInvalidStatusTransition)
		}
	}
	if successes != 1 || conflicts != 1 {
		t.Errorf("transition results = %d successes and %d conflicts, want one of each", successes, conflicts)
	}

	saved, err := repo.FindByID(ctx, order.ID())
	if err != nil {
		t.Fatalf("FindByID(...) returned unexpected error: %v", err)
	}
	if saved.Status() != domain.StatusPaid && saved.Status() != domain.StatusCancelled {
		t.Errorf("Status() = %v, want paid or cancelled", saved.Status())
	}
}
