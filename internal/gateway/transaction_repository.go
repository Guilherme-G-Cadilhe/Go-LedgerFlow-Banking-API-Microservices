package gateway

import (
	"context"

	"github.com/Guilherme-G-Cadilhe/Go-LedgerFlow-Banking-API-Microservices/internal/domain"
)

type TransactionRepository interface {
	Create(ctx context.Context, transaction *domain.Transaction) error
	// WithTx segue o mesmo padrão da Wallet para participar da transação atômica
	WithTx(tx TransactionObject) TransactionRepository
}
