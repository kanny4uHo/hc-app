package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

const makeReservationPath = "/api/v1/internal/inventory/item/reserve"
const cancelReservationPathFmt = "/api/v1/internal/inventory/reservation/%d/cancel"

var ErrReservationNotFound = fmt.Errorf("reservation is not found")

type InventoryClient struct {
	baseUserHost string
	httpClient   *http.Client
}

func NewInventoryClient(baseHost string, httpClient *http.Client) *InventoryClient {
	return &InventoryClient{
		baseUserHost: baseHost,
		httpClient:   httpClient,
	}
}

type MakeReservationRequest struct {
	OrderID uint64 `json:"order_id"`
	Item    Item   `json:"item"`
}

type Item struct {
	ItemID string `json:"item_id"`
	Amount uint64 `json:"amount"`
}

type MakeReservationResponse struct {
	ReservationID uint64 `json:"reservation_id"`
}

func (c *InventoryClient) MakeReservation(_ context.Context, req MakeReservationRequest) (MakeReservationResponse, error) {
	requestPath := fmt.Sprintf("%s%s", c.baseUserHost, makeReservationPath)

	body, err := json.Marshal(req)

	if err != nil {
		return MakeReservationResponse{}, fmt.Errorf("failed to marshal request body: %w", err)
	}

	response, err := c.httpClient.Post(requestPath, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return MakeReservationResponse{}, fmt.Errorf("failed to do POST request: %w", err)
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return MakeReservationResponse{}, fmt.Errorf("unexpected status code: %d", response.StatusCode)
	}

	var makeReservationResponse MakeReservationResponse

	err = json.NewDecoder(response.Body).Decode(&makeReservationResponse)
	if err != nil {
		return MakeReservationResponse{}, fmt.Errorf("failed to decode response body: %w", err)
	}

	return makeReservationResponse, nil
}

func (c *InventoryClient) CancelReservation(ctx context.Context, reservationID uint64) error {
	requestPath := fmt.Sprintf(cancelReservationPathFmt, reservationID)
	requestURL := fmt.Sprintf("%s%s", c.baseUserHost, requestPath)

	response, err := c.httpClient.Post(requestURL, "application/json", bytes.NewBuffer([]byte{}))
	if err != nil {
		return fmt.Errorf("failed to do POST request: %w", err)
	}

	if response.StatusCode != http.StatusOK {
		if response.StatusCode == http.StatusNotFound {
			return fmt.Errorf("reservation does not exist: %w", ErrReservationNotFound)
		}

		return fmt.Errorf("unexpected status code: %d", response.StatusCode)
	}

	return nil
}
