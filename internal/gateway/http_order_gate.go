package gateway

import (
	"context"
	"fmt"
	"healthcheckProject/internal/entity"
	"healthcheckProject/internal/repository/httpclient"
	"healthcheckProject/internal/service"
)

type HttpOrderGate struct {
	orderHttpClient *httpclient.OrderClient
}

var _ service.OrderGateway = (*HttpOrderGate)(nil)

func NewHttpOrderGate(client *httpclient.OrderClient) *HttpOrderGate {
	return &HttpOrderGate{
		orderHttpClient: client,
	}
}

func (h *HttpOrderGate) GetOrderByID(ctx context.Context, id uint64) (entity.Order, error) {
	orderInfo, err := h.orderHttpClient.GetOrderInfo(ctx, id)

	if err != nil {
		return entity.Order{}, fmt.Errorf("failed to get order info: %w", err)
	}

	return entity.Order{
		ID:     id,
		Price:  orderInfo.Price,
		Item:   orderInfo.Item,
		Status: entity.Status(orderInfo.Status),
		Owner: entity.UserShort{
			ID: orderInfo.UserID,
		},
	}, nil
}
