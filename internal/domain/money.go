package domain

import "fmt"

// centsPerUnit is the number of cents in one unit of currency (e.g., one real).
const centsPerUnit = 100

// Money represents a monetary amount in cents, implicitly denominated in BRL.
// The zero value represents R$ 0,00 and is valid.
type Money struct {
	cents int64
}

// NewMoney creates a Money value from an amount in cents. Zero is allowed,
// since totals such as an empty order or a fully discounted item can be
// zero. NewMoney returns ErrNegativeMoney if cents is negative.
func NewMoney(cents int64) (Money, error) {
	if cents < 0 {
		return Money{}, ErrNegativeMoney
	}

	return Money{cents: cents}, nil
}

// Cents returns the amount represented by m, in cents.
func (m Money) Cents() int64 {
	return m.cents
}

// Add returns the sum of m and other as a new Money value.
func (m Money) Add(other Money) Money {
	return Money{cents: m.cents + other.cents}
}

// Multiply returns m scaled by n as a new Money value. It assumes n is
// non-negative; callers scaling by a quantity should rely on the
// quantity invariant (>= 1) being enforced beforehand.
// int64 overflow risk -> for the future
func (m Money) Multiply(n int) Money {
	return Money{cents: m.cents * int64(n)}
}

// String returns m formatted as a BRL amount, e.g. "R$ 19,99". It is intended
// for debugging and logs (the Stringer convention), not for locale-aware
// display; a presentation layer that needs other locales or currencies
// should format Cents() itself.
func (m Money) String() string {
	reais := m.cents / centsPerUnit
	centavos := m.cents % centsPerUnit

	return fmt.Sprintf("R$ %d,%02d", reais, centavos)
}
