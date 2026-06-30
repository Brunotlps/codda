package domain_test

import (
	"errors"
	"math"
	"testing"
	"time"

	"github.com/Brunotlps/codda/internal/domain"
)

// makeItem builds a domain.OrderItem for test setup, failing the test
// immediately if construction fails.
func makeItem(t *testing.T, productID, productName string, cents int64, quantity int) domain.OrderItem {
	t.Helper()

	item, err := domain.NewOrderItem(productID, productName, makeMoney(t, cents), quantity)
	if err != nil {
		t.Fatalf("NewOrderItem(%q, %q, %d, %d) failed during test setup: %v", productID, productName, cents, quantity, err)
	}

	return item
}

// makeOrder builds a domain.Order with a single item, for test setup,
// failing the test immediately if construction fails.
func makeOrder(t *testing.T) *domain.Order {
	t.Helper()

	order, err := domain.NewOrder([]domain.OrderItem{makeItem(t, "p1", "Widget", 1000, 1)})
	if err != nil {
		t.Fatalf("NewOrder(...) failed during test setup: %v", err)
	}

	return order
}

func TestNewOrder(t *testing.T) {
	t.Run("creates order successfully", func(t *testing.T) {
		before := time.Now()
		item := makeItem(t, "p1", "Widget", 1000, 2)

		order, err := domain.NewOrder([]domain.OrderItem{item})
		after := time.Now()

		if err != nil {
			t.Fatalf("NewOrder(...) returned unexpected error: %v", err)
		}
		if order.ID() == "" {
			t.Errorf("ID() = %q, want non-empty", order.ID())
		}
		if got := order.Status(); got != domain.StatusPending {
			t.Errorf("Status() = %v, want %v", got, domain.StatusPending)
		}
		if got := order.Items(); len(got) != 1 || got[0].ProductID() != "p1" {
			t.Errorf("Items() = %v, want a single item with product id %q", got, "p1")
		}

		createdAt := order.CreatedAt()
		if createdAt.Before(before) || createdAt.After(after) {
			t.Errorf("CreatedAt() = %v, want between %v and %v", createdAt, before, after)
		}
	})

	t.Run("empty slice returns error", func(t *testing.T) {
		order, err := domain.NewOrder([]domain.OrderItem{})
		if !errors.Is(err, domain.ErrOrderRequiresItems) {
			t.Errorf("NewOrder(empty) error = %v, want %v", err, domain.ErrOrderRequiresItems)
		}
		if order != nil {
			t.Errorf("NewOrder(empty) = %v, want nil", order)
		}
	})

	t.Run("nil slice returns error", func(t *testing.T) {
		order, err := domain.NewOrder(nil)
		if !errors.Is(err, domain.ErrOrderRequiresItems) {
			t.Errorf("NewOrder(nil) error = %v, want %v", err, domain.ErrOrderRequiresItems)
		}
		if order != nil {
			t.Errorf("NewOrder(nil) = %v, want nil", order)
		}
	})

	t.Run("merges items with same product id", func(t *testing.T) {
		items := []domain.OrderItem{
			makeItem(t, "p1", "Widget", 1000, 2),
			makeItem(t, "p1", "Widget", 1000, 3),
		}

		order, err := domain.NewOrder(items)
		if err != nil {
			t.Fatalf("NewOrder(...) returned unexpected error: %v", err)
		}

		got := order.Items()
		if len(got) != 1 {
			t.Fatalf("Items() returned %d items, want 1", len(got))
		}
		if got[0].Quantity() != 5 {
			t.Errorf("merged item quantity = %d, want %d", got[0].Quantity(), 5)
		}
	})

	t.Run("merge quantity overflow returns error", func(t *testing.T) {
		items := []domain.OrderItem{
			makeItem(t, "p1", "Widget", 1000, math.MaxInt),
			makeItem(t, "p1", "Widget", 1000, math.MaxInt),
		}

		order, err := domain.NewOrder(items)
		if !errors.Is(err, domain.ErrInvalidQuantity) {
			t.Errorf("NewOrder(...) error = %v, want %v", err, domain.ErrInvalidQuantity)
		}
		if order != nil {
			t.Errorf("NewOrder(...) = %v, want nil", order)
		}
	})

	t.Run("preserves order of distinct product ids", func(t *testing.T) {
		items := []domain.OrderItem{
			makeItem(t, "p1", "Widget", 1000, 1),
			makeItem(t, "p2", "Gadget", 2000, 1),
			makeItem(t, "p3", "Gizmo", 3000, 1),
		}

		order, err := domain.NewOrder(items)
		if err != nil {
			t.Fatalf("NewOrder(...) returned unexpected error: %v", err)
		}

		got := order.Items()
		wantIDs := []string{"p1", "p2", "p3"}
		if len(got) != len(wantIDs) {
			t.Fatalf("Items() returned %d items, want %d", len(got), len(wantIDs))
		}
		for i, want := range wantIDs {
			if got[i].ProductID() != want {
				t.Errorf("Items()[%d].ProductID() = %q, want %q", i, got[i].ProductID(), want)
			}
		}
	})
}

