package domain

import (
	"time"
)

// Wallet representa a carteira do usuário.
// Clean Architecture: Esta entidade não sabe o que é JSON nem SQL.
type Wallet struct {
	ID        int64
	Balance   int64
	Version   int32 // Para controle de concorrência otimista (se necessário)
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Métodos de domínio (Lógica pura)

// HasSufficientFunds valida se a carteira pode pagar antes mesmo de tocar no DB
func (w *Wallet) HasSufficientFunds(amount int64) bool {
	return w.Balance >= amount
}

func (w *Wallet) Debit(amount int64) error {
	if amount <= 0 {
		return ErrInvalidAmount
	}
	if !w.HasSufficientFunds(amount) {
		return ErrInsufficientFunds
	}
	w.Balance -= amount
	return nil
}

func (w *Wallet) Credit(amount int64) {
	if amount <= 0 {
		return
	}
	w.Balance += amount
}
