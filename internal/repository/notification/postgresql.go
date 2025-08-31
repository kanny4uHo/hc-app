package notification

import (
	"context"
	"database/sql"
	"fmt"
	"healthcheckProject/internal/entity"
	"time"

	"healthcheckProject/internal/service"
)

type PgNotificationRepo struct {
	db *sql.DB
}

func (p *PgNotificationRepo) GetNotificationsByEmail(ctx context.Context, recipient string) ([]entity.Notification, error) {
	rows, err := p.db.QueryContext(ctx, "SELECT id, timestamp, recipient_email, message FROM notifications WHERE recipient_email = $1", recipient)
	if err != nil {
		return nil, fmt.Errorf("getNotificationsByUserID query: %w", err)
	}
	defer rows.Close()
	notifications := make([]entity.Notification, 0, 10)
	for rows.Next() {
		var notification entity.Notification
		scanErr := rows.Scan(&notification.ID, &notification.Timestamp, &notification.Recipient, &notification.Message)
		if scanErr != nil {
			return nil, fmt.Errorf("getNotificationsByUserID scan: %w", scanErr)
		}

		notifications = append(notifications, notification)
	}

	return notifications, nil
}

func NewPgNotificationRepo(db *sql.DB) *PgNotificationRepo {
	return &PgNotificationRepo{db: db}
}

func (p *PgNotificationRepo) SendOrderPaidEmail(ctx context.Context, recipient string, orderID uint64) error {
	_, err := p.db.ExecContext(
		ctx,
		"INSERT INTO notifications(timestamp, order_id, recipient_email, message) VALUES ($1, $2, $3, $4)",
		time.Now(),
		orderID,
		recipient,
		fmt.Sprintf("success payment for order %d", orderID),
	)

	if err != nil {
		return fmt.Errorf("could not send order paid email: %w", err)
	}

	return nil
}

func (p *PgNotificationRepo) SendOrderPaymentFailed(ctx context.Context, recipient string, orderID uint64) error {
	_, err := p.db.ExecContext(
		ctx,
		"INSERT INTO notifications(timestamp, order_id, recipient_email, message) VALUES ($1, $2, $3, $4)",
		time.Now(),
		orderID,
		recipient,
		fmt.Sprintf("not enough money for order %d", orderID),
	)

	if err != nil {
		return fmt.Errorf("could not send order paid email: %w", err)
	}

	return nil
}

var _ service.NotificationGateway = (*PgNotificationRepo)(nil)
