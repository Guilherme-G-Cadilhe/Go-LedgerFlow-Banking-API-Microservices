package usecase

import (
	"context"
	"fmt"

	"github.com/Guilherme-G-Cadilhe/Go-LedgerFlow-Banking-API-Microservices/internal/domain"
	"github.com/Guilherme-G-Cadilhe/Go-LedgerFlow-Banking-API-Microservices/internal/gateway"
)

type GetWalletOutput struct {
	ID        int64  `json:"id"`
	Balance   int64  `json:"balance"`
	UpdatedAt string `json:"updated_at"`
}

type GetWalletUseCase struct {
	walletRepository gateway.WalletRepository
}

func NewGetWallet(walletRepo gateway.WalletRepository) *GetWalletUseCase {
	return &GetWalletUseCase{
		walletRepository: walletRepo,
	}
}

func (u *GetWalletUseCase) Execute(ctx context.Context, walletID int64) (*GetWalletOutput, error) {
	wallet, err := u.walletRepository.GetByID(ctx, walletID)
	if err != nil {
		// Se for erro de "não encontrado", retornamos o erro de domínio
		if err == domain.ErrWalletNotFound {
			return nil, err
		}
		// Outros erros (banco fora do ar, etc)
		return nil, fmt.Errorf("erro ao buscar carteira: %w", err)
	}

	return &GetWalletOutput{
		ID:        wallet.ID,
		Balance:   wallet.Balance,
		UpdatedAt: wallet.UpdatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}
