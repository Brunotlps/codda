package postgres

import (
	"time"

	"github.com/Brunotlps/codda/internal/domain"
)

// orderRow carries a single orders row as scanned from the database.
type orderRow struct {
	ID        string
	Status    string
	CreatedAt time.Time
}

// orderItemRow carries a single order_items row as scanned from the
// database.
type orderItemRow struct {
	OrderID     string
	Position    int
	ProductID   string
	ProductName string
	PriceCents  int64
	Quantity    int
}

// rowsToOrder reconstructs a domain.Order from an orderRow and its
// associated itemRows. itemRows must already be ordered by position. It
// returns an error if any row holds data that violates a domain invariant.
func rowsToOrder(orderRow orderRow, itemRows []orderItemRow) (*domain.Order, error) {
	items := make([]domain.OrderItem, len(itemRows))
	for i, row := range itemRows {
		price, err := domain.NewMoney(row.PriceCents)
		if err != nil {
			return nil, err
		}

		item, err := domain.NewOrderItem(row.ProductID, row.ProductName, price, row.Quantity)
		if err != nil {
			return nil, err
		}

		items[i] = item
	}

	return domain.HydrateOrder(
		domain.OrderID(orderRow.ID),
		items,
		domain.OrderStatus(orderRow.Status),
		orderRow.CreatedAt,
	)
}
