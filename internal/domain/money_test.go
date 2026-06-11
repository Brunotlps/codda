package domain_test

import (
	"errors"
	"testing"

	"github.com/Brunotlps/codda/internal/domain"
)

func TestNewMoney(t *testing.T) {
	tests := []struct {
		name      string
		cents     int64
		wantCents int64
		wantErr   error
	}{
		{name: "positive cents", cents: 1999, wantCents: 1999, wantErr: nil},
		{name: "zero is valid", cents: 0, wantCents: 0, wantErr: nil},
		{name: "negative is rejected", cents: -1, wantCents: 0, wantErr: domain.ErrNegativeMoney},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := domain.NewMoney(tt.cents)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("NewMoney(%d) error = %v, want %v", tt.cents, err, tt.wantErr)
			}
			if got.Cents() != tt.wantCents {
				t.Errorf("NewMoney(%d) = %d cents, want %d cents", tt.cents, got.Cents(), tt.wantCents)
			}
		})
	}
}

func TestMoney_Cents(t *testing.T) {
	m, err := domain.NewMoney(1234)
	if err != nil {
		t.Fatalf("NewMoney(1234) returned unexpected error: %v", err)
	}

	if got := m.Cents(); got != 1234 {
		t.Errorf("Cents() = %d, want %d", got, 1234)
	}
}

func TestMoney_Add(t *testing.T) {
	tests := []struct {
		name      string
		a, b      int64
		wantCents int64
	}{
		{name: "sum of two positives", a: 100, b: 250, wantCents: 350},
		{name: "adding zero is identity", a: 999, b: 0, wantCents: 999},
		{name: "zero plus zero", a: 0, b: 0, wantCents: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a, err := domain.NewMoney(tt.a)
			if err != nil {
				t.Fatalf("NewMoney(%d) returned unexpected error: %v", tt.a, err)
			}
			b, err := domain.NewMoney(tt.b)
			if err != nil {
				t.Fatalf("NewMoney(%d) returned unexpected error: %v", tt.b, err)
			}

			got := a.Add(b)
			if got.Cents() != tt.wantCents {
				t.Errorf("(%d).Add(%d) = %d cents, want %d cents", tt.a, tt.b, got.Cents(), tt.wantCents)
			}
		})
	}
}

func TestMoney_AddDoesNotMutate(t *testing.T) {
	original, err := domain.NewMoney(100)
	if err != nil {
		t.Fatalf("NewMoney(100) returned unexpected error: %v", err)
	}
	other, err := domain.NewMoney(50)
	if err != nil {
		t.Fatalf("NewMoney(50) returned unexpected error: %v", err)
	}

	_ = original.Add(other)

	if got := original.Cents(); got != 100 {
		t.Errorf("original was mutated by Add: got %d cents, want 100", got)
	}
}

func TestMoney_Multiply(t *testing.T) {
	tests := []struct {
		name      string
		cents     int64
		factor    int
		wantCents int64
	}{
		{name: "multiply by positive factor", cents: 250, factor: 4, wantCents: 1000},
		{name: "multiply by zero", cents: 999, factor: 0, wantCents: 0},
		{name: "multiply by one is identity", cents: 1999, factor: 1, wantCents: 1999},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := domain.NewMoney(tt.cents)
			if err != nil {
				t.Fatalf("NewMoney(%d) returned unexpected error: %v", tt.cents, err)
			}

			got := m.Multiply(tt.factor)
			if got.Cents() != tt.wantCents {
				t.Errorf("(%d).Multiply(%d) = %d cents, want %d cents", tt.cents, tt.factor, got.Cents(), tt.wantCents)
			}
		})
	}
}

func TestMoney_MultiplyDoesNotMutate(t *testing.T) {
	original, err := domain.NewMoney(250)
	if err != nil {
		t.Fatalf("NewMoney(250) returned unexpected error: %v", err)
	}

	_ = original.Multiply(4)

	if got := original.Cents(); got != 250 {
		t.Errorf("original was mutated by Multiply: got %d cents, want 250", got)
	}
}

func TestMoney_String(t *testing.T) {
	tests := []struct {
		name  string
		cents int64
		want  string
	}{
		{name: "reais and centavos", cents: 1999, want: "R$ 19,99"},
		{name: "centavos padding", cents: 1905, want: "R$ 19,05"},
		{name: "zero", cents: 0, want: "R$ 0,00"},
		{name: "reais without centavos", cents: 1900, want: "R$ 19,00"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := domain.NewMoney(tt.cents)
			if err != nil {
				t.Fatalf("NewMoney(%d) returned unexpected error: %v", tt.cents, err)
			}

			if got := m.String(); got != tt.want {
				t.Errorf("(%d).String() = %q, want %q", tt.cents, got, tt.want)
			}
		})
	}
}
