package domain

import "errors"

var (
	ErrLoginAlreadyTaken      = errors.New("login already taken")
	ErrNotFound               = errors.New("not found")
	ErrInvalidLoginOrPassword = errors.New("invalid login or password")
	ErrOrderRegisteredByYou   = errors.New("order already registered by you")
	ErrOrderRegisteredByOther = errors.New("order already registered by other user")
	ErrInsufficientFunds      = errors.New("insufficient funds")
)
