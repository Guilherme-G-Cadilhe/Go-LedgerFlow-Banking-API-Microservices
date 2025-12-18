package usecase

import (
	"context"
	"fmt"

	"github.com/Guilherme-G-Cadilhe/Go-LedgerFlow-Banking-API-Microservices/internal/domain"
	"github.com/Guilherme-G-Cadilhe/Go-LedgerFlow-Banking-API-Microservices/internal/gateway"
)

// TransferMoneyInput define os dados necessários para realizar uma transferência.
// Usamos DTOs (Data Transfer Objects) para não acoplar a API HTTP ao UseCase.
type TransferMoneyInput struct {
	FromWalletID   int64
	ToWalletID     int64
	Amount         int64 // Valor em centavos (ex: 1000 = R$ 10,00)
	IdempotencyKey *string
}

// TransferMoneyOutput define o que devolvemos para quem chamou.
type TransferMoneyOutput struct {
	TransactionID string
	Status        string
}

// TransferMoneyUseCase contém as dependências necessárias.
type TransferMoneyUseCase struct {
	walletRepository      gateway.WalletRepository
	transactionRepository gateway.TransactionRepository
	transactionManager    gateway.TransactionManager // Nosso "Unit of Work"
	eventPublisher        gateway.EventPublisher
}

// NewTransferMoney cria uma nova instância do UseCase.
func NewTransferMoney(
	walletRepo gateway.WalletRepository,
	transactionRepo gateway.TransactionRepository,
	txManager gateway.TransactionManager,
	publisher gateway.EventPublisher,
) *TransferMoneyUseCase {
	return &TransferMoneyUseCase{
		walletRepository:      walletRepo,
		transactionRepository: transactionRepo,
		transactionManager:    txManager,
		eventPublisher:        publisher,
	}
}

// Execute roda a lógica de negócio.
func (u *TransferMoneyUseCase) Execute(ctx context.Context, input TransferMoneyInput) (*TransferMoneyOutput, error) {
	// Variável para capturar o resultado de dentro da transação
	var createdTransaction *domain.Transaction
	// Variáveis para o evento
	transactionStatus := "failed"
	var createdTransactionID string

	// Isso roda SEMPRE antes da função retornar, seja sucesso ou erro.
	defer func() {
		if u.eventPublisher != nil {
			event := map[string]interface{}{
				"transaction_id": createdTransactionID, // Pode estar vazio se falhar antes de criar
				"from_wallet":    input.FromWalletID,
				"to_wallet":      input.ToWalletID,
				"amount":         input.Amount,
				"status":         transactionStatus,
				"reason":         "", // Poderíamos adicionar a razão do erro aqui
			}

			// Define o tópico baseado no status
			routingKey := "transaction." + transactionStatus // transaction.created ou transaction.failed

			_ = u.eventPublisher.Publish(ctx, "ledger_events", routingKey, event)
		}
	}()

	// u.transactionManager.Run inicia uma transação no banco (BEGIN).
	// Se a função anônima retornar erro, ele faz ROLLBACK automático.
	// Se retornar nil, ele faz COMMIT.
	err := u.transactionManager.Run(ctx, func(contextWithTx context.Context) error {

		// Recuperar o "crachá" da transação que está dentro do contexto.
		// Isso foi injetado pelo TransactionManager.Run
		transactionObject := contextWithTx.Value(gateway.TransactionKey)
		if transactionObject == nil {
			return fmt.Errorf("erro crítico: transação não encontrada no contexto")
		}

		// Criar cópias dos repositórios que usam ESSA transação específica.
		// Agora, qualquer comando dado a 'walletRepoTx' rodará dentro do 'BEGIN...COMMIT'.
		walletRepoTx := u.walletRepository.WithTx(transactionObject)
		transactionRepoTx := u.transactionRepository.WithTx(transactionObject)

		// Ordenação de IDs para evitar Deadlock (Lock Pessimista)
		// Se a Transferência A->B e B->A acontecerem ao mesmo tempo,
		// ordenamos para que ambas travem sempre o ID menor primeiro.
		firstID, secondID := input.FromWalletID, input.ToWalletID
		if firstID > secondID {
			firstID, secondID = secondID, firstID
		}

		// Lock nas Carteiras (SELECT ... FOR UPDATE)
		// Isso faz o banco TRAVAR essas linhas. Ninguém mais mexe nelas até o Commit.
		_, err := walletRepoTx.GetByIDForUpdate(contextWithTx, firstID)
		if err != nil {
			return fmt.Errorf("falha ao travar carteira %d: %w", firstID, err)
		}

		_, err = walletRepoTx.GetByIDForUpdate(contextWithTx, secondID)
		if err != nil {
			return fmt.Errorf("falha ao travar carteira %d: %w", secondID, err)
		}

		// Operação de Débito (Quem envia)
		// O método Debit do repositório já verifica se tem saldo (balance >= amount).
		err = walletRepoTx.Debit(contextWithTx, input.FromWalletID, input.Amount)
		if err != nil {
			// Se falhar (saldo insuficiente), retornamos erro e o txManager faz Rollback.
			return fmt.Errorf("falha no débito (origem %d): %w", input.FromWalletID, err)
		}

		// Operação de Crédito (Quem recebe)
		err = walletRepoTx.Credit(contextWithTx, input.ToWalletID, input.Amount)
		if err != nil {
			return fmt.Errorf("falha no crédito (destino %d): %w", input.ToWalletID, err)
		}

		// Registrar o Histórico (Auditoria)
		createdTransaction = &domain.Transaction{
			FromWalletID:   input.FromWalletID,
			ToWalletID:     input.ToWalletID,
			Amount:         input.Amount,
			Status:         "completed", // Sucesso!
			IdempotencyKey: input.IdempotencyKey,
		}

		err = transactionRepoTx.Create(contextWithTx, createdTransaction)
		if err != nil {
			return fmt.Errorf("falha ao salvar histórico da transação: %w", err)
		}

		createdTransactionID = createdTransaction.ID
		transactionStatus = "completed"

		return nil // Sucesso! O Commit será executado agora.
	})

	if err != nil {
		return nil, err
	}

	return &TransferMoneyOutput{
		TransactionID: createdTransaction.ID,
		Status:        createdTransaction.Status,
	}, nil
}
