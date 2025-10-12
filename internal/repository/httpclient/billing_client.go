package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

const getUserAccountPath = "/api/v1/billing/account/"
const withdrawMoneyPath = "/api/v1/billing/money/withdraw"
const creditMoneyPath = "/api/v1/billing/money/credit"

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

type WithdrawMoneyResponse struct {
	UserID  uint64 `json:"user_id"`
	Balance int64  `json:"balance"`
}

type withdrawMoneyRequest struct {
	UserID uint64 `json:"user_id"`
	Amount uint64 `json:"amount"`
}

func (c *BillingClient) WithdrawMoney(_ context.Context, userID uint64, amount uint64) (WithdrawMoneyResponse, error) {
	requestPath := fmt.Sprintf("%s%s", c.baseUserHost, withdrawMoneyPath)
	body, err := json.Marshal(withdrawMoneyRequest{
		UserID: userID,
		Amount: amount,
	})

	if err != nil {
		return WithdrawMoneyResponse{}, fmt.Errorf("failed to marshal withdraw money request: %w", err)
	}

	response, err := c.httpClient.Post(requestPath, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return WithdrawMoneyResponse{}, fmt.Errorf("failed to get order info: %w", err)
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return WithdrawMoneyResponse{}, fmt.Errorf("failed to withdraw money, status: %d", response.StatusCode)
	}

	var withdrawMoneyResponse WithdrawMoneyResponse

	err = json.NewDecoder(response.Body).Decode(&withdrawMoneyResponse)
	if err != nil {
		return WithdrawMoneyResponse{}, fmt.Errorf("failed to decode withdraw response")
	}

	return withdrawMoneyResponse, nil
}

func (c *BillingClient) CreditMoney(_ context.Context, userID uint64, amount uint64) (WithdrawMoneyResponse, error) {
	requestPath := fmt.Sprintf("%s%s", c.baseUserHost, creditMoneyPath)
	body, err := json.Marshal(withdrawMoneyRequest{
		UserID: userID,
		Amount: amount,
	})

	if err != nil {
		return WithdrawMoneyResponse{}, fmt.Errorf("failed to marshal withdraw money request: %w", err)
	}

	response, err := c.httpClient.Post(requestPath, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return WithdrawMoneyResponse{}, fmt.Errorf("failed to get order info: %w", err)
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return WithdrawMoneyResponse{}, fmt.Errorf("failed to withdraw money, status: %d", response.StatusCode)
	}

	var withdrawMoneyResponse WithdrawMoneyResponse

	err = json.NewDecoder(response.Body).Decode(&withdrawMoneyResponse)
	if err != nil {
		return WithdrawMoneyResponse{}, fmt.Errorf("failed to decode withdraw response")
	}

	return withdrawMoneyResponse, nil
}
