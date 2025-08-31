package service

import (
	"context"
	"fmt"
	"healthcheckProject/internal/entity"
	"healthcheckProject/internal/utils"
)

type OrderService struct {
	eventRepo EventRepo
	orderRepo OrderRepo
}

func NewOrderService(
	orderRepo OrderRepo,
	eventRepo EventRepo,
) *OrderService {
	return &OrderService{
		orderRepo: orderRepo,
		eventRepo: eventRepo,
	}
}

func (s *OrderService) CreateOrder(ctx context.Context, order entity.CreateOrderArgs) (entity.Order, error) {
	user := utils.GetUser(ctx)
	createdOrder, err := s.orderRepo.CreateOrder(ctx, *user, order.Price, order.Item)
	if err != nil {
		return entity.Order{}, fmt.Errorf("failed to create order: %w", err)
	}

	createdOrder.Status = entity.StatusCreated

	err = s.eventRepo.OnOrderCreated(ctx, createdOrder)
	if err != nil {
		return entity.Order{}, fmt.Errorf("failed to notify about order creation: %w", err)
	}

	return createdOrder, nil
}

func (s *OrderService) GetOrder(ctx context.Context, id uint64) (entity.Order, error) {
	order, err := s.orderRepo.GetOrderByID(ctx, id)
	if err != nil {
		return entity.Order{}, fmt.Errorf("failed to get order: %w", err)
	}

	return order, nil
}

func (s *OrderService) OnOrderPaid(ctx context.Context, orderID uint64) error {
	order, err := s.orderRepo.GetOrderByID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("failed to get order: %w", err)
	}

	err = s.orderRepo.SetOrderStatus(ctx, order.ID, entity.StatusPaid)
	if err != nil {
		return fmt.Errorf("failed to set order status: %w", err)
	}

	return nil
}
