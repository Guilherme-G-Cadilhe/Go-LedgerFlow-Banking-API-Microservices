package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Guilherme-G-Cadilhe/Go-LedgerFlow-Banking-API-Microservices/internal/infra/mongodb"
	"github.com/joho/godotenv"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// Estrutura do evento que vem do RabbitMQ (JSON)
type TransactionEvent struct {
	TransactionID string `json:"transaction_id"`
	FromWallet    int64  `json:"from_wallet"`
	ToWallet      int64  `json:"to_wallet"`
	Amount        int64  `json:"amount"`
	Status        string `json:"status"`
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Arquivo .env não encontrado, usando variáveis de ambiente")
	}
	mongoUser := os.Getenv("MONGO_USER")
	mongoPass := os.Getenv("MONGO_PASS")
	// Em docker compose, o host é o nome do serviço 'mongodb'. Localmente, mapeamos porta.
	// Se rodar go run local, precisa ser localhost:27017
	mongoURI := "mongodb://" + mongoUser + ":" + mongoPass + "@localhost:27017"

	clientOptions := options.Client().ApplyURI(mongoURI)
	mongoClient, err := mongo.Connect(clientOptions)
	if err != nil {
		log.Fatalf("Erro ao criar client MongoDB: %v", err)
	}

	defer func() {
		if err := mongoClient.Disconnect(context.Background()); err != nil {
			log.Printf("Erro ao desconectar Mongo: %v", err)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// Verifica conexão
	if err := mongoClient.Ping(ctx, nil); err != nil {
		log.Fatalf("Erro ao pinar MongoDB: %v", err)
	}
	log.Println("✅ Conectado ao MongoDB!")
	auditRepo := mongodb.NewAuditRepository(mongoClient, "ledgerflow_audit")

	rabbitUser := os.Getenv("RABBITMQ_USER")
	rabbitPass := os.Getenv("RABBITMQ_PASS")
	rabbitHost := os.Getenv("RABBITMQ_HOST")
	if rabbitHost == "" {
		rabbitHost = "localhost"
	}

	rabbitURL := "amqp://" + rabbitUser + ":" + rabbitPass + "@" + rabbitHost + ":5672/"
	conn, err := amqp.DialConfig(rabbitURL, amqp.Config{
		Properties: amqp.Table{
			"connection_name": "AuditWorker_Consumer",
		},
	})
	if err != nil {
		log.Fatalf("Erro ao conectar no RabbitMQ: %v", err)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("Erro ao fechar conexão RabbitMQ: %v", err)
		}
	}()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Erro ao abrir canal: %v", err)
	}
	defer func() {
		if err := ch.Close(); err != nil {
			log.Printf("Erro ao fechar canal RabbitMQ: %v", err)
		}
	}()

	// Definir QoS (Prefetch Count = 1)
	// Isso garante que o RabbitMQ mande apenas 1 mensagem por vez e espere o Ack.
	// Resolve problemas de "travar" ou buffer encher.
	if err := ch.Qos(1, 0, false); err != nil {
		log.Fatalf("Erro ao configurar QoS: %v", err)
	}

	// Declarar a Exchange (Garantia de que ela existe, idempotente)
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
		log.Fatalf("Erro ao declarar exchange: %v", err)
	}

	// Declarar a Fila (QUEUE) - Onde as mensagens ficam guardadas
	q, err := ch.QueueDeclare(
		"audit_queue", // name
		true,          // durable (sobrevive a restart do server)
		false,         // delete when unused
		false,         // exclusive
		false,         // no-wait
		nil,           // arguments
	)
	if err != nil {
		log.Fatalf("Erro ao declarar fila: %v", err)
	}

	//  Bind (Amarração) - Ligar a Fila ao Exchange
	// "Tudo que começar com 'transaction.' vai para a 'audit_queue'"
	err = ch.QueueBind(
		q.Name,          // queue name
		"transaction.#", // routing key (# é curinga/wildcard)
		"ledger_events", // exchange
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Erro ao fazer bind da fila: %v", err)
	}

	// Iniciar Consumo
	msgs, err := ch.Consume(
		q.Name,         // queue
		"audit_worker", // consumer tag
		true,           // auto-ack (True para simplificar agora, em prod usamos manual ack)
		false,          // exclusive
		false,          // no-local
		false,          // no-wait
		nil,            // args
	)
	if err != nil {
		log.Fatalf("Erro ao registrar consumidor: %v", err)
	}

	// Monitoramento de queda de conexão
	notifyClose := make(chan *amqp.Error)
	ch.NotifyClose(notifyClose)

	log.Printf(" [*] Worker iniciado. Aguardando mensagens na fila %s...", q.Name)

	go func() {
		for {
			select {
			case err := <-notifyClose:
				if err != nil {
					log.Printf("🔴 Canal RabbitMQ fechado: %v", err)
					os.Exit(1) // Força o worker a cair para podermos reiniciar ou o Docker subir de novo
				}
				return
			case d, ok := <-msgs:
				if !ok {
					log.Println("🔴 Canal de mensagens fechado.")
					os.Exit(1)
				}

				log.Printf(" [⬇️] Recebido: %s", d.Body)

				var event TransactionEvent
				if err := json.Unmarshal(d.Body, &event); err != nil {
					log.Printf("Erro ao decodificar JSON: %v", err)
					// Linter Fix: Tratar erro do Nack
					if err := d.Nack(false, false); err != nil {
						log.Printf("Erro ao enviar Nack (JSON inválido): %v", err)
					}
					continue
				}

				auditLog := mongodb.AuditLog{
					TransactionID: event.TransactionID,
					FromWallet:    event.FromWallet,
					ToWallet:      event.ToWallet,
					Amount:        event.Amount,
					Status:        event.Status,
				}

				saveCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				if err := auditRepo.Save(saveCtx, auditLog); err != nil {
					log.Printf("Erro ao salvar no Mongo: %v", err)
					if err := d.Nack(false, true); err != nil {
						log.Printf("Erro ao enviar Nack (Mongo erro): %v", err)
					}
					cancel()
					continue
				}
				cancel()

				if err := d.Ack(false); err != nil {
					log.Printf("Erro ao enviar Ack: %v", err)
				}
				log.Println(" [✅] Salvo no MongoDB e Ack enviado.")
			}
		}
	}()

	// Graceful Shutdown (Bloqueia a main até receber sinal)
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

	<-stopChan // <--- O programa fica parado AQUI até você dar Ctrl+C

	log.Println("Shutting down worker...")

}
