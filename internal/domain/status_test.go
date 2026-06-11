package domain_test

import (
	"testing"

	"github.com/Brunotlps/codda/internal/domain"
)

func TestOrderStatusIsValid(t *testing.T) {
	tests := []struct {
		name   string
		status domain.OrderStatus
		want   bool
	}{
		{name: "pending", status: domain.StatusPending, want: true},
		{name: "paid", status: domain.StatusPaid, want: true},
		{name: "shipped", status: domain.StatusShipped, want: true},
		{name: "cancelled", status: domain.StatusCancelled, want: true},
		{name: "unknown", status: domain.OrderStatus("unknown"), want: false},
		{name: "empty", status: domain.OrderStatus(""), want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.status.IsValid(); got != tt.want {
				t.Errorf("%q.IsValid() = %v, want %v", tt.status, got, tt.want)
			}
		})
	}
}

func TestOrderStatusIsTerminal(t *testing.T) {
	tests := []struct {
		name   string
		status domain.OrderStatus
		want   bool
	}{
		{name: "pending", status: domain.StatusPending, want: false},
		{name: "paid", status: domain.StatusPaid, want: false},
		{name: "shipped", status: domain.StatusShipped, want: true},
		{name: "cancelled", status: domain.StatusCancelled, want: true},
		{name: "unknown", status: domain.OrderStatus("unknown"), want: false},
		{name: "empty", status: domain.OrderStatus(""), want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.status.IsTerminal(); got != tt.want {
				t.Errorf("%q.IsTerminal() = %v, want %v", tt.status, got, tt.want)
			}
		})
	}
}

func TestOrderStatusCanTransitionTo(t *testing.T) {
	tests := []struct {
		name string
		from domain.OrderStatus
		to   domain.OrderStatus
		want bool
	}{
		// from pending
		{name: "pending to pending", from: domain.StatusPending, to: domain.StatusPending, want: false},
		{name: "pending to paid", from: domain.StatusPending, to: domain.StatusPaid, want: true},
		{name: "pending to shipped", from: domain.StatusPending, to: domain.StatusShipped, want: false},
		{name: "pending to cancelled", from: domain.StatusPending, to: domain.StatusCancelled, want: true},

		// from paid
		{name: "paid to pending", from: domain.StatusPaid, to: domain.StatusPending, want: false},
		{name: "paid to paid", from: domain.StatusPaid, to: domain.StatusPaid, want: false},
		{name: "paid to shipped", from: domain.StatusPaid, to: domain.StatusShipped, want: true},
		{name: "paid to cancelled", from: domain.StatusPaid, to: domain.StatusCancelled, want: true},

		// from shipped (terminal — rejects everything)
		{name: "shipped to pending", from: domain.StatusShipped, to: domain.StatusPending, want: false},
		{name: "shipped to paid", from: domain.StatusShipped, to: domain.StatusPaid, want: false},
		{name: "shipped to shipped", from: domain.StatusShipped, to: domain.StatusShipped, want: false},
		{name: "shipped to cancelled", from: domain.StatusShipped, to: domain.StatusCancelled, want: false},

		// from cancelled (terminal — rejects everything)
		{name: "cancelled to pending", from: domain.StatusCancelled, to: domain.StatusPending, want: false},
		{name: "cancelled to paid", from: domain.StatusCancelled, to: domain.StatusPaid, want: false},
		{name: "cancelled to shipped", from: domain.StatusCancelled, to: domain.StatusShipped, want: false},
		{name: "cancelled to cancelled", from: domain.StatusCancelled, to: domain.StatusCancelled, want: false},

		// invalid origin
		{name: "unknown to paid", from: domain.OrderStatus("unknown"), to: domain.StatusPaid, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.from.CanTransitionTo(tt.to); got != tt.want {
				t.Errorf("%q.CanTransitionTo(%q) = %v, want %v", tt.from, tt.to, got, tt.want)
			}
		})
	}
}

func TestAllValidStatus(t *testing.T) {
	statuses := domain.AllValidStatus()

	want := []domain.OrderStatus{
		domain.StatusPending,
		domain.StatusPaid,
		domain.StatusShipped,
		domain.StatusCancelled,
	}

	if len(statuses) != len(want) {
		t.Fatalf("AllValidStatus() returned %d statuses, want %d", len(statuses), len(want))
	}

	for i, status := range statuses {
		if status != want[i] {
			t.Errorf("AllValidStatus()[%d] = %q, want %q", i, status, want[i])
		}
		if !status.IsValid() {
			t.Errorf("AllValidStatus()[%d] = %q is not reported as valid", i, status)
		}
	}
}
