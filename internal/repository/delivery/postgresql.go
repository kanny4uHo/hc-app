package delivery

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"healthcheckProject/internal/entity"
	"healthcheckProject/internal/service"
)

type PgDeliveryRepository struct {
	db *sql.DB
}

var _ service.DeliveryRepository = (*PgDeliveryRepository)(nil)

func NewPgDeliveryRepository(db *sql.DB) *PgDeliveryRepository {
	return &PgDeliveryRepository{
		db: db,
	}
}

func (p *PgDeliveryRepository) GetDeliveryByID(ctx context.Context, id uint64) (entity.Delivery, error) {
	var delivery entity.Delivery

	err := p.db.QueryRowContext(
		ctx,
		"SELECT id, order_id, courier_id, status FROM deliveries WHERE id = $1",
		id,
	).Scan(
		&delivery.ID,
		&delivery.OrderID,
		&delivery.CourierID,
		&delivery.Status,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.Delivery{}, service.ErrDeliveryNotFound
		}

		return entity.Delivery{}, fmt.Errorf("failed to get delivery by SQL: %w", err)
	}

	return delivery, nil
}

func (p *PgDeliveryRepository) GetDeliveries(ctx context.Context) ([]entity.Delivery, error) {
	rows, err := p.db.QueryContext(ctx, "SELECT id, order_id, courier_id, status FROM deliveries")
	if err != nil {
		return nil, fmt.Errorf("failed to do SQL select: %w", err)
	}

	defer rows.Close()

	deliveries := make([]entity.Delivery, 0, 10)

	for rows.Next() {
		var delivery entity.Delivery

		scanErr := rows.Scan(&delivery.ID, &delivery.OrderID, &delivery.CourierID, &delivery.Status)
		if scanErr != nil {
			return nil, fmt.Errorf("failed to scan value: %w", scanErr)
		}

		deliveries = append(deliveries, delivery)
	}

	return deliveries, nil
}

func (p *PgDeliveryRepository) GetCouriersOnShiftWithoutActiveDeliveries(ctx context.Context) ([]entity.Courier, error) {
	rows, err := p.db.QueryContext(
		ctx,
		`SELECT c.id, c.name, c.is_on_shift
					FROM couriers c
					LEFT JOIN deliveries d ON c.id = d.courier_id AND d.status = 'on_the_way'
					WHERE c.is_on_shift = 1
					AND d.id IS NULL`,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to SQL query available couriers: %w", err)
	}

	defer rows.Close()

	couriers := make([]entity.Courier, 0, 8)

	for rows.Next() {
		var courier entity.Courier

		var shiftValue int

		scanErr := rows.Scan(
			&courier.ID,
			&courier.Name,
			&shiftValue,
		)

		if scanErr != nil {
			return nil, fmt.Errorf("failed to scan courier row: %w", err)
		}

		courier.IsOnShift = shiftValue == 1

		couriers = append(couriers, courier)
	}

	return couriers, nil
}

func (p *PgDeliveryRepository) AddDelivery(ctx context.Context, delivery entity.Delivery) (entity.Delivery, error) {
	err := p.db.QueryRowContext(
		ctx,
		`INSERT into deliveries (order_id, courier_id, status) VALUES ($1, $2, $3) RETURNING id`,
		delivery.OrderID,
		delivery.CourierID,
		delivery.Status,
	).Scan(&delivery.ID)

	if err != nil {
		return entity.Delivery{}, fmt.Errorf("failed to insert delivery to database: %w", err)
	}

	return delivery, nil
}
