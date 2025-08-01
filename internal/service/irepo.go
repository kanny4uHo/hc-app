package service

import (
	"context"

	"healthcheckProject/internal/entity"
)

type UserRepo interface {
	WithinTransaction(ctx context.Context, f func(ctx context.Context) error) error
	AddUser(ctx context.Context, user entity.AddUserArgs, passwordHash string) (entity.User, error)
	GetUserByID(ctx context.Context, id int64) (entity.User, error)
	GetUserByLogin(ctx context.Context, login string) (entity.User, error)
	GetUserByEmail(ctx context.Context, email string) (entity.User, error)
	UpdateUser(ctx context.Context, id int64, args UpdateUserArgs) error
	DeleteUser(ctx context.Context, id int64) error
}

type AuthGateway interface {
	RegisterUser(ctx context.Context, user entity.AddUserArgs) (entity.User, error)
}

type CredentialRepo interface {
	GetUserByLogin(ctx context.Context, login string) (entity.UserCreds, error)
}
