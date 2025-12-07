package gateway

import "context"

type EventPublisher interface {
	Publish(ctx context.Context, exchange, routingKey string, body interface{}) error
}
