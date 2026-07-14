package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Brunotlps/codda/internal/application"
	"github.com/Brunotlps/codda/internal/domain"
)

// OrderRepository implements application.OrderRepository against PostgreSQL.
type OrderRepository struct {
	pool *pgxpool.Pool
}

// NewOrderRepository creates an OrderRepository backed by pool.
func NewOrderRepository(pool *pgxpool.Pool) *OrderRepository {
	return &OrderRepository{pool: pool}
}

// Save persists order using UPSERT on orders and a full replace on
// order_items, all within a single transaction.
func (r *OrderRepository) Save(ctx context.Context, order *domain.Order) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	const upsertOrder = `
		INSERT INTO orders (id, status, created_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (id) DO UPDATE SET
			status = EXCLUDED.status`

	if _, err := tx.Exec(ctx, upsertOrder, string(order.ID()), string(order.Status()), order.CreatedAt()); err != nil {
		return fmt.Errorf("upsert order: %w", err)
	}

	const deleteItems = `DELETE FROM order_items WHERE order_id = $1`
	if _, err := tx.Exec(ctx, deleteItems, string(order.ID())); err != nil {
		return fmt.Errorf("delete order items: %w", err)
	}

	const insertItem = `
		INSERT INTO order_items (order_id, position, product_id, product_name, price_cents, quantity)
		VALUES ($1, $2, $3, $4, $5, $6)`

	for i, item := range order.Items() {
		_, err := tx.Exec(ctx, insertItem,
			string(order.ID()),
			i, // Use the iteration index as the table position. Items() preserves the original order (guaranteed by the domain), so the database position matches the original ordering.
			item.ProductID(),
			item.ProductName(),
			item.Price().Cents(),
			item.Quantity(),
		)
		if err != nil {
			return fmt.Errorf("insert item at position %d: %w", i, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

// FindByID retrieves the order with the given ID, reconstructing the domain
// aggregate via rowsToOrder. It returns application.ErrOrderNotFound if no
// order with that ID exists.
func (r *OrderRepository) FindByID(ctx context.Context, id domain.OrderID) (*domain.Order, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	const selectOrder = `SELECT id, status, created_at FROM orders WHERE id = $1`

	var row orderRow
	err := r.pool.QueryRow(ctx, selectOrder, string(id)).Scan(&row.ID, &row.Status, &row.CreatedAt)
	if err != nil {
		// Convert persistence-specific errors into application-level errors.
		// This prevents driver details from leaking into the use case and
		// preserves the application boundary.
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, application.ErrOrderNotFound
		}
		return nil, fmt.Errorf("query order: %w", err)
	}

	const selectItems = `
		SELECT order_id, position, product_id, product_name, price_cents, quantity
		FROM order_items
		WHERE order_id = $1
		ORDER BY position ASC`

	itemRows, err := r.pool.Query(ctx, selectItems, string(id))
	if err != nil {
		return nil, fmt.Errorf("query order items: %w", err)
	}
	defer itemRows.Close()

	var items []orderItemRow
	for itemRows.Next() {
		var item orderItemRow
		if err := itemRows.Scan(
			&item.OrderID, &item.Position,
			&item.ProductID, &item.ProductName,
			&item.PriceCents, &item.Quantity,
		); err != nil {
			return nil, fmt.Errorf("scan order item: %w", err)
		}
		items = append(items, item)
	}
	if err := itemRows.Err(); err != nil {
		return nil, fmt.Errorf("iterate order items: %w", err)
	}

	return rowsToOrder(row, items)
}

// ListResumes returns the page of order summaries selected by pagination from
// those matching filters, along with whether further pages exist. Resumes are sorted by
// createdAt descending, with ID descending as a tiebreaker.
func (r *OrderRepository) ListResumes(ctx context.Context, filters application.ListOrdersFilters, pagination application.Pagination) ([]application.OrderResume, bool, error) {
	if err := ctx.Err(); err != nil {
		return nil, false, err
	}

	var conditions []string
	var args []any
	n := 1

	if filters.Status != nil {
		conditions = append(conditions, fmt.Sprintf("o.status = $%d", n))
		args = append(args, string(*filters.Status))
		n++
	}
	if filters.CreatedFrom != nil {
		conditions = append(conditions, fmt.Sprintf("o.created_at >= $%d", n))
		args = append(args, *filters.CreatedFrom)
		n++
	}
	if filters.CreatedTo != nil {
		conditions = append(conditions, fmt.Sprintf("o.created_at <= $%d", n))
		args = append(args, *filters.CreatedTo)
		n++
	}
	if filters.PriceMin != nil {
		conditions = append(conditions, fmt.Sprintf("COALESCE(ot.total_cents, 0) >= $%d", n))
		args = append(args, filters.PriceMin.Cents())
		n++
	}
	if filters.PriceMax != nil {
		conditions = append(conditions, fmt.Sprintf("COALESCE(ot.total_cents, 0) <= $%d", n))
		args = append(args, filters.PriceMax.Cents())
		n++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Request limit+1 orders in the subquery: if we get limit+1 back, there
	// is a next page.
	fetchLimit := pagination.Limit + 1
	args = append(args, fetchLimit, pagination.Offset)
	limitN := fmt.Sprintf("$%d", n)
	offsetN := fmt.Sprintf("$%d", n+1)

	query := fmt.Sprintf(`
		SELECT o.id, o.status, o.created_at, COALESCE(ot.total_cents, 0) AS total_cents
		FROM orders o
		LEFT JOIN (
			SELECT order_id, SUM(price_cents * quantity) AS total_cents
			FROM order_items
			GROUP BY order_id
		) ot ON ot.order_id = o.id
		%s
		ORDER BY o.created_at DESC, o.id DESC
		LIMIT %s OFFSET %s`,
		whereClause, limitN, offsetN)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, false, fmt.Errorf("list order resumes: %w", err)
	}
	defer rows.Close()

	var resumes []application.OrderResume

	for rows.Next() {
		var id string
		var status string
		var createdAt time.Time
		var totalCents int64

		if err := rows.Scan(&id, &status, &createdAt, &totalCents); err != nil {
			return nil, false, fmt.Errorf("scan list resume row: %w", err)
		}

		orderStatus := domain.OrderStatus(status)
		if !orderStatus.IsValid() {
			return nil, false, domain.ErrInvalidStatus
		}

		total, err := domain.NewMoney(totalCents)
		if err != nil {
			return nil, false, err
		}

		resumes = append(resumes, application.OrderResume{
			ID:        domain.OrderID(id),
			Status:    orderStatus,
			Total:     total,
			CreatedAt: createdAt,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, false, fmt.Errorf("iterate list resume rows: %w", err)
	}

	hasMore := len(resumes) > pagination.Limit
	if hasMore {
		resumes = resumes[:pagination.Limit]
	}

	return resumes, hasMore, nil
}
