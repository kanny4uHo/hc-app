package entity

type Courier struct {
	ID        uint64
	Name      string
	IsOnShift bool
}

type DeliveryStatus string

const (
	DeliveryStatusOnTheWay  DeliveryStatus = "on_the_way"
	DeliveryStatusDelivered DeliveryStatus = "delivered"
)

type Delivery struct {
	ID        uint64
	OrderID   uint64
	CourierID uint64
	Status    DeliveryStatus
}
