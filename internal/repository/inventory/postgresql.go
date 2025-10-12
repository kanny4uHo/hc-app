package notification

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"

	"healthcheckProject/internal/entity"
	"healthcheckProject/internal/service"
)

type PgInventoryRepo struct {
	db *sql.DB
}

var _ service.InventoryRepo = (*PgInventoryRepo)(nil)

func NewPgInventoryRepo(db *sql.DB) *PgInventoryRepo {
	return &PgInventoryRepo{db: db}
}

func (p *PgInventoryRepo) GetReservationItem(ctx context.Context, reservationID uint64) (entity.Reservation, error) {
	var reservation entity.Reservation

	err := p.db.QueryRowContext(ctx, "SELECT id, item_id, order_id, amount FROM reservations where id = $1", reservationID).
		Scan(
			&reservation.ID,
			&reservation.Item.ItemID,
			&reservation.OrderID,
			&reservation.Item.Amount,
		)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.Reservation{}, service.ErrNoReservation
		}

		return entity.Reservation{}, fmt.Errorf("failed to query reservation: %w", err)
	}

	return reservation, nil
}

func (p *PgInventoryRepo) AddItem(ctx context.Context, item entity.Item) error {
	_, err := p.db.ExecContext(
		ctx,
		"INSERT INTO items(id, description, amount) VALUES ($1, $2, $3)",
		item.ItemID,
		"",
		item.Amount,
	)

	if err != nil {
		return fmt.Errorf("could not send order paid email: %w", err)
	}

	return nil
}

func (p *PgInventoryRepo) UpdateItem(ctx context.Context, item entity.Item) error {
	_, err := p.db.ExecContext(ctx, "UPDATE items set amount = $1 WHERE id = $2 ", item.Amount, item.ItemID)
	if err != nil {
		return fmt.Errorf("could not update items: %w", err)
	}

	return nil
}

func (p *PgInventoryRepo) GetItem(ctx context.Context, itemID string, opts ...service.Opts) (entity.Item, error) {
	item := entity.Item{
		ItemID: itemID,
	}

	query := "SELECT amount FROM items WHERE id = $1"

	if len(opts) > 0 && opts[0].ToUpdate {
		query = query + " FOR UPDATE"
	}

	err := p.db.QueryRowContext(ctx, query, itemID).
		Scan(
			&item.Amount,
		)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.Item{}, service.ErrNoItemsInInventory
		}

		return item, fmt.Errorf("failed to fetch item from pg: %w", err)
	}

	return item, nil
}

func (p *PgInventoryRepo) ReserveItem(ctx context.Context, orderID uint64, item entity.Item) (entity.Reservation, error) {
	reservation := entity.Reservation{
		OrderID: orderID,
		Item:    item,
	}

	err := p.db.QueryRowContext(
		ctx,
		"INSERT INTO reservations(order_id, item_id, amount) VALUES ($1, $2, $3) returning id",
		orderID,
		item.ItemID,
		item.Amount,
	).Scan(&reservation.ID)

	if err != nil {
		return entity.Reservation{}, fmt.Errorf("could not send order paid email: %w", err)
	}

	return reservation, nil
}

func (p *PgInventoryRepo) DeleteReservationItem(ctx context.Context, itemID uint64) error {
	_, err := p.db.ExecContext(ctx, `DELETE FROM reservations WHERE id=$1`, itemID)
	if err != nil {
		return fmt.Errorf("failed to exec delete query: %w", err)
	}

	return nil
}

func (p *PgInventoryRepo) WithinTransaction(ctx context.Context, f func(ctx context.Context) error) error {
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	defer func() {
		if err != nil {
			txErr := tx.Rollback()
			if txErr != nil {
				log.Printf("failed to rollback transaction: %s", err)
			}
		} else {
			txErr := tx.Commit()
			if txErr != nil {
				log.Printf("failed to commit transaction: %s", err)
			}
		}
	}()

	if err = f(ctx); err != nil {
		return fmt.Errorf("failed to execute transaction: %w", err)
	}

	return nil
}
