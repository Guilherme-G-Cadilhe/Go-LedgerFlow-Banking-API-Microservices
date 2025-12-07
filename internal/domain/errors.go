package domain

import "errors"

var (
	ErrInsufficientFunds = errors.New("insufficient funds")
	ErrInvalidAmount     = errors.New("transaction amount must be greater than zero")
	ErrWalletNotFound    = errors.New("wallet not found")
	ErrTransactionFailed = errors.New("transaction failed")
	ErrIdempotencyKey    = errors.New("idempotency key conflict")
)
