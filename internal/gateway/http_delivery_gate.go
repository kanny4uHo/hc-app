package gateway

import (
	"context"
	"fmt"
	"healthcheckProject/internal/entity"

	"healthcheckProject/internal/repository/httpclient"
	"healthcheckProject/internal/service"
)

type HttpDeliveryGate struct {
	deliveryHttpClient *httpclient.DeliveryClient
}

var _ service.DeliveryGateway = (*HttpDeliveryGate)(nil)

func NewHttpDeliveryGate(client *httpclient.DeliveryClient) *HttpDeliveryGate {
	return &HttpDeliveryGate{
		deliveryHttpClient: client,
	}
}

func (h *HttpDeliveryGate) ApplyCourierForOrder(ctx context.Context, orderID uint64) (entity.Delivery, error) {
	courierResponse, err := h.deliveryHttpClient.ApplyCourierForOrder(ctx, httpclient.ApplyCourierRequest{OrderID: orderID})
	if err != nil {
		return entity.Delivery{}, fmt.Errorf("failed to apply courier by HTTP: %w", err)
	}

	return entity.Delivery{
		ID:        courierResponse.DeliveryID,
		OrderID:   orderID,
		CourierID: courierResponse.CourierID,
	}, nil
}
