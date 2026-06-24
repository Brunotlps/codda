package application_test

import (
	"testing"

	"github.com/Brunotlps/codda/internal/domain"
)

// makeOrder builds a domain.Order with a single item, for test setup,
// failing the test immediately if construction fails.
func makeOrder(t *testing.T) *domain.Order {
	t.Helper()

	price, err := domain.NewMoney(1000)
	if err != nil {
		t.Fatalf("NewMoney(...) failed during test setup: %v", err)
	}

	item, err := domain.NewOrderItem("p1", "Widget", price, 1)
	if err != nil {
		t.Fatalf("NewOrderItem(...) failed during test setup: %v", err)
	}

	order, err := domain.NewOrder([]domain.OrderItem{item})
	if err != nil {
		t.Fatalf("NewOrder(...) failed during test setup: %v", err)
	}

	return order
}
