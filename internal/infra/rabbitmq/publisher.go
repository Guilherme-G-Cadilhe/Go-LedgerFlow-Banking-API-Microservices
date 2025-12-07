package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"
)

type RabbitMQPublisher struct {
	channel *amqp.Channel
}

func NewRabbitMQPublisher(ch *amqp.Channel) *RabbitMQPublisher {
	return &RabbitMQPublisher{channel: ch}
}

func (p *RabbitMQPublisher) Publish(ctx context.Context, exchange, routingKey string, body interface{}) error {
	bytes, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	err = p.channel.PublishWithContext(ctx,
		exchange,   // exchange
		routingKey, // routing key
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         bytes,
			DeliveryMode: amqp.Persistent, // Garante que a mensagem não suma se o Rabbit reiniciar
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	log.Info().Str("routing_key", routingKey).Msg("Evento publicado no RabbitMQ")
	return nil
}