func TestHydrateOrder(t *testing.T) {
	createdAt := time.Date(2024, time.January, 10, 12, 0, 0, 0, time.UTC)

	t.Run("hydrates successfully", func(t *testing.T) {
		items := []domain.OrderItem{makeItem(t, "p1", "Widget", 1000, 2)}

		order, err := domain.HydrateOrder("order-1", items, domain.StatusPaid, createdAt)
		if err != nil {
			t.Fatalf("HydrateOrder(...) returned unexpected error: %v", err)
		}
		if order.ID() != domain.OrderID("order-1") {
			t.Errorf("ID() = %v, want %v", order.ID(), domain.OrderID("order-1"))
		}
		if order.Status() != domain.StatusPaid {
			t.Errorf("Status() = %v, want %v", order.Status(), domain.StatusPaid)
		}
		if !order.CreatedAt().Equal(createdAt) {
			t.Errorf("CreatedAt() = %v, want %v", order.CreatedAt(), createdAt)
		}
		if got := order.Items(); len(got) != 1 || got[0].ProductID() != "p1" {
			t.Errorf("Items() = %v, want a single item with product id %q", got, "p1")
		}
	})

	t.Run("does not merge duplicate product ids", func(t *testing.T) {
		items := []domain.OrderItem{
			makeItem(t, "p1", "Widget", 1000, 2),
			makeItem(t, "p1", "Widget", 1000, 3),
		}

		order, err := domain.HydrateOrder("order-1", items, domain.StatusPending, createdAt)
		if err != nil {
			t.Fatalf("HydrateOrder(...) returned unexpected error: %v", err)
		}
		if got := order.Items(); len(got) != 2 {
			t.Errorf("Items() returned %d items, want 2 (HydrateOrder must not merge)", len(got))
		}
	})

	t.Run("returns defensive copy of items", func(t *testing.T) {
		items := []domain.OrderItem{makeItem(t, "p1", "Widget", 1000, 2)}

		order, err := domain.HydrateOrder("order-1", items, domain.StatusPending, createdAt)
		if err != nil {
			t.Fatalf("HydrateOrder(...) returned unexpected error: %v", err)
		}

		items[0] = domain.OrderItem{}
		if got := order.Items(); got[0] == (domain.OrderItem{}) {
			t.Errorf("HydrateOrder aliased the input items slice")
		}
	})

	t.Run("empty id returns error", func(t *testing.T) {
		items := []domain.OrderItem{makeItem(t, "p1", "Widget", 1000, 1)}

		order, err := domain.HydrateOrder("", items, domain.StatusPending, createdAt)
		if !errors.Is(err, domain.ErrEmptyOrderID) {
			t.Errorf("HydrateOrder(...) error = %v, want %v", err, domain.ErrEmptyOrderID)
		}
		if order != nil {
			t.Errorf("HydrateOrder(...) = %v, want nil", order)
		}
	})

	t.Run("empty items returns error", func(t *testing.T) {
		order, err := domain.HydrateOrder("order-1", nil, domain.StatusPending, createdAt)
		if !errors.Is(err, domain.ErrOrderRequiresItems) {
			t.Errorf("HydrateOrder(...) error = %v, want %v", err, domain.ErrOrderRequiresItems)
		}
		if order != nil {
			t.Errorf("HydrateOrder(...) = %v, want nil", order)
		}
	})

	t.Run("invalid status returns error", func(t *testing.T) {
		items := []domain.OrderItem{makeItem(t, "p1", "Widget", 1000, 1)}

		order, err := domain.HydrateOrder("order-1", items, domain.OrderStatus("bogus"), createdAt)
		if !errors.Is(err, domain.ErrInvalidStatus) {
			t.Errorf("HydrateOrder(...) error = %v, want %v", err, domain.ErrInvalidStatus)
		}
		if order != nil {
			t.Errorf("HydrateOrder(...) = %v, want nil", order)
		}
	})
}

