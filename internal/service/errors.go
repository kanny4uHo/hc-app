package service

import "fmt"

var ErrUserNotFound = fmt.Errorf("user is not found")
var ErrNotEnoughMoney = fmt.Errorf("not enough money")

type InvalidArgumentError struct {
	Field  string
	Value  interface{}
	Reason string
}

func (e InvalidArgumentError) Error() string {
	return fmt.Sprintf("invalid argument %s=%v: %s", e.Field, e.Value, e.Reason)
}
