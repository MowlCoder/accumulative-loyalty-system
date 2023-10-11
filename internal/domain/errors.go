package domain

import (
	"errors"
	"fmt"
)

var (
	ErrLoginAlreadyTaken                = errors.New("login already taken")
	ErrNotFound                         = errors.New("not found")
	ErrInvalidLoginOrPassword           = errors.New("invalid login or password")
	ErrOrderRegisteredByYou             = errors.New("order already registered by you")
	ErrOrderRegisteredByOther           = errors.New("order already registered by other user")
	ErrInsufficientFunds                = errors.New("insufficient funds")
	ErrMatchKeyAlreadyExists            = errors.New("match key already exists")
	ErrOrderAlreadyRegisteredForAccrual = errors.New("order already registered for accrual")
	ErrInternalServer                   = errors.New("internal server error")
)

type RetryAfterError struct {
	Seconds int
}

func (a RetryAfterError) Error() string {
	return fmt.Sprintf("retry after %d seconds", a.Seconds)
}
