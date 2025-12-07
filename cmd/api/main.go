package main

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/Guilherme-G-Cadilhe/Go-LedgerFlow-Banking-API-Microservices/internal/infra/postgres"
	"github.com/Guilherme-G-Cadilhe/Go-LedgerFlow-Banking-API-Microservices/internal/usecase"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	// 1. Configuração de Logs (Zerolog - estruturado e rápido)
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr}) // Log bonito no terminal

	ctx := context.Background()

	// 2. Conexão com Banco de Dados (Postgres via PGXPool)
	// Em prod, essa string viria de variável de ambiente (os.Getenv)
	dbURL := "postgresql://ledger:secret123@localhost:5432/ledgerflow?sslmode=disable"

	dbPool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatal().Err(err).Msg("Não foi possível conectar ao banco de dados")
	}
	defer dbPool.Close()

	// Verifica se o banco está respondendo
	if err := dbPool.Ping(ctx); err != nil {
		log.Fatal().Err(err).Msg("Banco de dados não está respondendo")
	}
	log.Info().Msg("✅ Conectado ao PostgreSQL com sucesso!")

	// 3. Inicialização da Camada de Infraestrutura (Repositories)
	walletRepository := postgres.NewWalletRepository(dbPool)
	transactionRepository := postgres.NewTransactionRepository(dbPool)

	// O Unit of Work (Gerenciador de Transações) também precisa do pool
	uow := postgres.NewUow(dbPool)

	// 4. Inicialização da Camada de UseCase (Regras de Negócio)
	// Injetamos os repositórios e o transaction manager aqui
	usecase.NewTransferMoney(
		walletRepository,
		transactionRepository,
		uow,
	)

	// 5. Configuração do Servidor HTTP (Router Chi)
	router := chi.NewRouter()

	// Middlewares básicos
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer) // Evita crash se der panic
	router.Use(middleware.Timeout(60 * time.Second))

	// Rota de Health Check (para o Docker saber se estamos vivos)
	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("OK")); err != nil {
			log.Error().Err(err).Msg("Falha ao escrever resposta de health check")
		}
	})

	// Rota de Transferência (Vamos criar um handler dedicado depois, por enquanto inline para teste)
	// Em breve moveremos isso para internal/infra/http/handler
	router.Post("/transfers", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
		if _, err := w.Write([]byte("Endpoint de transferências será implementado no próximo passo")); err != nil {
			log.Error().Err(err).Msg("Falha ao escrever resposta de transfers")
		}
	})

	// 6. Subir o Servidor
	port := ":8080"
	log.Info().Msgf("🚀 Servidor rodando na porta %s", port)
	if err := http.ListenAndServe(port, router); err != nil {
		log.Fatal().Err(err).Msg("Falha ao iniciar servidor HTTP")
	}
}
