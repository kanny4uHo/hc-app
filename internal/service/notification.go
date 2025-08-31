package service

import (
	"context"
	"fmt"
	"healthcheckProject/internal/entity"
	"healthcheckProject/internal/utils"
)

type NotificationService struct {
	sender       NotificationGateway
	userRepo     CredentialRepo
	orderGateway OrderGateway
}

func NewNotificationService(
	sender NotificationGateway,
	userRepo CredentialRepo,
	orderGateway OrderGateway,
) *NotificationService {
	return &NotificationService{
		sender:       sender,
		userRepo:     userRepo,
		orderGateway: orderGateway,
	}
}

func (s *NotificationService) OnOrderPaid(ctx context.Context, orderID uint64) error {
	order, err := s.orderGateway.GetOrderByID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("failed to get order by id %d, err %s", orderID, err)
	}

	userInfo, err := s.userRepo.GetUserByID(ctx, order.Owner.ID)
	if err != nil {
		return fmt.Errorf("get user by id: %w", err)
	}

	err = s.sender.SendOrderPaidEmail(ctx, userInfo.Email, orderID)
	if err != nil {
		return fmt.Errorf("send order paid email: %w", err)
	}

	return nil
}

func (s *NotificationService) OnOrderPaymentFailed(ctx context.Context, orderID uint64) error {
	order, err := s.orderGateway.GetOrderByID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("failed to get order by id %d, err %s", orderID, err)
	}

	userInfo, err := s.userRepo.GetUserByID(ctx, order.Owner.ID)
	if err != nil {
		return fmt.Errorf("get user by id: %w", err)
	}

	err = s.sender.SendOrderPaymentFailed(ctx, userInfo.Email, orderID)
	if err != nil {
		return fmt.Errorf("send order payment failed email: %w", err)
	}

	return nil
}

func (s *NotificationService) GetNotifications(ctx context.Context) ([]entity.Notification, error) {
	user := utils.GetUser(ctx)
	userInfo, err := s.userRepo.GetUserByID(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("get user by id: %w", err)
	}

	notifications, err := s.sender.GetNotificationsByEmail(ctx, userInfo.Email)
	if err != nil {
		return nil, fmt.Errorf("get notifications: %w", err)
	}

	return notifications, nil
}
