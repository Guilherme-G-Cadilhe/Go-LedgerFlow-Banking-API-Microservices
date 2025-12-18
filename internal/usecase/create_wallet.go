package usecase

import (
	"context"

	"github.com/Guilherme-G-Cadilhe/Go-LedgerFlow-Banking-API-Microservices/internal/gateway"
)

type CreateWalletInput struct {
	Balance int64
}

type CreateWalletOutput struct {
	ID      int64
	Balance int64
}

type CreateWalletUseCase struct {
	walletRepo gateway.WalletRepository
}

func NewCreateWallet(walletRepo gateway.WalletRepository) *CreateWalletUseCase {
	return &CreateWalletUseCase{
		walletRepo: walletRepo,
	}
}

func (uc *CreateWalletUseCase) Execute(ctx context.Context, input CreateWalletInput) (*CreateWalletOutput, error) {
	// A criação de wallet é uma operação atômica simples (um insert),
	// então não precisamos abrir uma transação complexa (Begin/Commit) aqui,
	// a menos que tivéssemos que salvar eventos ou outras coisas juntas.
	wallet, err := uc.walletRepo.Create(ctx, input.Balance)
	if err != nil {
		return nil, err
	}

	return &CreateWalletOutput{
		ID:      wallet.ID,
		Balance: wallet.Balance,
	}, nil
}
