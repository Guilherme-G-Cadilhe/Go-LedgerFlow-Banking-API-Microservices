package postgres

import (
	"context"
	"fmt"

	"github.com/Guilherme-G-Cadilhe/Go-LedgerFlow-Banking-API-Microservices/internal/gateway"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Uow implementa gateway.TransactionManager
type Uow struct {
	pool *pgxpool.Pool
}

func NewUow(pool *pgxpool.Pool) *Uow {
	return &Uow{pool: pool}
}

// Run executa uma função dentro de uma transação ACID.
// Se a função retornar erro, faz Rollback. Se sucesso, Commit.
func (u *Uow) Run(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := u.pool.BeginTx(ctx, pgx.TxOptions{
		IsoLevel: pgx.ReadCommitted, // Ou Serializable para proteção máxima
	})
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Defer Rollback: Se commit não for chamado (pânico ou erro), garante rollback
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	// Injeta a transação no contexto (opcional) ou podemos passar via closure.
	// Aqui vamos usar um truque: passaremos a TX como valor no contexto para
	// que os repositórios possam recuperá-la se necessário, OU (design atual)
	// o UseCase orquestra isso chamando WithTx.

	// Design Simplificado para este projeto:
	// Vamos injetar a tx no contexto sob uma chave específica?
	// Não, vamos usar a chave específica do nosso gateway.
	ctxWithTx := context.WithValue(ctx, gateway.TransactionKey, tx)

	if err := fn(ctxWithTx); err != nil {
		return err // Rollback automático pelo defer
	}

	return tx.Commit(ctx)
}
