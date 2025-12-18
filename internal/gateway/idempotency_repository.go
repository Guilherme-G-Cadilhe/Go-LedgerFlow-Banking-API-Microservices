package gateway

import (
	"context"
	"time"
)

// Representa o que salvamos no Redis
type CachedResponse struct {
	StatusCode int
	Body       []byte
	Headers    map[string][]string // bom para headers customizados
}

type IdempotencyRepository interface {
	// Get retorna a resposta cacheada se existir. Erro se não existir.
	Get(ctx context.Context, key string) (*CachedResponse, error)

	// Save armazena a resposta com um TTL (Time To Live)
	Save(ctx context.Context, key string, response CachedResponse, ttl time.Duration) error
}
