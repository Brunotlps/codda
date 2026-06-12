package memory_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/Brunotlps/codda/internal/adapters/memory"
	"github.com/Brunotlps/codda/internal/application"
	"github.com/Brunotlps/codda/internal/domain"
)

// makeOrder builds a domain.Order with a single item, for test setup,
// failing the test immediately if construction fails.
func makeOrder(t *testing.T, productID string, cents int64, quantity int) *domain.Order {
	t.Helper()

	price, err := domain.NewMoney(cents)
	if err != nil {
		t.Fatalf("NewMoney(%d) failed during test setup: %v", cents, err)
	}

	item, err := domain.NewOrderItem(productID, "Widget", price, quantity)
	if err != nil {
		t.Fatalf("NewOrderItem(...) failed during test setup: %v", err)
	}

	order, err := domain.NewOrder([]domain.OrderItem{item})
	if err != nil {
		t.Fatalf("NewOrder(...) failed during test setup: %v", err)
	}

	return order
}

func TestOrderRepository_Save(t *testing.T) {
	t.Run("saves a new order", func(t *testing.T) {
		repo := memory.NewOrderRepository()
		order := makeOrder(t, "p1", 1000, 1)

		if err := repo.Save(context.Background(), order); err != nil {
			t.Fatalf("Save(...) returned unexpected error: %v", err)
		}

		got, err := repo.FindByID(context.Background(), order.ID())
		if err != nil {
			t.Fatalf("FindByID(...) returned unexpected error: %v", err)
		}
		if got.ID() != order.ID() {
			t.Errorf("FindByID(...).ID() = %v, want %v", got.ID(), order.ID())
		}
	})

	t.Run("updates an existing order", func(t *testing.T) {
		repo := memory.NewOrderRepository()
		order := makeOrder(t, "p1", 1000, 1)

		if err := repo.Save(context.Background(), order); err != nil {
			t.Fatalf("Save(...) returned unexpected error: %v", err)
		}

		if err := order.MarkAsPaid(); err != nil {
			t.Fatalf("MarkAsPaid() returned unexpected error: %v", err)
		}
		if err := repo.Save(context.Background(), order); err != nil {
			t.Fatalf("Save(...) returned unexpected error: %v", err)
		}

		got, err := repo.FindByID(context.Background(), order.ID())
		if err != nil {
			t.Fatalf("FindByID(...) returned unexpected error: %v", err)
		}
		if got.Status() != domain.StatusPaid {
			t.Errorf("FindByID(...).Status() = %v, want %v", got.Status(), domain.StatusPaid)
		}

		all, _, err := repo.List(context.Background(), application.ListOrdersFilters{}, application.Pagination{Limit: 10})
		if err != nil {
			t.Fatalf("List(...) returned unexpected error: %v", err)
		}
		if len(all) != 1 {
			t.Errorf("List(...) returned %d orders, want 1", len(all))
		}
	})

	t.Run("cancelled context returns error", func(t *testing.T) {
		repo := memory.NewOrderRepository()
		order := makeOrder(t, "p1", 1000, 1)

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		if err := repo.Save(ctx, order); !errors.Is(err, context.Canceled) {
			t.Errorf("Save(...) error = %v, want %v", err, context.Canceled)
		}
	})
}

func TestOrderRepository_FindByID(t *testing.T) {
	t.Run("order present", func(t *testing.T) {
		repo := memory.NewOrderRepository()
		order := makeOrder(t, "p1", 1000, 1)

		if err := repo.Save(context.Background(), order); err != nil {
			t.Fatalf("Save(...) returned unexpected error: %v", err)
		}

		got, err := repo.FindByID(context.Background(), order.ID())
		if err != nil {
			t.Fatalf("FindByID(...) returned unexpected error: %v", err)
		}
		if got.ID() != order.ID() {
			t.Errorf("FindByID(...).ID() = %v, want %v", got.ID(), order.ID())
		}
	})

	t.Run("order absent", func(t *testing.T) {
		repo := memory.NewOrderRepository()

		_, err := repo.FindByID(context.Background(), domain.OrderID("missing"))
		if !errors.Is(err, application.ErrOrderNotFound) {
			t.Errorf("FindByID(...) error = %v, want %v", err, application.ErrOrderNotFound)
		}
	})

	t.Run("cancelled context returns error", func(t *testing.T) {
		repo := memory.NewOrderRepository()

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := repo.FindByID(ctx, domain.OrderID("p1"))
		if !errors.Is(err, context.Canceled) {
			t.Errorf("FindByID(...) error = %v, want %v", err, context.Canceled)
		}
	})
}

