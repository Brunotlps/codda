package domain

// OrderStatus represents the state of an order within its lifecycle.
type OrderStatus string

// Order status constants enumerate the lifecycle states a pedido may have.
const (
	StatusPending   OrderStatus = "pending"
	StatusPaid      OrderStatus = "paid"
	StatusShipped   OrderStatus = "shipped"
	StatusCancelled OrderStatus = "cancelled"
)

// IsValid reports whether s is one of the known order states.
func (s OrderStatus) IsValid() bool {
	switch s {
	case StatusPending, StatusPaid, StatusShipped, StatusCancelled:
		return true
	default:
		return false
	}
}

// IsTerminal reports whether s is a final state, from which no transitions are allowed.
func (s OrderStatus) IsTerminal() bool {
	switch s {
	case StatusShipped, StatusCancelled:
		return true
	default:
		return false
	}
}

// CanTransitionTo reports whether an order may move from state s to target,
// according to the transitions allowed by the business rules. Self-transitions
// (s == target) and transitions from terminal states return false.
func (s OrderStatus) CanTransitionTo(target OrderStatus) bool {
	switch s {
	case StatusPending:
		return target == StatusPaid || target == StatusCancelled
	case StatusPaid:
		return target == StatusShipped || target == StatusCancelled
	default:
		return false
	}
}

// AllValidStatus returns every valid order status, in lifecycle order.
// Useful for input validation, API documentation, and exhaustive testing.
func AllValidStatus() []OrderStatus {
	return []OrderStatus{
		StatusPending,
		StatusPaid,
		StatusShipped,
		StatusCancelled,
	}
}
