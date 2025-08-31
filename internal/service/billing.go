package service

import (
	"context"
	"errors"
	"fmt"
	"log"

	"healthcheckProject/internal/entity"
)

type BillingService struct {
	userRepo     UserAccountRepo
	orderGateway OrderGateway
	eventRepo    EventRepo
}

func NewBillingService(userRepo UserAccountRepo, gateway OrderGateway, eventRepo EventRepo) *BillingService {
	return &BillingService{
		userRepo:     userRepo,
		orderGateway: gateway,
		eventRepo:    eventRepo,
	}
}

func (s *BillingService) OnNewUserRegistered(ctx context.Context, userID uint64) error {
	newUser := entity.UserAccount{
		UserID:  userID,
		Balance: 0,
	}

	err := s.userRepo.SaveUser(ctx, newUser)
	if err != nil {
		return fmt.Errorf("failed to save new user: %w", err)
	}

	return nil
}

func (s *BillingService) OnOrderCreated(ctx context.Context, orderID uint64) error {
	log.Printf("OnOrderCreated orderID: %d", orderID)
	orderInfo, err := s.orderGateway.GetOrderByID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("failed to get order by id %d: %w", orderID, err)
	}

	userAccount, err := s.WithdrawMoney(ctx, orderInfo.Owner.ID, orderInfo.Price)
	if err != nil {
		if errors.Is(err, ErrNotEnoughMoney) {
			log.Printf("failed to withdraw money for orderID %d, err %s", orderID, err)
			err = s.eventRepo.OnOrderPaymentFailed(ctx, orderID)
			if err != nil {
				log.Printf("failed to handle order payment failed, err %s", err)
			}

			return nil
		}

		return fmt.Errorf("failed to withdraw money for orderID %d: %w", orderID, err)
	}

	log.Printf("succeded with payment, userAccount.Balance: %d", userAccount.Balance)
	err = s.eventRepo.OnOrderPaid(ctx, orderID)
	if err != nil {
		log.Printf("failed to send order paid event for orderID %d, err %s", orderID, err)
	}

	return nil
}

func (s *BillingService) CreditMoney(ctx context.Context, userID, amount uint64) (entity.UserAccount, error) {
	var userAccount entity.UserAccount
	err := s.userRepo.WithinTransaction(ctx, func(ctx context.Context) error {
		var err error
		userAccount, err = s.userRepo.GetUserAccountByID(ctx, userID)
		if err != nil {
			return fmt.Errorf("failed to get user by id: %w", err)
		}

		newBalance := userAccount.Balance + int64(amount)

		newBalance, err = s.userRepo.SetBalance(ctx, userID, newBalance)
		if err != nil {
			return fmt.Errorf("failed to set balance to user: %w", err)
		}

		userAccount.Balance = newBalance

		return nil
	})

	if err != nil {
		return userAccount, fmt.Errorf("failed to add money to user: %w", err)
	}

	return userAccount, nil
}

func (s *BillingService) WithdrawMoney(ctx context.Context, userID, amount uint64) (entity.UserAccount, error) {
	var userAccount entity.UserAccount
	err := s.userRepo.WithinTransaction(ctx, func(ctx context.Context) error {
		var err error
		userAccount, err = s.userRepo.GetUserAccountByID(ctx, userID)
		if err != nil {
			return fmt.Errorf("failed to get user by id: %w", err)
		}

		if userAccount.Balance < int64(amount) {
			return fmt.Errorf("failed to withdraw money, %w: balance %d", ErrNotEnoughMoney, userAccount.Balance)
		}

		newBalance := userAccount.Balance - int64(amount)

		newBalance, err = s.userRepo.SetBalance(ctx, userID, newBalance)
		if err != nil {
			return fmt.Errorf("failed to set balance to user: %w", err)
		}

		userAccount.Balance = newBalance

		return nil
	})

	if err != nil {
		return userAccount, fmt.Errorf("failed to add money to user: %w", err)
	}

	return userAccount, nil
}

func (s *BillingService) GetUserAccountByID(ctx context.Context, userID uint64) (entity.UserAccount, error) {
	userAccount, err := s.userRepo.GetUserAccountByID(ctx, userID)
	if err != nil {
		return userAccount, fmt.Errorf("failed to get user account by id: %w", err)
	}

	return userAccount, nil
}
