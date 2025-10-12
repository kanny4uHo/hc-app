package service

import (
	"context"
	"fmt"
	"log"
	"math/rand"

	"healthcheckProject/internal/entity"
)

type DeliveryService struct {
	repository DeliveryRepository
}

func NewDeliveryService(repository DeliveryRepository) *DeliveryService {
	return &DeliveryService{
		repository: repository,
	}
}

func (s *DeliveryService) ApplyCourierForOrder(ctx context.Context, orderID uint64) (entity.Delivery, error) {
	couriers, err := s.repository.GetCouriersOnShiftWithoutActiveDeliveries(ctx)
	if err != nil {
		return entity.Delivery{}, fmt.Errorf("failed to get couriers: %w", err)
	}

	if len(couriers) == 0 {
		return entity.Delivery{}, fmt.Errorf("no couriers")
	}

	randomCourier := couriers[rand.Intn(len(couriers))]

	delivery := entity.Delivery{
		OrderID:   orderID,
		CourierID: randomCourier.ID,
		Status:    entity.DeliveryStatusOnTheWay,
	}

	appliedDelivery, err := s.repository.AddDelivery(ctx, delivery)

	if err != nil {
		return entity.Delivery{}, fmt.Errorf("failed to add delivery: %w", err)
	}

	return appliedDelivery, nil
}

func (s *DeliveryService) GetDeliveries(ctx context.Context) ([]entity.Delivery, error) {
	deliveries, err := s.repository.GetDeliveries(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get deliveries from repo: %w", err)
	}

	log.Printf("fetched deliveries %+v", deliveries)

	return deliveries, nil
}

func (s *DeliveryService) GetDeliveryInfo(ctx context.Context, deliveryID uint64) (entity.Delivery, error) {
	delivery, err := s.repository.GetDeliveryByID(ctx, deliveryID)
	if err != nil {
		return entity.Delivery{}, fmt.Errorf("failed to get delivery from repo: %w", err)
	}

	return delivery, nil
}
