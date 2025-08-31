package httpclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

const getOrderPath = "/api/v1/internal/order/"

var ErrOrderNotFound = fmt.Errorf("order not found")

type OrderClient struct {
	baseUserHost string
	httpClient   *http.Client
}

func NewOrderClient(baseHost string, httpClient *http.Client) *OrderClient {
	return &OrderClient{
		baseUserHost: baseHost,
		httpClient:   httpClient,
	}
}

type OrderResponse struct {
	ID     uint64 `json:"id"`
	Price  uint64 `json:"price"`
	Status string `json:"status"`
	Item   string `json:"item"`
	UserID uint64 `json:"user_id"`
}

func (c *OrderClient) GetOrderInfo(_ context.Context, orderID uint64) (OrderResponse, error) {
	requestPath := fmt.Sprintf("%s%s%d", c.baseUserHost, getOrderPath, orderID)
	response, err := c.httpClient.Get(requestPath)
	if err != nil {
		return OrderResponse{}, fmt.Errorf("failed to get order info: %w", err)
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		if response.StatusCode == http.StatusNotFound {
			return OrderResponse{}, ErrOrderNotFound
		}

		return OrderResponse{}, fmt.Errorf("failed to get order info, status: %d", response.StatusCode)
	}

	var orderResponse OrderResponse

	err = json.NewDecoder(response.Body).Decode(&orderResponse)

	if err != nil {
		return OrderResponse{}, fmt.Errorf("failed to decode order info")
	}

	return orderResponse, nil
}
