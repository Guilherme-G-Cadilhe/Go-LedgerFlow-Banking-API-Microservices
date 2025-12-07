package postgres

import (
	"context"
	"fmt"

	"github.com/Guilherme-G-Cadilhe/Go-LedgerFlow-Banking-API-Microservices/internal/domain"
	"github.com/Guilherme-G-Cadilhe/Go-LedgerFlow-Banking-API-Microservices/internal/gateway"
	"github.com/Guilherme-G-Cadilhe/Go-LedgerFlow-Banking-API-Microservices/internal/infra/postgres/db"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TransactionRepository struct {
	db      *pgxpool.Pool
	queries *db.Queries
}

func NewTransactionRepository(pool *pgxpool.Pool) *TransactionRepository {
	return &TransactionRepository{
		db:      pool,
		queries: db.New(pool),
	}
}

func (r *TransactionRepository) Create(ctx context.Context, tx *domain.Transaction) error {
	// Conversão do domínio para o formato do SQLC
	params := db.CreateTransactionParams{
		FromWalletID: tx.FromWalletID,
		ToWalletID:   tx.ToWalletID,
		Amount:       tx.Amount,
		Status:       tx.Status,
		// IdempotencyKey é *string no domínio, mas pgtype.Text no banco
		IdempotencyKey: textToPgType(tx.IdempotencyKey),
	}

	row, err := r.queries.CreateTransaction(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to create transaction: %w", err)
	}

	// Atualiza o ID e CreatedAt gerados pelo banco de volta no objeto de domínio
	tx.ID = row.ID.String() // UUID to String
	tx.CreatedAt = row.CreatedAt.Time

	return nil
}

func (r *TransactionRepository) WithTx(tx gateway.TransactionObject) gateway.TransactionRepository {
	pgTx, ok := tx.(pgx.Tx)
	if !ok {
		return r
	}
	return &TransactionRepository{
		db:      r.db,
		queries: r.queries.WithTx(pgTx),
	}
}

// Helper para converter *string -> pgtype.Text
func textToPgType(s *string) pgtype.Text {
	if s == nil {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: *s, Valid: true}
}
