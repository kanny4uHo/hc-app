package entity

type Item struct {
	ItemID string
	Amount uint64
}

type Reservation struct {
	ID      uint64
	OrderID uint64
	Item    Item
}
