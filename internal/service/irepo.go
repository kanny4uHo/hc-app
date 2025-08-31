package service

import (
	"context"

	"healthcheckProject/internal/entity"
)

type UserRepo interface {
	WithinTransaction(ctx context.Context, f func(ctx context.Context) error) error
	AddUser(ctx context.Context, user entity.AddUserArgs, passwordHash string) (entity.User, error)
	GetUserByID(ctx context.Context, id uint64) (entity.User, error)
	GetUserByLogin(ctx context.Context, login string) (entity.User, error)
	GetUserByEmail(ctx context.Context, email string) (entity.User, error)
	UpdateUser(ctx context.Context, id uint64, args UpdateUserArgs) error
	DeleteUser(ctx context.Context, id uint64) error
}

type AuthGateway interface {
	RegisterUser(ctx context.Context, user entity.AddUserArgs) (entity.User, error)
}

type CredentialRepo interface {
	GetUserByLogin(ctx context.Context, login string) (entity.UserCreds, error)
	GetUserByID(ctx context.Context, id uint64) (entity.UserCreds, error)
}

type BillingGateway interface {
	GetUserAccountByID(ctx context.Context, id uint64) (entity.UserAccount, error)
}

type OrderRepo interface {
	CreateOrder(ctx context.Context, user entity.UserShort, price uint64, item string) (entity.Order, error)
	DeleteOrder(ctx context.Context, id uint64) error
	GetOrderByID(ctx context.Context, id uint64) (entity.Order, error)
	SetOrderStatus(ctx context.Context, id uint64, status entity.Status) error
}

type EventRepo interface {
	OnOrderCreated(ctx context.Context, order entity.Order) error
	OnOrderPaid(ctx context.Context, orderID uint64) error
	OnOrderPaymentFailed(ctx context.Context, orderID uint64) error
	OnUserRegistered(ctx context.Context, user entity.User) error
}

type UserAccountRepo interface {
	WithinTransaction(ctx context.Context, f func(ctx context.Context) error) error
	SaveUser(ctx context.Context, account entity.UserAccount) error
	SetBalance(ctx context.Context, userID uint64, amount int64) (int64, error)
	GetUserAccountByID(ctx context.Context, userID uint64) (entity.UserAccount, error)
}

type OrderGateway interface {
	GetOrderByID(ctx context.Context, id uint64) (entity.Order, error)
}

type NotificationGateway interface {
	SendOrderPaidEmail(ctx context.Context, recipient string, orderID uint64) error
	SendOrderPaymentFailed(ctx context.Context, recipient string, orderID uint64) error
	GetNotificationsByEmail(ctx context.Context, email string) ([]entity.Notification, error)
}
