package gateway

import (
	"context"
	"fmt"

	"healthcheckProject/internal/entity"
	"healthcheckProject/internal/repository/httpclient"
	"healthcheckProject/internal/service"
)

type HttpAuthGate struct {
	authHttpClient *httpclient.AuthClient
}

var _ service.AuthGateway = (*HttpAuthGate)(nil)

func NewHttpAuthGate(client *httpclient.AuthClient) *HttpAuthGate {
	return &HttpAuthGate{
		authHttpClient: client,
	}
}

func (h *HttpAuthGate) RegisterUser(ctx context.Context, user entity.AddUserArgs) (entity.User, error) {
	registerResponse, err := h.authHttpClient.RegisterUser(ctx, httpclient.RegisterUserRequest{
		Login:    user.Login,
		Password: user.Password,
	})

	if err != nil {
		return entity.User{}, fmt.Errorf("auth http client failed to register user: %w", err)
	}

	return entity.User{
		PasswordHash: registerResponse.PasswordHash,
	}, nil
}
