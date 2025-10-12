package gateway

import (
	"context"
	"fmt"

	"healthcheckProject/internal/entity"
	"healthcheckProject/internal/repository/httpclient"
	"healthcheckProject/internal/service"
)

type HttpBillingGate struct {
	billingHttpClient *httpclient.BillingClient
}

var _ service.BillingGateway = (*HttpBillingGate)(nil)

func NewHttpBillingGate(client *httpclient.BillingClient) *HttpBillingGate {
	return &HttpBillingGate{
		billingHttpClient: client,
	}
}

func (h *HttpBillingGate) GetUserAccountByID(ctx context.Context, id uint64) (entity.UserAccount, error) {
	userAccount, err := h.billingHttpClient.GetUserAccountInfo(ctx, id)

	if err != nil {
		return entity.UserAccount{}, fmt.Errorf("failed to get order info: %w", err)
	}

	return entity.UserAccount{
		ID:      userAccount.ID,
		UserID:  userAccount.UserID,
		Balance: userAccount.Balance,
	}, nil
}

func (h *HttpBillingGate) WithdrawMoney(ctx context.Context, userID uint64, amount uint64) error {
	_, err := h.billingHttpClient.WithdrawMoney(ctx, userID, amount)
	if err != nil {
		return fmt.Errorf("failed to withdraw money by HTTP call: %w", err)
	}

	return nil
}

func (h *HttpBillingGate) CreditMoney(ctx context.Context, userID uint64, amount uint64) error {
	_, err := h.billingHttpClient.CreditMoney(ctx, userID, amount)
	if err != nil {
		return fmt.Errorf("failed to credit money by HTTP call: %w", err)
	}

	return nil
}