func TestOrder_ItemsReturnsDefensiveCopy(t *testing.T) {
	order := makeOrder(t)

	items := order.Items()
	items[0] = domain.OrderItem{}

	again := order.Items()
	if again[0] == (domain.OrderItem{}) {
		t.Errorf("Items() returned a slice that aliases internal state")
	}
}

func TestOrder_Total(t *testing.T) {
	t.Run("single item", func(t *testing.T) {
		item := makeItem(t, "p1", "Widget", 1000, 3)

		order, err := domain.NewOrder([]domain.OrderItem{item})
		if err != nil {
			t.Fatalf("NewOrder(...) returned unexpected error: %v", err)
		}

		if got, want := order.Total().Cents(), item.Total().Cents(); got != want {
			t.Errorf("Total() = %d cents, want %d cents", got, want)
		}
	})

	t.Run("multiple items", func(t *testing.T) {
		item1 := makeItem(t, "p1", "Widget", 1000, 2)
		item2 := makeItem(t, "p2", "Gadget", 500, 3)

		order, err := domain.NewOrder([]domain.OrderItem{item1, item2})
		if err != nil {
			t.Fatalf("NewOrder(...) returned unexpected error: %v", err)
		}

		want := item1.Total().Cents() + item2.Total().Cents()
		if got := order.Total().Cents(); got != want {
			t.Errorf("Total() = %d cents, want %d cents", got, want)
		}
	})
}

func TestOrder_MarkAsPaid(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(t *testing.T, order *domain.Order)
		wantErr    error
		wantStatus domain.OrderStatus
	}{
		{
			name:       "from pending succeeds",
			setup:      func(t *testing.T, order *domain.Order) {},
			wantErr:    nil,
			wantStatus: domain.StatusPaid,
		},
		{
			name: "from paid fails",
			setup: func(t *testing.T, order *domain.Order) {
				if err := order.MarkAsPaid(); err != nil {
					t.Fatalf("MarkAsPaid() setup failed: %v", err)
				}
			},
			wantErr:    domain.ErrInvalidStatusTransition,
			wantStatus: domain.StatusPaid,
		},
		{
			name: "from shipped fails",
			setup: func(t *testing.T, order *domain.Order) {
				if err := order.MarkAsPaid(); err != nil {
					t.Fatalf("MarkAsPaid() setup failed: %v", err)
				}
				if err := order.Ship(); err != nil {
					t.Fatalf("Ship() setup failed: %v", err)
				}
			},
			wantErr:    domain.ErrInvalidStatusTransition,
			wantStatus: domain.StatusShipped,
		},
		{
			name: "from cancelled fails",
			setup: func(t *testing.T, order *domain.Order) {
				if err := order.Cancel(); err != nil {
					t.Fatalf("Cancel() setup failed: %v", err)
				}
			},
			wantErr:    domain.ErrInvalidStatusTransition,
			wantStatus: domain.StatusCancelled,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			order := makeOrder(t)
			tt.setup(t, order)

			err := order.MarkAsPaid()
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("MarkAsPaid() error = %v, want %v", err, tt.wantErr)
			}
			if got := order.Status(); got != tt.wantStatus {
				t.Errorf("Status() = %v, want %v", got, tt.wantStatus)
			}
		})
	}
}

