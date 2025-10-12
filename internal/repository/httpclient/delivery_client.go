package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

const applyCourierPath = "/api/v1/internal/delivery/apply"

type DeliveryClient struct {
	baseUserHost string
	httpClient   *http.Client
}

func NewDeliveryClient(baseHost string, httpClient *http.Client) *DeliveryClient {
	return &DeliveryClient{
		baseUserHost: baseHost,
		httpClient:   httpClient,
	}
}

type ApplyCourierRequest struct {
	OrderID uint64 `json:"order_id"`
}

type ApplyCourierResponse struct {
	CourierID  uint64 `json:"courier_id"`
	DeliveryID uint64 `json:"delivery_id"`
}

func (c *DeliveryClient) ApplyCourierForOrder(_ context.Context, req ApplyCourierRequest) (ApplyCourierResponse, error) {
	requestPath := fmt.Sprintf("%s%s", c.baseUserHost, applyCourierPath)

	body, err := json.Marshal(req)

	if err != nil {
		return ApplyCourierResponse{}, fmt.Errorf("failed to marshal request body: %w", err)
	}

	response, err := c.httpClient.Post(requestPath, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return ApplyCourierResponse{}, fmt.Errorf("failed to do POST request: %w", err)
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return ApplyCourierResponse{}, fmt.Errorf("unexpected status code: %d", response.StatusCode)
	}

	var applyCourierResponse ApplyCourierResponse

	err = json.NewDecoder(response.Body).Decode(&applyCourierResponse)
	if err != nil {
		return ApplyCourierResponse{}, fmt.Errorf("failed to decode response body: %w", err)
	}

	return applyCourierResponse, nil
}
