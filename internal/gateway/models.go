package gateway

type EventType int

const OrderCreated EventType = 1

type OrderEvent struct {
	Type EventType `json:"type"`
	ID   uint64    `json:"id"`
}

type NewUserEvent struct {
	ID uint64 `json:"id"`
}

type OrderPaymentFailedEvent struct {
	OrderID uint64 `json:"order_id"`
}

type OrderIsPaidEvent struct {
	OrderID uint64 `json:"order_id"`
}
