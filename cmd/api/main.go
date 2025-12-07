package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/Guilherme-G-Cadilhe/Go-LedgerFlow-Banking-API-Microservices/internal/gateway"
	"github.com/Guilherme-G-Cadilhe/Go-LedgerFlow-Banking-API-Microservices/internal/infra/http/handler"
	internalMiddleware "github.com/Guilherme-G-Cadilhe/Go-LedgerFlow-Banking-API-Microservices/internal/infra/http/middleware"
	"github.com/Guilherme-G-Cadilhe/Go-LedgerFlow-Banking-API-Microservices/internal/infra/postgres"
	"github.com/Guilherme-G-Cadilhe/Go-LedgerFlow-Banking-API-Microservices/internal/infra/rabbitmq"
	redisInfra "github.com/Guilherme-G-Cadilhe/Go-LedgerFlow-Banking-API-Microservices/internal/infra/redis"
	"github.com/Guilherme-G-Cadilhe/Go-LedgerFlow-Banking-API-Microservices/internal/usecase"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	// 1. Configuração de Logs (Zerolog - estruturado e rápido)
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr}) // Log bonito no terminal

	// O erro é ignorado de propósito, pois em Produção (Docker/K8s)
	// não usamos arquivo .env, usamos variáveis reais do sistema.
	if err := godotenv.Load(); err != nil {
		log.Warn().Msg("Arquivo .env não encontrado, usando variáveis de ambiente do sistema")
	}
	ctx := context.Background()

	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASSWORD")
	dbHost := "localhost" // Em docker seria o nome do service, local é localhost
	if os.Getenv("DB_HOST") != "" {
		dbHost = os.Getenv("DB_HOST")
	}
	dbName := os.Getenv("DB_NAME")

	dbURL := fmt.Sprintf("postgres://%s:%s@%s:5432/%s?sslmode=disable", dbUser, dbPass, dbHost, dbName)
	// Fallback para dev local se as envs não estiverem setadas
	if dbUser == "" {
		dbURL = "postgres://ledger:secret123@localhost:5432/ledgerflow?sslmode=disable"
	}

	dbPool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatal().Err(err).Msg("Não foi possível conectar ao banco de dados")
	}
	defer dbPool.Close()

	if err := dbPool.Ping(ctx); err != nil {
		log.Fatal().Err(err).Msg("Banco de dados não está respondendo")
	}
	log.Info().Msg("✅ Conectado ao PostgreSQL com sucesso!")

	redisHost := os.Getenv("REDIS_HOST")
	if redisHost == "" {
		redisHost = "localhost"
	}
	redisClient := redis.NewClient(&redis.Options{
		Addr: redisHost + ":6379",
	})
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Warn().Err(err).Msg("Não foi possível conectar ao Redis (Idempotência desabilitada)")
	} else {
		log.Info().Msg("✅ Conectado ao Redis!")
	}

	rabbitUser := os.Getenv("RABBITMQ_USER")
	rabbitPass := os.Getenv("RABBITMQ_PASS")
	rabbitHost := os.Getenv("RABBITMQ_HOST")
	if rabbitHost == "" {
		rabbitHost = "localhost"
	} // Fallback local

	rabbitURL := fmt.Sprintf("amqp://%s:%s@%s:5672/", rabbitUser, rabbitPass, rabbitHost)
	rabbitConn, err := amqp.DialConfig(rabbitURL, amqp.Config{
		Properties: amqp.Table{
			"connection_name": "LedgerAPI_Publisher", // <--- O Nome Mágico
		},
	})
	if err != nil {
		log.Warn().Err(err).Msg("Falha ao conectar no RabbitMQ (Eventos não serão enviados)")
	} else {
		defer rabbitConn.Close()
		log.Info().Msg("✅ Conectado ao RabbitMQ!")
	}

	var eventPublisher gateway.EventPublisher
	if rabbitConn != nil {
		ch, err := rabbitConn.Channel()
		if err != nil {
			log.Fatal().Err(err).Msg("Falha ao abrir canal RabbitMQ")
		}
		defer ch.Close()

		// Declarar Exchange (Tópico)
		err = ch.ExchangeDeclare(
			"ledger_events", // name
			"topic",         // type
			true,            // durable
			false,           // auto-deleted
			false,           // internal
			false,           // no-wait
			nil,             // arguments
		)
		if err != nil {
			log.Fatal().Err(err).Msg("Falha ao declarar Exchange")
		}

		eventPublisher = rabbitmq.NewRabbitMQPublisher(ch)
	}

	// Inicialização da Camada de Infraestrutura (Repositories)
	idempotencyRepo := redisInfra.NewIdempotencyRepository(redisClient)
	walletRepository := postgres.NewWalletRepository(dbPool)
	transactionRepository := postgres.NewTransactionRepository(dbPool)
	//  Unit of Work (Gerenciador de Transações)
	uow := postgres.NewUow(dbPool)

	// Inicialização da Camada de UseCase (Regras de Negócio)
	transferUseCase := usecase.NewTransferMoney(walletRepository, transactionRepository, uow, eventPublisher)
	createWalletUseCase := usecase.NewCreateWallet(walletRepository)

	// Handlers
	transferHandler := handler.NewTransferHandler(transferUseCase)
	walletHandler := handler.NewWalletHandler(createWalletUseCase)

	// Configuração do Servidor HTTP (Router Chi)
	router := chi.NewRouter()

	// Middlewares básicos
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer) // Evita crash se der panic
	router.Use(middleware.Timeout(60 * time.Second))
	idempotencyMiddleware := internalMiddleware.Idempotency(idempotencyRepo)

	// Rota de Health Check (para o Docker saber se estamos vivos)
	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("OK")); err != nil {
			log.Error().Err(err).Msg("Falha ao escrever resposta de health check")
		}
	})

	// Rotas
	router.Group(func(r chi.Router) {
		r.Use(idempotencyMiddleware)
		r.Post("/transfers", transferHandler.Create)
	})
	router.Post("/wallets", walletHandler.Create)

	// 6. Subir o Servidor
	port := ":8080"
	log.Info().Msgf("🚀 Servidor rodando na porta %s", port)
	if err := http.ListenAndServe(port, router); err != nil {
		log.Fatal().Err(err).Msg("Falha ao iniciar servidor HTTP")
	}
}
