package order

import (
	"context"
	"database/sql"
	"fmt"
	"healthcheckProject/internal/entity"
	"healthcheckProject/internal/service"
)

type PgOrderRepo struct {
	db *sql.DB
}

var _ service.OrderRepo = (*PgOrderRepo)(nil)

func NewPgOrderRepo(db *sql.DB) *PgOrderRepo {
	return &PgOrderRepo{
		db: db,
	}
}

func (p *PgOrderRepo) CreateOrder(ctx context.Context, user entity.UserShort, price uint64, item string, status entity.Status) (entity.Order, error) {
	order := entity.Order{
		Price:  price,
		Item:   item,
		Status: status,
		Owner:  user,
	}

	err := p.db.QueryRowContext(
		ctx,
		"INSERT into orders(item, price, status, user_id) VALUES ($1, $2, $3, $4) RETURNING id",
		item, price, entity.StatusCreated, user.ID,
	).Scan(&order.ID)

	if err != nil {
		return entity.Order{}, fmt.Errorf("failed to insert order: %w", err)
	}

	return order, nil
}

func (p *PgOrderRepo) SetOrderStatus(ctx context.Context, id uint64, status entity.Status) error {
	_, err := p.db.ExecContext(ctx, "UPDATE orders SET status = $1 WHERE id = $2", status, id)
	if err != nil {
		return fmt.Errorf("failed to update order: %w", err)
	}

	return nil
}

func (p *PgOrderRepo) DeleteOrder(ctx context.Context, id uint64) error {
	_, err := p.db.ExecContext(
		ctx,
		"DELETE FROM orders WHERE id = $1", id,
	)

	if err != nil {
		return fmt.Errorf("failed to delete order with ID %d: %w", id, err)
	}

	return nil
}

func (p *PgOrderRepo) GetOrderByID(ctx context.Context, id uint64) (entity.Order, error) {
	order := entity.Order{}

	err := p.db.QueryRowContext(
		ctx,
		"SELECT id, item, price, status, user_id FROM orders WHERE id = $1",
		id,
	).Scan(&order.ID, &order.Item, &order.Price, &order.Status, &order.Owner.ID)

	if err != nil {
		return entity.Order{}, fmt.Errorf("failed to get order with ID %d: %w", id, err)
	}

	return order, nil
}
