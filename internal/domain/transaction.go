package domain

import "time"

type Transaction struct {
	ID             string
	FromWalletID   int64
	ToWalletID     int64
	Amount         int64
	Status         string
	IdempotencyKey *string
	CreatedAt      time.Time
}