func TestOrder_Cancel(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(t *testing.T, order *domain.Order)
		wantErr    error
		wantStatus domain.OrderStatus
	}{
		{
			name:       "from pending succeeds",
			setup:      func(t *testing.T, order *domain.Order) {},
			wantErr:    nil,
			wantStatus: domain.StatusCancelled,
		},
		{
			name: "from paid succeeds",
			setup: func(t *testing.T, order *domain.Order) {
				if err := order.MarkAsPaid(); err != nil {
					t.Fatalf("MarkAsPaid() setup failed: %v", err)
				}
			},
			wantErr:    nil,
			wantStatus: domain.StatusCancelled,
		},
		{
			name: "from shipped fails",
			setup: func(t *testing.T, order *domain.Order) {
				if err := order.MarkAsPaid(); err != nil {
					t.Fatalf("MarkAsPaid() setup failed: %v", err)
				}
				if err := order.Ship(); err != nil {
					t.Fatalf("Ship() setup failed: %v", err)
				}
			},
			wantErr:    domain.ErrInvalidStatusTransition,
			wantStatus: domain.StatusShipped,
		},
		{
			name: "from cancelled fails",
			setup: func(t *testing.T, order *domain.Order) {
				if err := order.Cancel(); err != nil {
					t.Fatalf("Cancel() setup failed: %v", err)
				}
			},
			wantErr:    domain.ErrInvalidStatusTransition,
			wantStatus: domain.StatusCancelled,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			order := makeOrder(t)
			tt.setup(t, order)

			err := order.Cancel()
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Cancel() error = %v, want %v", err, tt.wantErr)
			}
			if got := order.Status(); got != tt.wantStatus {
				t.Errorf("Status() = %v, want %v", got, tt.wantStatus)
			}
		})
	}
}

func TestOrder_Ship(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(t *testing.T, order *domain.Order)
		wantErr    error
		wantStatus domain.OrderStatus
	}{
		{
			name:       "from pending fails",
			setup:      func(t *testing.T, order *domain.Order) {},
			wantErr:    domain.ErrInvalidStatusTransition,
			wantStatus: domain.StatusPending,
		},
		{
			name: "from paid succeeds",
			setup: func(t *testing.T, order *domain.Order) {
				if err := order.MarkAsPaid(); err != nil {
					t.Fatalf("MarkAsPaid() setup failed: %v", err)
				}
			},
			wantErr:    nil,
			wantStatus: domain.StatusShipped,
		},
		{
			name: "from shipped fails",
			setup: func(t *testing.T, order *domain.Order) {
				if err := order.MarkAsPaid(); err != nil {
					t.Fatalf("MarkAsPaid() setup failed: %v", err)
				}
				if err := order.Ship(); err != nil {
					t.Fatalf("Ship() setup failed: %v", err)
				}
			},
			wantErr:    domain.ErrInvalidStatusTransition,
			wantStatus: domain.StatusShipped,
		},
		{
			name: "from cancelled fails",
			setup: func(t *testing.T, order *domain.Order) {
				if err := order.Cancel(); err != nil {
					t.Fatalf("Cancel() setup failed: %v", err)
				}
			},
			wantErr:    domain.ErrInvalidStatusTransition,
			wantStatus: domain.StatusCancelled,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			order := makeOrder(t)
			tt.setup(t, order)

			err := order.Ship()
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Ship() error = %v, want %v", err, tt.wantErr)
			}
			if got := order.Status(); got != tt.wantStatus {
				t.Errorf("Status() = %v, want %v", got, tt.wantStatus)
			}
		})
	}
}

func TestOrder_Getters(t *testing.T) {
	item := makeItem(t, "p1", "Widget", 1000, 2)

	order, err := domain.NewOrder([]domain.OrderItem{item})
	if err != nil {
		t.Fatalf("NewOrder(...) returned unexpected error: %v", err)
	}

	if order.ID() == "" {
		t.Errorf("ID() = %q, want non-empty", order.ID())
	}
	if got := order.Status(); got != domain.StatusPending {
		t.Errorf("Status() = %v, want %v", got, domain.StatusPending)
	}
	if got := order.Items(); len(got) != 1 {
		t.Errorf("Items() returned %d items, want %d", len(got), 1)
	}
	if order.CreatedAt().IsZero() {
		t.Errorf("CreatedAt() = zero value, want non-zero")
	}
	if got, want := order.Total().Cents(), item.Total().Cents(); got != want {
		t.Errorf("Total() = %d cents, want %d cents", got, want)
	}
}
