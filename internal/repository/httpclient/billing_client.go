package httpclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

const getUserAccountPath = "/api/v1/billing/account/"

var ErrUserAccountNotFound = fmt.Errorf("user account not found")

type BillingClient struct {
	baseUserHost string
	httpClient   *http.Client
}

func NewBillingClient(baseHost string, httpClient *http.Client) *BillingClient {
	return &BillingClient{
		baseUserHost: baseHost,
		httpClient:   httpClient,
	}
}

type GetUserAccountByIDResponse struct {
	ID      uint64 `json:"id"`
	UserID  uint64 `json:"user_id"`
	Balance int64  `json:"balance"`
}

func (c *BillingClient) GetUserAccountInfo(_ context.Context, userID uint64) (GetUserAccountByIDResponse, error) {
	requestPath := fmt.Sprintf("%s%s%d", c.baseUserHost, getUserAccountPath, userID)
	response, err := c.httpClient.Get(requestPath)
	if err != nil {
		return GetUserAccountByIDResponse{}, fmt.Errorf("failed to get order info: %w", err)
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		if response.StatusCode == http.StatusNotFound {
			return GetUserAccountByIDResponse{}, ErrUserAccountNotFound
		}

		return GetUserAccountByIDResponse{}, fmt.Errorf("failed to get order info, status: %d", response.StatusCode)
	}

	var userAccountResponse GetUserAccountByIDResponse

	err = json.NewDecoder(response.Body).Decode(&userAccountResponse)

	if err != nil {
		return GetUserAccountByIDResponse{}, fmt.Errorf("failed to decode order info")
	}

	return userAccountResponse, nil
}
