package mongodb

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

// AuditLog representa o documento que será salvo no Mongo.
// Usamos tags 'bson' em vez de 'json'.
type AuditLog struct {
	ID            string    `bson:"_id,omitempty"` // O Mongo gera automático se vazio, ou usamos o UUID da msg
	TransactionID string    `bson:"transaction_id"`
	FromWallet    int64     `bson:"from_wallet"`
	ToWallet      int64     `bson:"to_wallet"`
	Amount        int64     `bson:"amount"`
	Status        string    `bson:"status"`
	ProcessedAt   time.Time `bson:"processed_at"`
}

type AuditRepository struct {
	collection *mongo.Collection
}

func NewAuditRepository(client *mongo.Client, dbName string) *AuditRepository {
	// Cria/Obtém a collection "audit_logs"
	collection := client.Database(dbName).Collection("audit_logs")
	return &AuditRepository{collection: collection}
}

func (r *AuditRepository) Save(ctx context.Context, log AuditLog) error {
	// Adiciona timestamp de processamento
	log.ProcessedAt = time.Now()

	// InsertOne salva o documento
	_, err := r.collection.InsertOne(ctx, log)
	if err != nil {
		return fmt.Errorf("failed to insert audit log: %w", err)
	}
	return nil
}