func TestOrderRepository_List(t *testing.T) {
	t.Run("empty repository", func(t *testing.T) {
		repo := memory.NewOrderRepository()

		got, hasMore, err := repo.List(context.Background(), application.ListOrdersFilters{}, application.Pagination{Limit: 10})
		if err != nil {
			t.Fatalf("List(...) returned unexpected error: %v", err)
		}
		if len(got) != 0 {
			t.Errorf("List(...) returned %d orders, want 0", len(got))
		}
		if hasMore {
			t.Errorf("List(...) hasMore = true, want false")
		}
	})

	t.Run("no filters returns all orders", func(t *testing.T) {
		repo := memory.NewOrderRepository()

		for _, productID := range []string{"p1", "p2", "p3"} {
			order := makeOrder(t, productID, 1000, 1)
			if err := repo.Save(context.Background(), order); err != nil {
				t.Fatalf("Save(...) returned unexpected error: %v", err)
			}
		}

		got, hasMore, err := repo.List(context.Background(), application.ListOrdersFilters{}, application.Pagination{Limit: 10})
		if err != nil {
			t.Fatalf("List(...) returned unexpected error: %v", err)
		}
		if len(got) != 3 {
			t.Errorf("List(...) returned %d orders, want 3", len(got))
		}
		if hasMore {
			t.Errorf("List(...) hasMore = true, want false")
		}
	})

	t.Run("filters by status", func(t *testing.T) {
		repo := memory.NewOrderRepository()

		pending := makeOrder(t, "p1", 1000, 1)
		paid := makeOrder(t, "p2", 1000, 1)
		if err := paid.MarkAsPaid(); err != nil {
			t.Fatalf("MarkAsPaid() returned unexpected error: %v", err)
		}

		for _, order := range []*domain.Order{pending, paid} {
			if err := repo.Save(context.Background(), order); err != nil {
				t.Fatalf("Save(...) returned unexpected error: %v", err)
			}
		}

		status := domain.StatusPaid
		got, _, err := repo.List(context.Background(), application.ListOrdersFilters{Status: &status}, application.Pagination{Limit: 10})
		if err != nil {
			t.Fatalf("List(...) returned unexpected error: %v", err)
		}
		if len(got) != 1 || got[0].ID() != paid.ID() {
			t.Errorf("List(Status=paid) = %v, want only %v", got, paid.ID())
		}
	})

	t.Run("filters by created date range", func(t *testing.T) {
		repo := memory.NewOrderRepository()

		older := makeOrder(t, "p1", 1000, 1)
		if err := repo.Save(context.Background(), older); err != nil {
			t.Fatalf("Save(...) returned unexpected error: %v", err)
		}

		time.Sleep(time.Millisecond)
		cutoff := time.Now()
		time.Sleep(time.Millisecond)

		newer := makeOrder(t, "p2", 1000, 1)
		if err := repo.Save(context.Background(), newer); err != nil {
			t.Fatalf("Save(...) returned unexpected error: %v", err)
		}

		gotFrom, _, err := repo.List(context.Background(), application.ListOrdersFilters{CreatedFrom: &cutoff}, application.Pagination{Limit: 10})
		if err != nil {
			t.Fatalf("List(...) returned unexpected error: %v", err)
		}
		if len(gotFrom) != 1 || gotFrom[0].ID() != newer.ID() {
			t.Errorf("List(CreatedFrom=cutoff) = %v, want only %v", gotFrom, newer.ID())
		}

		gotTo, _, err := repo.List(context.Background(), application.ListOrdersFilters{CreatedTo: &cutoff}, application.Pagination{Limit: 10})
		if err != nil {
			t.Fatalf("List(...) returned unexpected error: %v", err)
		}
		if len(gotTo) != 1 || gotTo[0].ID() != older.ID() {
			t.Errorf("List(CreatedTo=cutoff) = %v, want only %v", gotTo, older.ID())
		}
	})

	t.Run("filters by price range", func(t *testing.T) {
		repo := memory.NewOrderRepository()

		cheap := makeOrder(t, "p1", 1000, 1)
		expensive := makeOrder(t, "p2", 5000, 1)

		for _, order := range []*domain.Order{cheap, expensive} {
			if err := repo.Save(context.Background(), order); err != nil {
				t.Fatalf("Save(...) returned unexpected error: %v", err)
			}
		}

		threshold, err := domain.NewMoney(2000)
		if err != nil {
			t.Fatalf("NewMoney(...) failed: %v", err)
		}

		gotMin, _, err := repo.List(context.Background(), application.ListOrdersFilters{PriceMin: &threshold}, application.Pagination{Limit: 10})
		if err != nil {
			t.Fatalf("List(...) returned unexpected error: %v", err)
		}
		if len(gotMin) != 1 || gotMin[0].ID() != expensive.ID() {
			t.Errorf("List(PriceMin=2000) = %v, want only %v", gotMin, expensive.ID())
		}

		gotMax, _, err := repo.List(context.Background(), application.ListOrdersFilters{PriceMax: &threshold}, application.Pagination{Limit: 10})
		if err != nil {
			t.Fatalf("List(...) returned unexpected error: %v", err)
		}
		if len(gotMax) != 1 || gotMax[0].ID() != cheap.ID() {
			t.Errorf("List(PriceMax=2000) = %v, want only %v", gotMax, cheap.ID())
		}
	})

	t.Run("combines multiple filters", func(t *testing.T) {
		repo := memory.NewOrderRepository()

		pendingCheap := makeOrder(t, "p1", 1000, 1)

		paidCheap := makeOrder(t, "p2", 1000, 1)
		if err := paidCheap.MarkAsPaid(); err != nil {
			t.Fatalf("MarkAsPaid() returned unexpected error: %v", err)
		}

		paidExpensive := makeOrder(t, "p3", 5000, 1)
		if err := paidExpensive.MarkAsPaid(); err != nil {
			t.Fatalf("MarkAsPaid() returned unexpected error: %v", err)
		}

		for _, order := range []*domain.Order{pendingCheap, paidCheap, paidExpensive} {
			if err := repo.Save(context.Background(), order); err != nil {
				t.Fatalf("Save(...) returned unexpected error: %v", err)
			}
		}

		status := domain.StatusPaid
		priceMax, err := domain.NewMoney(2000)
		if err != nil {
			t.Fatalf("NewMoney(...) failed: %v", err)
		}

		got, _, err := repo.List(context.Background(), application.ListOrdersFilters{Status: &status, PriceMax: &priceMax}, application.Pagination{Limit: 10})
		if err != nil {
			t.Fatalf("List(...) returned unexpected error: %v", err)
		}
		if len(got) != 1 || got[0].ID() != paidCheap.ID() {
			t.Errorf("List(Status=paid, PriceMax=2000) = %v, want only %v", got, paidCheap.ID())
		}
	})

	t.Run("pagination", func(t *testing.T) {
		repo := memory.NewOrderRepository()

		var orders []*domain.Order
		for i := range 5 {
			order := makeOrder(t, fmt.Sprintf("p%d", i), 1000, 1)
			if err := repo.Save(context.Background(), order); err != nil {
				t.Fatalf("Save(...) returned unexpected error: %v", err)
			}
			orders = append(orders, order)
			time.Sleep(time.Millisecond)
		}

		// orders are sorted by createdAt descending, so orders[4] (created
		// last) comes first and orders[0] (created first) comes last.

		t.Run("first page has more", func(t *testing.T) {
			got, hasMore, err := repo.List(context.Background(), application.ListOrdersFilters{}, application.Pagination{Offset: 0, Limit: 2})
			if err != nil {
				t.Fatalf("List(...) returned unexpected error: %v", err)
			}
			if len(got) != 2 || got[0].ID() != orders[4].ID() || got[1].ID() != orders[3].ID() {
				t.Errorf("List(Offset=0, Limit=2) = %v, want [%v, %v]", got, orders[4].ID(), orders[3].ID())
			}
			if !hasMore {
				t.Errorf("List(...) hasMore = false, want true")
			}
		})

		t.Run("last page has no more", func(t *testing.T) {
			got, hasMore, err := repo.List(context.Background(), application.ListOrdersFilters{}, application.Pagination{Offset: 4, Limit: 2})
			if err != nil {
				t.Fatalf("List(...) returned unexpected error: %v", err)
			}
			if len(got) != 1 || got[0].ID() != orders[0].ID() {
				t.Errorf("List(Offset=4, Limit=2) = %v, want [%v]", got, orders[0].ID())
			}
			if hasMore {
				t.Errorf("List(...) hasMore = true, want false")
			}
		})

		t.Run("offset beyond total", func(t *testing.T) {
			got, hasMore, err := repo.List(context.Background(), application.ListOrdersFilters{}, application.Pagination{Offset: 100, Limit: 2})
			if err != nil {
				t.Fatalf("List(...) returned unexpected error: %v", err)
			}
			if len(got) != 0 {
				t.Errorf("List(Offset=100, Limit=2) returned %d orders, want 0", len(got))
			}
			if hasMore {
				t.Errorf("List(...) hasMore = true, want false")
			}
		})
	})

	t.Run("cancelled context returns error", func(t *testing.T) {
		repo := memory.NewOrderRepository()

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, _, err := repo.List(ctx, application.ListOrdersFilters{}, application.Pagination{Limit: 10})
		if !errors.Is(err, context.Canceled) {
			t.Errorf("List(...) error = %v, want %v", err, context.Canceled)
		}
	})
}

// TestOrderRepository_ConcurrentAccess exercises Save, FindByID, and List
// from multiple goroutines at once. Run with -race to detect data races.
func TestOrderRepository_ConcurrentAccess(t *testing.T) {
	repo := memory.NewOrderRepository()

	const n = 50

	orders := make([]*domain.Order, n)
	for i := range orders {
		orders[i] = makeOrder(t, fmt.Sprintf("p%d", i), 1000, 1)
	}

	var wg sync.WaitGroup
	for _, order := range orders {
		wg.Add(1)
		go func(order *domain.Order) {
			defer wg.Done()

			if err := repo.Save(context.Background(), order); err != nil {
				t.Errorf("Save(...) returned unexpected error: %v", err)
			}
			if _, err := repo.FindByID(context.Background(), order.ID()); err != nil {
				t.Errorf("FindByID(...) returned unexpected error: %v", err)
			}
			if _, _, err := repo.List(context.Background(), application.ListOrdersFilters{}, application.Pagination{Limit: n}); err != nil {
				t.Errorf("List(...) returned unexpected error: %v", err)
			}
		}(order)
	}

	wg.Wait()
}
