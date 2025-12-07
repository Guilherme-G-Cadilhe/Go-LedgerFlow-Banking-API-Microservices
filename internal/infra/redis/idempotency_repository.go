package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Guilherme-G-Cadilhe/Go-LedgerFlow-Banking-API-Microservices/internal/gateway"
	"github.com/redis/go-redis/v9"
)

type IdempotencyRepository struct {
	client *redis.Client
}

func NewIdempotencyRepository(client *redis.Client) *IdempotencyRepository {
	return &IdempotencyRepository{client: client}
}

func (r *IdempotencyRepository) Get(ctx context.Context, key string) (*gateway.CachedResponse, error) {
	val, err := r.client.Get(ctx, "idempotency:"+key).Result()
	if errors.Is(err, redis.Nil) {
		return nil, nil // Não encontrado (cache miss)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get idempotency key: %w", err)
	}

	var resp gateway.CachedResponse
	if err := json.Unmarshal([]byte(val), &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cached response: %w", err)
	}

	return &resp, nil
}

func (r *IdempotencyRepository) Save(ctx context.Context, key string, response gateway.CachedResponse, ttl time.Duration) error {
	bytes, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}

	return r.client.Set(ctx, "idempotency:"+key, bytes, ttl).Err()
}
