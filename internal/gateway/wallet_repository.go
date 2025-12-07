package gateway

import (
	"context"

	"github.com/Guilherme-G-Cadilhe/Go-LedgerFlow-Banking-API-Microservices/internal/domain"
)

// WalletRepository define o contrato para persistência de carteiras.
// O Usecase só interage com isso, sem saber se é Postgres ou MySQL.
type WalletRepository interface {
	Create(ctx context.Context, balance int64) (*domain.Wallet, error)
	GetByID(ctx context.Context, id int64) (*domain.Wallet, error)

	// Lock Pessimista: Retorna a wallet travando a linha no banco
	GetByIDForUpdate(ctx context.Context, id int64) (*domain.Wallet, error)
	// Métodos Atômicos
	Debit(ctx context.Context, id int64, amount int64) error
	Credit(ctx context.Context, id int64, amount int64) error

	// WithTx permite que o repositório participe de uma transação iniciada no nível superior
	// Retorna uma nova instância do repositório ligada àquela transação.
	// (Isso é um padrão avançado para lidar com Atomicidade no Clean Arch)
	WithTx(tx TransactionObject) WalletRepository
}
