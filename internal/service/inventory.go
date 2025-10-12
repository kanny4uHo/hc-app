package service

import (
	"context"
	"errors"
	"fmt"
	"log"

	"healthcheckProject/internal/entity"
)

type InventoryService struct {
	inventoryRepo InventoryRepo
}

func NewInventoryService(inventoryRepo InventoryRepo) *InventoryService {
	return &InventoryService{inventoryRepo: inventoryRepo}
}

func (s *InventoryService) AddItem(ctx context.Context, item entity.Item) (entity.Item, error) {
	var existingItem entity.Item

	err := s.inventoryRepo.WithinTransaction(ctx, func(ctx context.Context) error {
		var err error

		existingItem, err = s.inventoryRepo.GetItem(ctx, item.ItemID, Opts{ToUpdate: true})

		if err != nil {
			if errors.Is(err, ErrNoItemsInInventory) {
				err = s.inventoryRepo.AddItem(ctx, item)
				if err != nil {
					return fmt.Errorf("add item to repo failed: %w", err)
				}

				existingItem = item

				return nil
			}

			return fmt.Errorf("get item from repo failed: %w", err)
		}

		existingItem.Amount += item.Amount
		err = s.inventoryRepo.UpdateItem(ctx, existingItem)
		if err != nil {
			return fmt.Errorf("update item to repo failed: %w", err)
		}

		return nil
	})

	if err != nil {
		return entity.Item{}, fmt.Errorf("add items transaction failed: %w", err)
	}

	return existingItem, nil
}

func (s *InventoryService) ReserveItem(ctx context.Context, orderID uint64, item entity.Item) (entity.Reservation, error) {
	var reservation entity.Reservation

	err := s.inventoryRepo.WithinTransaction(ctx, func(ctx context.Context) error {
		inventoryItem, err := s.inventoryRepo.GetItem(ctx, item.ItemID, Opts{ToUpdate: true})
		if err != nil {
			return fmt.Errorf("InventoryService.ReserveItems, no item count: %w", err)
		}

		if inventoryItem.Amount < item.Amount {
			return fmt.Errorf("%w: InventoryService.ReserveItems, not enough items available: %d", ErrNoItemsInInventory, inventoryItem.Amount)
		}

		inventoryItem.Amount -= item.Amount

		updateError := s.inventoryRepo.UpdateItem(ctx, inventoryItem)
		if updateError != nil {
			return fmt.Errorf("failed to update item in repo: %w", updateError)
		}

		var reserveError error

		reservation, reserveError = s.inventoryRepo.ReserveItem(ctx, orderID, item)
		if reserveError != nil {
			return fmt.Errorf("failed to reserve item in repo: %w", reserveError)
		}

		return nil
	})

	if err != nil {
		return entity.Reservation{}, fmt.Errorf("failed to reserve item: %w", err)
	}

	return reservation, nil
}

func (s *InventoryService) GetReservationInfo(ctx context.Context, reservationID uint64) (entity.Reservation, error) {
	item, err := s.inventoryRepo.GetReservationItem(ctx, reservationID)
	if err != nil {
		return entity.Reservation{}, fmt.Errorf("failed to get reservation item from repo: %w", err)
	}

	return item, nil
}

func (s *InventoryService) CancelReservation(ctx context.Context, reservationID uint64) error {
	log.Printf("InventoryService.CancelReservation: Reservation ID: %d", reservationID)

	reservationItem, err := s.inventoryRepo.GetReservationItem(ctx, reservationID)

	if err != nil {
		return fmt.Errorf("failed to get reservation: %w", err)
	}

	err = s.inventoryRepo.WithinTransaction(ctx, func(ctx context.Context) error {
		item, errTx := s.inventoryRepo.GetItem(ctx, reservationItem.Item.ItemID, Opts{ToUpdate: true})
		if errTx != nil {
			return fmt.Errorf("failed to get item: %w", errTx)
		}

		item.Amount += reservationItem.Item.Amount

		errTx = s.inventoryRepo.UpdateItem(ctx, item)
		if errTx != nil {
			return fmt.Errorf("failed to update item in repo: %w", err)
		}

		errTx = s.inventoryRepo.DeleteReservationItem(ctx, reservationID)
		if errTx != nil {
			return fmt.Errorf("failed to delete reservation from repo: %w", errTx)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to delete reservation within tx: %w", err)
	}

	return nil
}
