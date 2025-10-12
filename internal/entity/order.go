package entity

type CreateOrderArgs struct {
	Item  string
	Price uint64
}

type Status string

const (
	StatusCreated  = "created"
	StatusPaid     = "paid"
	StatusComplete = "complete"
)

type Order struct {
	ID     uint64
	Price  uint64
	Item   string
	Status Status
	Owner  UserShort

	ReservationID uint64
	DeliveryID    uint64
}
