package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/Guilherme-G-Cadilhe/Go-LedgerFlow-Banking-API-Microservices/internal/domain"
	"github.com/Guilherme-G-Cadilhe/Go-LedgerFlow-Banking-API-Microservices/internal/gateway"
	"github.com/Guilherme-G-Cadilhe/Go-LedgerFlow-Banking-API-Microservices/internal/infra/postgres/db"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// WalletRepository implementa gateway.WalletRepository usando pgx/v5
type WalletRepository struct {
	db      *pgxpool.Pool //  Usamos pgxpool em vez de sql.DB
	queries *db.Queries
}

// NewWalletRepository cria uma nova instância
func NewWalletRepository(pool *pgxpool.Pool) *WalletRepository {
	return &WalletRepository{
		db:      pool,
		queries: db.New(pool),
	}
}

// Create insere uma nova carteira
func (r *WalletRepository) Create(ctx context.Context, balance int64) (*domain.Wallet, error) {
	modelWallet, err := r.queries.CreateWallet(ctx, balance)
	if err != nil {
		return nil, fmt.Errorf("failed to create wallet: %w", err)
	}
	return toDomainWallet(modelWallet), nil
}

// GetByID busca uma carteira
func (r *WalletRepository) GetByID(ctx context.Context, id int64) (*domain.Wallet, error) {
	modelWallet, err := r.queries.GetWallet(ctx, id)
	if err != nil {
		// pgx retorna pgx.ErrNoRows, diferente de sql.ErrNoRows
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrWalletNotFound
		}
		return nil, fmt.Errorf("failed to get wallet: %w", err)
	}
	return toDomainWallet(modelWallet), nil
}

// 🔐 Implementação do Lock
func (r *WalletRepository) GetByIDForUpdate(ctx context.Context, id int64) (*domain.Wallet, error) {
	// Chama a query com "FOR UPDATE"
	modelWallet, err := r.queries.GetWalletForUpdate(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrWalletNotFound
		}
		return nil, fmt.Errorf("failed to lock wallet: %w", err)
	}
	return toDomainWallet(modelWallet), nil
}

// 💸 Débito Atômico (Valida saldo no banco)
func (r *WalletRepository) Debit(ctx context.Context, id int64, amount int64) error {
	params := db.DebitWalletParams{
		Amount: amount,
		ID:     id,
	}

	// ExecRows retorna quantas linhas foram afetadas
	rowsAffected, err := r.queries.DebitWallet(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to debit wallet: %w", err)
	}

	// Se 0 linhas foram afetadas, a cláusula "AND balance >= amount" falhou
	if rowsAffected == 0 {
		return domain.ErrInsufficientFunds
	}

	return nil
}

// 💰 Crédito Atômico
func (r *WalletRepository) Credit(ctx context.Context, id int64, amount int64) error {
	params := db.CreditWalletParams{
		Amount: amount,
		ID:     id,
	}
	return r.queries.CreditWallet(ctx, params)
}

// WithTx retorna uma cópia do repositório usando uma transação específica
func (r *WalletRepository) WithTx(tx gateway.TransactionObject) gateway.WalletRepository {
	pgTx, ok := tx.(pgx.Tx)
	if !ok {
		return r
	}
	return &WalletRepository{
		db:      r.db,
		queries: r.queries.WithTx(pgTx),
	}
}

// Mapper: pgtype -> Go types
func toDomainWallet(w db.Wallet) *domain.Wallet {
	return &domain.Wallet{
		ID:      w.ID,
		Balance: w.Balance,
		Version: w.Version,
		//  pgtype.Timestamptz é uma struct, acessamos o valor .Time
		CreatedAt: w.CreatedAt.Time,
		UpdatedAt: w.UpdatedAt.Time,
	}
}
