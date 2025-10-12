package service

import (
	"context"

	"healthcheckProject/internal/entity"
)

type Opts struct {
	ToUpdate bool
}

type UserRepo interface {
	WithinTransaction(ctx context.Context, f func(ctx context.Context) error) error
	AddUser(ctx context.Context, user entity.AddUserArgs, passwordHash string) (entity.User, error)
	GetUserByID(ctx context.Context, id uint64) (entity.User, error)
	GetUserByLogin(ctx context.Context, login string) (entity.User, error)
	GetUserByEmail(ctx context.Context, email string) (entity.User, error)
	UpdateUser(ctx context.Context, id uint64, args UpdateUserArgs) error
	DeleteUser(ctx context.Context, id uint64) error
}

type InventoryRepo interface {
	WithinTransaction(ctx context.Context, f func(ctx context.Context) error) error
	AddItem(ctx context.Context, item entity.Item) error
	UpdateItem(ctx context.Context, item entity.Item) error
	GetItem(ctx context.Context, itemID string, opts ...Opts) (entity.Item, error)
	ReserveItem(ctx context.Context, orderID uint64, item entity.Item) (entity.Reservation, error)
	DeleteReservationItem(ctx context.Context, itemID uint64) error
	GetReservationItem(ctx context.Context, itemID uint64) (entity.Reservation, error)
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
	WithdrawMoney(ctx context.Context, userID uint64, amount uint64) error
	CreditMoney(ctx context.Context, userID uint64, amount uint64) error
}

type InventoryGateway interface {
	MakeReservation(ctx context.Context, orderID uint64, itemID string, amount uint64) (uint64, error)
	CancelReservation(ctx context.Context, reservationID uint64) error
}

type DeliveryGateway interface {
	ApplyCourierForOrder(ctx context.Context, orderID uint64) (entity.Delivery, error)
}

type OrderRepo interface {
	CreateOrder(ctx context.Context, user entity.UserShort, price uint64, item string, status entity.Status) (entity.Order, error)
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

type DeliveryRepository interface {
	GetCouriersOnShiftWithoutActiveDeliveries(ctx context.Context) ([]entity.Courier, error)
	AddDelivery(ctx context.Context, delivery entity.Delivery) (entity.Delivery, error)
	GetDeliveries(ctx context.Context) ([]entity.Delivery, error)
	GetDeliveryByID(ctx context.Context, id uint64) (entity.Delivery, error)
}
