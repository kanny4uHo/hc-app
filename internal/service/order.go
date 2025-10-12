package service

import (
	"context"
	"fmt"
	"log"
	"slices"

	"healthcheckProject/internal/entity"
	"healthcheckProject/internal/utils"
)

type OrderService struct {
	eventRepo     EventRepo
	orderRepo     OrderRepo
	billingGate   BillingGateway
	inventoryGate InventoryGateway
	deliveryGate  DeliveryGateway
}

func NewOrderService(
	orderRepo OrderRepo,
	eventRepo EventRepo,
	billingGate BillingGateway,
	inventoryGate InventoryGateway,
	deliveryGate DeliveryGateway,
) *OrderService {
	return &OrderService{
		orderRepo:     orderRepo,
		eventRepo:     eventRepo,
		billingGate:   billingGate,
		inventoryGate: inventoryGate,
		deliveryGate:  deliveryGate,
	}
}

func (s *OrderService) CreateOrder(ctx context.Context, order entity.CreateOrderArgs) (entity.Order, error) {
	user := utils.GetUser(ctx)
	createdOrder, err := s.orderRepo.CreateOrder(ctx, *user, order.Price, order.Item, entity.StatusCreated)
	if err != nil {
		return entity.Order{}, fmt.Errorf("failed to create order: %w", err)
	}

	err = s.eventRepo.OnOrderCreated(ctx, createdOrder)
	if err != nil {
		return entity.Order{}, fmt.Errorf("failed to notify about order creation: %w", err)
	}

	return createdOrder, nil
}

func (s *OrderService) CreateOrderSaga(ctx context.Context, order entity.CreateOrderArgs) (entity.Order, error) {
	user := utils.GetUser(ctx)

	applyRevertActions := false
	revertActions := make([]func() error, 0, 5)

	defer func() {
		if !applyRevertActions {
			log.Printf("SAGA has processed successfully")

			return
		}

		log.Printf("SAGA has failed, using revert actions")
		slices.Reverse(revertActions)
		for _, action := range revertActions {
			for i := 0; i < 3; i++ {
				err := action()
				if err == nil {
					break
				}

				log.Printf("failed to revert action, retry %d: %v", i+1, err)
			}
		}
	}()

	createdOrder, err := s.orderRepo.CreateOrder(ctx, *user, order.Price, order.Item, entity.StatusPaid)
	if err != nil {
		return entity.Order{}, fmt.Errorf("failed to create order: %w", err)
	}

	revertActions = append(revertActions, func() error {
		log.Printf("reverting order creation")
		deletionError := s.orderRepo.DeleteOrder(ctx, createdOrder.ID)
		if deletionError != nil {
			return fmt.Errorf("failed to delete order: %w", deletionError)
		}

		return nil
	})

	err = s.billingGate.WithdrawMoney(ctx, user.ID, order.Price)
	if err != nil {
		applyRevertActions = true
		return entity.Order{}, fmt.Errorf("failed to pay for order: %w", err)
	}

	revertActions = append(revertActions, func() error {
		log.Printf("reverting money withdrawal")
		creditError := s.billingGate.CreditMoney(ctx, user.ID, order.Price)
		if creditError != nil {
			return fmt.Errorf("failed to credit money: %w", creditError)
		}

		return nil
	})

	reservationID, err := s.inventoryGate.MakeReservation(ctx, createdOrder.ID, createdOrder.Item, 1)

	if err != nil {
		applyRevertActions = true
		return entity.Order{}, fmt.Errorf("failed to make reservation: %w", err)
	}

	revertActions = append(revertActions, func() error {
		log.Printf("reverting inventory reservation")
		reservationError := s.inventoryGate.CancelReservation(ctx, reservationID)
		if reservationError != nil {
			return fmt.Errorf("failed to cancel reservation: %w", reservationError)
		}

		return nil
	})

	createdOrder.ReservationID = reservationID

	deliveryInfo, err := s.deliveryGate.ApplyCourierForOrder(ctx, createdOrder.ID)

	if err != nil {
		applyRevertActions = true
		return entity.Order{}, fmt.Errorf("failed to apply courier for order: %w", err)
	}

	createdOrder.DeliveryID = deliveryInfo.ID

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
