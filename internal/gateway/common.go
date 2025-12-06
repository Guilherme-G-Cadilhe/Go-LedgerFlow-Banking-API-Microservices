package gateway

import "context"

// TransactionObject é o "crachá" opaco que carrega a transação do banco
type TransactionObject interface{}

// TransactionManager define quem sabe iniciar/comitar transações (UoW)
type TransactionManager interface {
	Run(ctx context.Context, fn func(ctx context.Context) error) error
}

// TransactionKeyType evita colisão de chaves no contexto
type TransactionKeyType string

const TransactionKey TransactionKeyType = "transaction"
