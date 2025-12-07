package handler

import (
	"encoding/json"
	"net/http"

	"github.com/Guilherme-G-Cadilhe/Go-LedgerFlow-Banking-API-Microservices/internal/usecase"
	"github.com/rs/zerolog/log"
)

type WalletHandler struct {
	createWalletUC *usecase.CreateWalletUseCase
	// Futuro: getWalletUC
}

func NewWalletHandler(createWalletUC *usecase.CreateWalletUseCase) *WalletHandler {
	return &WalletHandler{
		createWalletUC: createWalletUC,
	}
}

func (h *WalletHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Balance int64 `json:"balance"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Payload inválido")
		return
	}

	output, err := h.createWalletUC.Execute(r.Context(), usecase.CreateWalletInput{
		Balance: req.Balance,
	})
	if err != nil {
		log.Error().Err(err).Msg("Falha ao criar carteira")
		respondError(w, http.StatusInternalServerError, "Erro interno")
		return
	}

	respondJSON(w, http.StatusCreated, output)
}
