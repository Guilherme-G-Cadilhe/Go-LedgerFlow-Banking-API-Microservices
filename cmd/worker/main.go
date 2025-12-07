package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Arquivo .env não encontrado, usando variáveis de ambiente")
	}

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
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Erro ao abrir canal: %v", err)
	}
	defer ch.Close()

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

	log.Printf(" [*] Worker iniciado. Aguardando mensagens na fila %s...", q.Name)

	go func() {
		for d := range msgs {
			log.Printf(" [x] Mensagem Recebida: %s", d.Body)
			// Aqui no futuro entra a lógica de salvar no MongoDB
		}
	}()

	// Graceful Shutdown (Bloqueia a main até receber sinal)
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

	<-stopChan // <--- O programa fica parado AQUI até você dar Ctrl+C

	log.Println("Shutting down worker...")

}
