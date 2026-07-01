package postgres_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Brunotlps/codda/internal/adapters/postgres"
	"github.com/Brunotlps/codda/internal/application"
	"github.com/Brunotlps/codda/internal/domain"
)

func makeOrder(t *testing.T, productID string, cents int64, quantity int) *domain.Order {
	t.Helper()

	price, err := domain.NewMoney(cents)
	if err != nil {
		t.Fatalf("NewMoney(%d): %v", cents, err)
	}
	item, err := domain.NewOrderItem(productID, "Product "+productID, price, quantity)
	if err != nil {
		t.Fatalf("NewOrderItem: %v", err)
	}
	order, err := domain.NewOrder([]domain.OrderItem{item})
	if err != nil {
		t.Fatalf("NewOrder: %v", err)
	}
	return order
}

func TestOrderRepository_Save(t *testing.T) {
	ctx := context.Background()

	t.Run("saves_new_order", func(t *testing.T) {
		truncateTables(t)
		repo := postgres.NewOrderRepository(testPool)

		order := makeOrder(t, "p1", 1000, 2)
		if err := repo.Save(ctx, order); err != nil {
			t.Fatalf("Save: %v", err)
		}

		saved, err := repo.FindByID(ctx, order.ID())
		if err != nil {
			t.Fatalf("FindByID: %v", err)
		}

		if saved.ID() != order.ID() {
			t.Errorf("ID = %q, want %q", saved.ID(), order.ID())
		}
		if saved.Status() != order.Status() {
			t.Errorf("Status = %q, want %q", saved.Status(), order.Status())
		}
		if got := len(saved.Items()); got != 1 {
			t.Fatalf("len(Items()) = %d, want 1", got)
		}
		if got := saved.Items()[0].ProductID(); got != "p1" {
			t.Errorf("Items[0].ProductID() = %q, want %q", got, "p1")
		}
		if got := saved.Items()[0].Quantity(); got != 2 {
			t.Errorf("Items[0].Quantity() = %d, want 2", got)
		}
		if got := saved.Items()[0].Price().Cents(); got != 1000 {
			t.Errorf("Items[0].Price().Cents() = %d, want 1000", got)
		}
		if !saved.CreatedAt().Round(time.Millisecond).Equal(order.CreatedAt().Round(time.Millisecond)) {
			t.Errorf("CreatedAt = %v, want ~%v", saved.CreatedAt(), order.CreatedAt())
		}
	})

	t.Run("updates_existing_order", func(t *testing.T) {
		truncateTables(t)
		repo := postgres.NewOrderRepository(testPool)

		order := makeOrder(t, "p1", 1000, 1)
		if err := repo.Save(ctx, order); err != nil {
			t.Fatalf("Save (initial): %v", err)
		}
		if err := order.MarkAsPaid(); err != nil {
			t.Fatalf("MarkAsPaid: %v", err)
		}
		if err := repo.Save(ctx, order); err != nil {
			t.Fatalf("Save (update): %v", err)
		}

		saved, err := repo.FindByID(ctx, order.ID())
		if err != nil {
			t.Fatalf("FindByID: %v", err)
		}
		if saved.Status() != domain.StatusPaid {
			t.Errorf("Status = %q, want %q", saved.Status(), domain.StatusPaid)
		}
	})

	t.Run("preserves_item_order", func(t *testing.T) {
		truncateTables(t)
		repo := postgres.NewOrderRepository(testPool)

		p1Price, _ := domain.NewMoney(1000)
		p2Price, _ := domain.NewMoney(2000)
		p3Price, _ := domain.NewMoney(500)
		item1, _ := domain.NewOrderItem("p1", "Widget", p1Price, 1)
		item2, _ := domain.NewOrderItem("p2", "Gadget", p2Price, 3)
		item3, _ := domain.NewOrderItem("p3", "Doohickey", p3Price, 2)
		order, err := domain.NewOrder([]domain.OrderItem{item1, item2, item3})
		if err != nil {
			t.Fatalf("NewOrder: %v", err)
		}

		if err := repo.Save(ctx, order); err != nil {
			t.Fatalf("Save: %v", err)
		}

		saved, err := repo.FindByID(ctx, order.ID())
		if err != nil {
			t.Fatalf("FindByID: %v", err)
		}

		savedItems := saved.Items()
		if got := len(savedItems); got != 3 {
			t.Fatalf("len(Items()) = %d, want 3", got)
		}
		wantProductIDs := []string{"p1", "p2", "p3"}
		for i, want := range wantProductIDs {
			if got := savedItems[i].ProductID(); got != want {
				t.Errorf("Items[%d].ProductID() = %q, want %q", i, got, want)
			}
		}
	})

	t.Run("replaces_items_on_save", func(t *testing.T) {
		truncateTables(t)
		repo := postgres.NewOrderRepository(testPool)

		order := makeOrder(t, "p1", 1000, 2)
		if err := repo.Save(ctx, order); err != nil {
			t.Fatalf("Save (first): %v", err)
		}
		if err := repo.Save(ctx, order); err != nil {
			t.Fatalf("Save (second): %v", err)
		}

		saved, err := repo.FindByID(ctx, order.ID())
		if err != nil {
			t.Fatalf("FindByID: %v", err)
		}
		if got := len(saved.Items()); got != 1 {
			t.Errorf("len(Items()) = %d after double save, want 1 (delete-then-insert must not accumulate)", got)
		}
	})
}

