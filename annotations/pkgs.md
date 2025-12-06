# Web Framework & Router

go get -u github.com/go-chi/chi/v5

# Database Drivers

go get -u github.com/lib/pq
go get -u github.com/jackc/pgx/v5 # Recomendado para SQLC moderno

# Infra Clients

go get -u github.com/redis/go-redis/v9
go get -u github.com/rabbitmq/amqp091-go
go get -u go.mongodb.org/mongo-driver/mongo

# Observability & Logs

go get -u github.com/rs/zerolog
go get -u go.opentelemetry.io/otel
go get -u go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp

# Testing

go get -u github.com/stretchr/testify
go get -u github.com/ory/dockertest/v3

# CI

go install github.com/evilmartians/lefthook/v2@v2.0.8

> lefthook install
