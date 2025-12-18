package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Guilherme-G-Cadilhe/Go-LedgerFlow-Banking-API-Microservices/internal/domain"
	"github.com/Guilherme-G-Cadilhe/Go-LedgerFlow-Banking-API-Microservices/internal/usecase"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
)

type WalletHandler struct {
	createWalletUC *usecase.CreateWalletUseCase
	getWalletUC    *usecase.GetWalletUseCase
}

func NewWalletHandler(
	createWalletUC *usecase.CreateWalletUseCase,
	getWalletUC *usecase.GetWalletUseCase,
) *WalletHandler {
	return &WalletHandler{
		createWalletUC: createWalletUC,
		getWalletUC:    getWalletUC,
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

func (h *WalletHandler) Get(w http.ResponseWriter, r *http.Request) {
	// Extrair ID da URL (usando Chi Router)
	idStr := chi.URLParam(r, "id")
	walletID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "ID da carteira inválido")
		return
	}

	output, err := h.getWalletUC.Execute(r.Context(), walletID)
	if err != nil {
		if err == domain.ErrWalletNotFound {
			respondError(w, http.StatusNotFound, "Carteira não encontrada")
			return
		}
		log.Error().Err(err).Msg("Erro ao buscar carteira")
		respondError(w, http.StatusInternalServerError, "Erro interno")
		return
	}

	respondJSON(w, http.StatusOK, output)
}
