package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/Guilherme-G-Cadilhe/Go-LedgerFlow-Banking-API-Microservices/internal/domain"
	"github.com/Guilherme-G-Cadilhe/Go-LedgerFlow-Banking-API-Microservices/internal/usecase"
	"github.com/rs/zerolog/log"
)

// TransferHandler expõe as operações de transferência via HTTP
type TransferHandler struct {
	transferUseCase *usecase.TransferMoneyUseCase
}

// NewTransferHandler cria uma nova instância
func NewTransferHandler(uc *usecase.TransferMoneyUseCase) *TransferHandler {
	return &TransferHandler{
		transferUseCase: uc,
	}
}

// DTOs (Data Transfer Objects) para Request/Response
// Usamos tags JSON para mapear snake_case (padrão de APIs)
type CreateTransferRequest struct {
	FromWalletID int64 `json:"from_wallet_id"`
	ToWalletID   int64 `json:"to_wallet_id"`
	Amount       int64 `json:"amount"` // Valor em centavos
}

type CreateTransferResponse struct {
	TransactionID string `json:"transaction_id"`
	Status        string `json:"status"`
}

// Create processa a requisição de transferência
func (h *TransferHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req CreateTransferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Payload inválido")
		return
	}

	idempotencyKey := r.Header.Get("Idempotency-Key")
	fmt.Println(idempotencyKey)
	var idempotencyKeyPtr *string
	if idempotencyKey != "" {
		idempotencyKeyPtr = &idempotencyKey
	}
	fmt.Println(idempotencyKeyPtr)

	input := usecase.TransferMoneyInput{
		FromWalletID:   req.FromWalletID,
		ToWalletID:     req.ToWalletID,
		Amount:         req.Amount,
		IdempotencyKey: idempotencyKeyPtr,
	}

	output, err := h.transferUseCase.Execute(ctx, input)
	if err != nil {
		// Mapeamento de Erros de Domínio -> HTTP Status Code
		switch {
		case errors.Is(err, domain.ErrWalletNotFound):
			respondError(w, http.StatusNotFound, "Carteira não encontrada")
		case errors.Is(err, domain.ErrInsufficientFunds):
			respondError(w, http.StatusUnprocessableEntity, "Saldo insuficiente")
		case errors.Is(err, domain.ErrInvalidAmount):
			respondError(w, http.StatusBadRequest, "Valor inválido")
		default:
			// Erro interno (banco caiu, bug, etc)
			log.Error().Err(err).Msg("Erro interno ao processar transferência")
			respondError(w, http.StatusInternalServerError, "Erro interno do servidor")
		}
		return
	}

	respondJSON(w, http.StatusCreated, CreateTransferResponse{
		TransactionID: output.TransactionID,
		Status:        output.Status,
	})
}

// Helpers para resposta JSON
func respondJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Error().Err(err).Msg("Falha ao codificar resposta JSON")
	}
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}
