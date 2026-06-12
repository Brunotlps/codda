package domain_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/Brunotlps/codda/internal/domain"
)

// makeMoney builds a domain.Money for test setup, failing the test
// immediately if construction fails.
func makeMoney(t *testing.T, cents int64) domain.Money {
	t.Helper()

	m, err := domain.NewMoney(cents)
	if err != nil {
		t.Fatalf("NewMoney(%d) failed during test setup: %v", cents, err)
	}

	return m
}

func TestNewOrderItem(t *testing.T) {
	tests := []struct {
		name        string
		productID   string
		productName string
		price       domain.Money
		quantity    int
		wantErr     error
	}{
		{
			name:        "valid item",
			productID:   "prod-1",
			productName: "Widget",
			price:       makeMoney(t, 1000),
			quantity:    2,
			wantErr:     nil,
		},
		{
			name:        "empty product id",
			productID:   "",
			productName: "Widget",
			price:       makeMoney(t, 1000),
			quantity:    2,
			wantErr:     domain.ErrEmptyProductID,
		},
		{
			name:        "empty product name",
			productID:   "prod-1",
			productName: "",
			price:       makeMoney(t, 1000),
			quantity:    2,
			wantErr:     domain.ErrEmptyProductName,
		},
		{
			name:        "product name too long",
			productID:   "prod-1",
			productName: strings.Repeat("a", domain.MaxProductNameLength+1),
			price:       makeMoney(t, 1000),
			quantity:    2,
			wantErr:     domain.ErrProductNameTooLong,
		},
		{
			name:        "zero price",
			productID:   "prod-1",
			productName: "Widget",
			price:       domain.Money{},
			quantity:    2,
			wantErr:     domain.ErrInvalidPrice,
		},
		{
			name:        "zero quantity",
			productID:   "prod-1",
			productName: "Widget",
			price:       makeMoney(t, 1000),
			quantity:    0,
			wantErr:     domain.ErrInvalidQuantity,
		},
		{
			name:        "negative quantity",
			productID:   "prod-1",
			productName: "Widget",
			price:       makeMoney(t, 1000),
			quantity:    -1,
			wantErr:     domain.ErrInvalidQuantity,
		},
		{
			name:        "product name at max length",
			productID:   "prod-1",
			productName: strings.Repeat("a", domain.MaxProductNameLength),
			price:       makeMoney(t, 1000),
			quantity:    2,
			wantErr:     nil,
		},
		{
			name:        "quantity of one",
			productID:   "prod-1",
			productName: "Widget",
			price:       makeMoney(t, 1000),
			quantity:    1,
			wantErr:     nil,
		},
		{
			name:        "product name with accents within rune limit",
			productID:   "prod-1",
			productName: strings.Repeat("ã", domain.MaxProductNameLength),
			price:       makeMoney(t, 1000),
			quantity:    1,
			wantErr:     nil,
		},
		{
			name:        "product name with accents exceeds rune limit",
			productID:   "prod-1",
			productName: strings.Repeat("ã", domain.MaxProductNameLength+1),
			price:       makeMoney(t, 1000),
			quantity:    1,
			wantErr:     domain.ErrProductNameTooLong,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := domain.NewOrderItem(tt.productID, tt.productName, tt.price, tt.quantity)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("NewOrderItem(...) error = %v, want %v", err, tt.wantErr)
			}

			if tt.wantErr != nil {
				if got != (domain.OrderItem{}) {
					t.Errorf("NewOrderItem(...) = %+v, want zero value", got)
				}
				return
			}

			if got.ProductID() != tt.productID {
				t.Errorf("ProductID() = %q, want %q", got.ProductID(), tt.productID)
			}
			if got.ProductName() != tt.productName {
				t.Errorf("ProductName() = %q, want %q", got.ProductName(), tt.productName)
			}
			if got.Price().Cents() != tt.price.Cents() {
				t.Errorf("Price().Cents() = %d, want %d", got.Price().Cents(), tt.price.Cents())
			}
			if got.Quantity() != tt.quantity {
				t.Errorf("Quantity() = %d, want %d", got.Quantity(), tt.quantity)
			}
		})
	}
}

func TestOrderItem_Getters(t *testing.T) {
	price := makeMoney(t, 1500)

	item, err := domain.NewOrderItem("prod-1", "Widget", price, 3)
	if err != nil {
		t.Fatalf("NewOrderItem(...) returned unexpected error: %v", err)
	}

	if got := item.ProductID(); got != "prod-1" {
		t.Errorf("ProductID() = %q, want %q", got, "prod-1")
	}
	if got := item.ProductName(); got != "Widget" {
		t.Errorf("ProductName() = %q, want %q", got, "Widget")
	}
	if got := item.Price().Cents(); got != price.Cents() {
		t.Errorf("Price().Cents() = %d, want %d", got, price.Cents())
	}
	if got := item.Quantity(); got != 3 {
		t.Errorf("Quantity() = %d, want %d", got, 3)
	}
}

func TestOrderItem_Total(t *testing.T) {
	tests := []struct {
		name      string
		cents     int64
		quantity  int
		wantCents int64
	}{
		{name: "multiple units", cents: 1000, quantity: 3, wantCents: 3000},
		{name: "single unit equals price", cents: 1999, quantity: 1, wantCents: 1999},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item, err := domain.NewOrderItem("prod-1", "Widget", makeMoney(t, tt.cents), tt.quantity)
			if err != nil {
				t.Fatalf("NewOrderItem(...) returned unexpected error: %v", err)
			}

			if got := item.Total().Cents(); got != tt.wantCents {
				t.Errorf("Total() = %d cents, want %d cents", got, tt.wantCents)
			}
		})
	}
}

func TestOrderItem_WithQuantity(t *testing.T) {
	tests := []struct {
		name         string
		quantity     int
		wantErr      error
		wantQuantity int
	}{
		{name: "valid new quantity", quantity: 5, wantErr: nil, wantQuantity: 5},
		{name: "zero quantity", quantity: 0, wantErr: domain.ErrInvalidQuantity},
		{name: "negative quantity", quantity: -1, wantErr: domain.ErrInvalidQuantity},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original, err := domain.NewOrderItem("prod-1", "Widget", makeMoney(t, 1000), 1)
			if err != nil {
				t.Fatalf("NewOrderItem(...) returned unexpected error: %v", err)
			}

			got, err := original.WithQuantity(tt.quantity)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("WithQuantity(%d) error = %v, want %v", tt.quantity, err, tt.wantErr)
			}

			if tt.wantErr != nil {
				if got != (domain.OrderItem{}) {
					t.Errorf("WithQuantity(%d) = %+v, want zero value", tt.quantity, got)
				}
				return
			}

			if got.Quantity() != tt.wantQuantity {
				t.Errorf("WithQuantity(%d).Quantity() = %d, want %d", tt.quantity, got.Quantity(), tt.wantQuantity)
			}
		})
	}
}

func TestOrderItem_WithQuantityDoesNotMutate(t *testing.T) {
	original, err := domain.NewOrderItem("prod-1", "Widget", makeMoney(t, 1000), 2)
	if err != nil {
		t.Fatalf("NewOrderItem(...) returned unexpected error: %v", err)
	}

	if _, err := original.WithQuantity(5); err != nil {
		t.Fatalf("WithQuantity(5) returned unexpected error: %v", err)
	}

	if got := original.Quantity(); got != 2 {
		t.Errorf("original was mutated by WithQuantity: got quantity %d, want 2", got)
	}
}
