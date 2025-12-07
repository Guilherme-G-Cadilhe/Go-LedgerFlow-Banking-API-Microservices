package middleware

import (
	"bytes"
	"net/http"
	"time"

	"github.com/Guilherme-G-Cadilhe/Go-LedgerFlow-Banking-API-Microservices/internal/gateway"
	"github.com/rs/zerolog/log"
)

// responseRecorder é um "espião" que grava o que o handler escreve
type responseRecorder struct {
	http.ResponseWriter
	statusCode int
	body       *bytes.Buffer
}

func (r *responseRecorder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	r.body.Write(b)                  // Grava no nosso buffer
	return r.ResponseWriter.Write(b) // Manda pro cliente
}

func Idempotency(store gateway.IdempotencyRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := r.Header.Get("Idempotency-Key")
			if key == "" {
				// Se não tem chave, segue
				next.ServeHTTP(w, r)
				return
			}

			ctx := r.Context()

			// Verificar no Redis
			cached, err := store.Get(ctx, key)
			if err != nil {
				log.Error().Err(err).Msg("Falha ao buscar chave de idempotência")
				// Em caso de erro no Redis, deixamos passar para não travar a API (Fail Open)
				next.ServeHTTP(w, r)
				return
			}

			// Cache Hit: Retornar o que já tínhamos gravado
			if cached != nil {
				log.Info().Str("key", key).Msg("Idempotency cache hit")
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("X-Idempotency-Hit", "true")
				w.WriteHeader(cached.StatusCode)
				if _, err := w.Write(cached.Body); err != nil {
					log.Error().Err(err).Msg("Falha ao escrever resposta cacheada")
				}
				return
			}

			// Cache Miss: Processar a requisição e gravar a resposta
			recorder := &responseRecorder{
				ResponseWriter: w,
				statusCode:     http.StatusOK, // Default
				body:           &bytes.Buffer{},
			}

			next.ServeHTTP(recorder, r)

			// Salvar no Redis (Apenas sucessos 2xx ou erros de cliente 4xx que não devem mudar)
			// Erros 500 geralmente não queremos cachear para permitir retry.
			if recorder.statusCode < 500 {
				err := store.Save(ctx, key, gateway.CachedResponse{
					StatusCode: recorder.statusCode,
					Body:       recorder.body.Bytes(),
				}, 24*time.Hour) // TTL de 24h

				if err != nil {
					log.Error().Err(err).Msg("Falha ao salvar chave de idempotência")
				}
			}
		})
	}
}
