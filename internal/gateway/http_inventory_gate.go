package gateway

import (
	"context"
	"fmt"

	"healthcheckProject/internal/repository/httpclient"
	"healthcheckProject/internal/service"
)

type HttpInventoryGate struct {
	inventoryHttpClient *httpclient.InventoryClient
}

var _ service.InventoryGateway = (*HttpInventoryGate)(nil)

func NewHttpInventoryGate(client *httpclient.InventoryClient) *HttpInventoryGate {
	return &HttpInventoryGate{
		inventoryHttpClient: client,
	}
}

func (h *HttpInventoryGate) MakeReservation(ctx context.Context, orderID uint64, itemID string, amount uint64) (uint64, error) {
	reservationResponse, err := h.inventoryHttpClient.MakeReservation(ctx, httpclient.MakeReservationRequest{
		OrderID: orderID,
		Item: httpclient.Item{
			ItemID: itemID,
			Amount: amount,
		},
	})

	if err != nil {
		return 0, fmt.Errorf("failed to make reservation: %w", err)
	}

	return reservationResponse.ReservationID, nil
}

func (h *HttpInventoryGate) CancelReservation(ctx context.Context, reservationID uint64) error {
	err := h.inventoryHttpClient.CancelReservation(ctx, reservationID)
	if err != nil {
		return fmt.Errorf("failed to cancel reservation: %w", err)
	}

	return nil
}
