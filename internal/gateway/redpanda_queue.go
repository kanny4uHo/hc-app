package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/segmentio/kafka-go"
	"healthcheckProject/internal/entity"
	"healthcheckProject/internal/service"
)

type KafkaLogger struct{}

func (k KafkaLogger) Printf(format string, v ...interface{}) {
	log.Printf(format, v...)
}

type KafkaEventGateway struct {
	writer                  *kafka.Writer
	newOrdersTopic          string
	newUsersTopic           string
	orderIsPaidTopic        string
	orderPaymentFailedTopic string
}

func NewKafkaEventGateway(
	writer *kafka.Writer,
	newOrdersTopic string,
	newUsersTopic string,
	orderIsPaidTopic string,
	orderPaymentFailedTopic string,
) *KafkaEventGateway {
	return &KafkaEventGateway{
		writer:                  writer,
		newOrdersTopic:          newOrdersTopic,
		newUsersTopic:           newUsersTopic,
		orderIsPaidTopic:        orderIsPaidTopic,
		orderPaymentFailedTopic: orderPaymentFailedTopic,
	}
}

func (k *KafkaEventGateway) OnOrderPaid(ctx context.Context, orderID uint64) error {
	event := OrderIsPaidEvent{
		OrderID: orderID,
	}

	message, err := json.Marshal(&event)
	if err != nil {
		return fmt.Errorf("failed to marshal event for order %d: %w", orderID, err)
	}

	err = k.writer.WriteMessages(ctx, kafka.Message{Topic: k.orderIsPaidTopic, Value: message})
	if err != nil {
		return fmt.Errorf("failed to write messages for order %d: %w", orderID, err)
	}

	return nil
}

func (k *KafkaEventGateway) OnOrderPaymentFailed(ctx context.Context, orderID uint64) error {
	event := OrderPaymentFailedEvent{OrderID: orderID}

	message, err := json.Marshal(&event)
	if err != nil {
		return fmt.Errorf("failed to marshal event for order %d: %w", orderID, err)
	}

	err = k.writer.WriteMessages(ctx, kafka.Message{Topic: k.orderPaymentFailedTopic, Value: message})
	if err != nil {
		return fmt.Errorf("failed to write messages for order %d: %w", orderID, err)
	}

	return nil
}

func (k *KafkaEventGateway) OnUserRegistered(ctx context.Context, user entity.User) error {
	event := NewUserEvent{ID: user.UserShort.ID}

	message, err := json.Marshal(&event)
	if err != nil {
		return fmt.Errorf("failed to marshal event for user %d: %w", user.UserShort.ID, err)
	}

	err = k.writer.WriteMessages(ctx, kafka.Message{Topic: k.newUsersTopic, Value: message})
	if err != nil {
		return fmt.Errorf("failed to write messages for user %d: %w", user.UserShort.ID, err)
	}

	return nil
}

func (k *KafkaEventGateway) OnOrderCreated(ctx context.Context, order entity.Order) error {
	event := OrderEvent{
		Type: OrderCreated,
		ID:   order.ID,
	}

	message, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	err = k.writer.WriteMessages(ctx, kafka.Message{Topic: k.newOrdersTopic, Value: message})
	if err != nil {
		return fmt.Errorf("failed to write messages: %w", err)
	}

	return nil
}

var _ service.EventRepo = (*KafkaEventGateway)(nil)