func TestOrderRepository_FindByID(t *testing.T) {
	ctx := context.Background()

	t.Run("found", func(t *testing.T) {
		truncateTables(t)
		repo := postgres.NewOrderRepository(testPool)

		order := makeOrder(t, "p1", 1500, 1)
		if err := repo.Save(ctx, order); err != nil {
			t.Fatalf("Save: %v", err)
		}

		saved, err := repo.FindByID(ctx, order.ID())
		if err != nil {
			t.Fatalf("FindByID: unexpected error: %v", err)
		}
		if saved.ID() != order.ID() {
			t.Errorf("ID = %q, want %q", saved.ID(), order.ID())
		}
	})

	t.Run("not_found", func(t *testing.T) {
		truncateTables(t)
		repo := postgres.NewOrderRepository(testPool)

		_, err := repo.FindByID(ctx, domain.OrderID("ffffffff-ffff-ffff-ffff-ffffffffffff"))
		if !errors.Is(err, application.ErrOrderNotFound) {
			t.Errorf("FindByID() error = %v, want %v", err, application.ErrOrderNotFound)
		}
	})
}

func TestOrderRepository_List(t *testing.T) {
	ctx := context.Background()

	t.Run("no_filters_returns_all", func(t *testing.T) {
		truncateTables(t)
		repo := postgres.NewOrderRepository(testPool)

		for i := range 3 {
			order := makeOrder(t, "p1", 1000, 1)
			if err := repo.Save(ctx, order); err != nil {
				t.Fatalf("Save [%d]: %v", i, err)
			}
		}

		orders, hasMore, err := repo.List(ctx, application.ListOrdersFilters{}, application.Pagination{Limit: 10})
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		if got := len(orders); got != 3 {
			t.Errorf("len(orders) = %d, want 3", got)
		}
		if hasMore {
			t.Errorf("hasMore = true, want false")
		}
	})

	t.Run("orders_returned_by_created_at_desc", func(t *testing.T) {
		truncateTables(t)
		repo := postgres.NewOrderRepository(testPool)

		now := time.Now().UTC()
		price, _ := domain.NewMoney(1000)
		item, _ := domain.NewOrderItem("p1", "Widget", price, 1)
		items := []domain.OrderItem{item}

		oldest, _ := domain.HydrateOrder(domain.OrderID("aaaa0000-0000-0000-0000-000000000001"), items, domain.StatusPending, now.Add(-2*time.Hour))
		middle, _ := domain.HydrateOrder(domain.OrderID("aaaa0000-0000-0000-0000-000000000002"), items, domain.StatusPending, now.Add(-1*time.Hour))
		newest, _ := domain.HydrateOrder(domain.OrderID("aaaa0000-0000-0000-0000-000000000003"), items, domain.StatusPending, now)

		for _, o := range []*domain.Order{oldest, middle, newest} {
			if err := repo.Save(ctx, o); err != nil {
				t.Fatalf("Save: %v", err)
			}
		}

		orders, _, err := repo.List(ctx, application.ListOrdersFilters{}, application.Pagination{Limit: 10})
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		if got := len(orders); got != 3 {
			t.Fatalf("len(orders) = %d, want 3", got)
		}

		wantIDs := []domain.OrderID{newest.ID(), middle.ID(), oldest.ID()}
		for i, want := range wantIDs {
			if got := orders[i].ID(); got != want {
				t.Errorf("orders[%d].ID() = %q, want %q", i, got, want)
			}
		}
	})

	t.Run("filters_by_status", func(t *testing.T) {
		truncateTables(t)
		repo := postgres.NewOrderRepository(testPool)

		pending := makeOrder(t, "p1", 1000, 1)
		if err := repo.Save(ctx, pending); err != nil {
			t.Fatalf("Save pending: %v", err)
		}

		paid := makeOrder(t, "p2", 2000, 1)
		if err := paid.MarkAsPaid(); err != nil {
			t.Fatalf("MarkAsPaid: %v", err)
		}
		if err := repo.Save(ctx, paid); err != nil {
			t.Fatalf("Save paid: %v", err)
		}

		status := domain.StatusPaid
		orders, _, err := repo.List(ctx, application.ListOrdersFilters{Status: &status}, application.Pagination{Limit: 10})
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		if got := len(orders); got != 1 {
			t.Fatalf("len(orders) = %d, want 1", got)
		}
		if got := orders[0].ID(); got != paid.ID() {
			t.Errorf("orders[0].ID() = %q, want %q", got, paid.ID())
		}
	})

	t.Run("filters_by_price_range", func(t *testing.T) {
		truncateTables(t)
		repo := postgres.NewOrderRepository(testPool)

		cheap := makeOrder(t, "p1", 500, 1)
		if err := repo.Save(ctx, cheap); err != nil {
			t.Fatalf("Save cheap: %v", err)
		}

		expensive := makeOrder(t, "p2", 5000, 1)
		if err := repo.Save(ctx, expensive); err != nil {
			t.Fatalf("Save expensive: %v", err)
		}

		minPrice, _ := domain.NewMoney(1000)
		orders, _, err := repo.List(ctx, application.ListOrdersFilters{PriceMin: &minPrice}, application.Pagination{Limit: 10})
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		if got := len(orders); got != 1 {
			t.Fatalf("len(orders) = %d, want 1", got)
		}
		if got := orders[0].ID(); got != expensive.ID() {
			t.Errorf("orders[0].ID() = %q, want %q", got, expensive.ID())
		}
	})

	t.Run("pagination_reports_has_more", func(t *testing.T) {
		truncateTables(t)
		repo := postgres.NewOrderRepository(testPool)

		for i := range 5 {
			order := makeOrder(t, "p1", 1000, 1)
			if err := repo.Save(ctx, order); err != nil {
				t.Fatalf("Save [%d]: %v", i, err)
			}
		}

		orders, hasMore, err := repo.List(ctx, application.ListOrdersFilters{}, application.Pagination{Limit: 2})
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		if got := len(orders); got != 2 {
			t.Errorf("len(orders) = %d, want 2", got)
		}
		if !hasMore {
			t.Errorf("hasMore = false, want true")
		}
	})
}
