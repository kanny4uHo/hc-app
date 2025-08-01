package repository

import (
	"context"
	"errors"
	"fmt"
	"log"

	"healthcheckProject/internal/entity"
	"healthcheckProject/internal/repository/httpclient"
	"healthcheckProject/internal/service"
)

type UserServiceRepo struct {
	userClient *httpclient.UserClient
}

var _ service.CredentialRepo = (*UserServiceRepo)(nil)

func NewUserServiceRepo(userClient *httpclient.UserClient) *UserServiceRepo {
	return &UserServiceRepo{
		userClient: userClient,
	}
}

func (u *UserServiceRepo) GetUserByLogin(ctx context.Context, login string) (entity.UserCreds, error) {
	info, err := u.userClient.GetUserInfo(ctx, login)
	if err != nil {
		log.Printf("failed to get user info %s\n", err)

		if errors.Is(err, httpclient.ErrUserNotFound) {
			return entity.UserCreds{}, service.ErrUserNotFound
		}

		return entity.UserCreds{}, fmt.Errorf("failed to get user info: %w", err)
	}

	return entity.UserCreds{
		ID:           info.UserID,
		Login:        info.Login,
		PasswordHash: info.PasswordHash,
	}, nil
}
