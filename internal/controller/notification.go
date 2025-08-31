package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/segmentio/kafka-go"
	"healthcheckProject/internal/entity"
	"healthcheckProject/internal/gateway"
	"healthcheckProject/internal/service"
	"net/http"
	"time"
)

type NotificationController struct {
	service *service.NotificationService
}

func NewNotificationController(service *service.NotificationService) *NotificationController {
	return &NotificationController{
		service: service,
	}
}

type Notification struct {
	Timestamp time.Time `json:"timestamp"`
	Message   string    `json:"message"`
	Recipient string    `json:"recipient"`
}

type GetNotificationsResponse struct {
	Notifications []entity.Notification `json:"notifications"`
}

func (c *NotificationController) GetNotificationList(ctx *gin.Context) {
	notifications, err := c.service.GetNotifications(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

		return
	}

	respNotifications := make([]Notification, 0, len(notifications))

	for _, notification := range notifications {
		respNotifications = append(respNotifications, Notification{
			Timestamp: notification.Timestamp,
			Message:   notification.Message,
			Recipient: notification.Recipient,
		})
	}

	ctx.JSON(http.StatusOK,
		GetNotificationsResponse{
			Notifications: notifications,
		},
	)
}

func (c *NotificationController) ConsumeOrderIdPaid(ctx context.Context, reader *kafka.Reader) error {
	err := consumeTopic(ctx, reader, func(message kafka.Message) error {
		event := gateway.OrderIsPaidEvent{}

		err := json.Unmarshal(message.Value, &event)
		if err != nil {
			return fmt.Errorf("could not unmarshal event: %w", err)
		}

		err = c.service.OnOrderPaid(ctx, event.OrderID)
		if err != nil {
			return fmt.Errorf("could not process event: %w", err)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("could not process event: %w", err)
	}

	return nil
}

func (c *NotificationController) ConsumeOrderPaymentIsFailed(ctx context.Context, reader *kafka.Reader) error {
	err := consumeTopic(ctx, reader, func(message kafka.Message) error {
		event := gateway.OrderPaymentFailedEvent{}

		err := json.Unmarshal(message.Value, &event)
		if err != nil {
			return fmt.Errorf("could not unmarshal event: %w", err)
		}

		err = c.service.OnOrderPaymentFailed(ctx, event.OrderID)
		if err != nil {
			return fmt.Errorf("could not process event: %w", err)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("could not process event: %w", err)
	}

	return nil
}
