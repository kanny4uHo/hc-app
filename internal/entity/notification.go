package entity

import "time"

type Notification struct {
	ID        uint64
	Timestamp time.Time
	Recipient string
	Message   string
}
