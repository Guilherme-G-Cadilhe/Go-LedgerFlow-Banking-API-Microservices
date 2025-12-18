### Setup

> Executar ./ledger.sh

### Run Local Go Lint

Install: https://golangci-lint.run/docs/welcome/install/#local-installation

> golangci-lint run

## Banco Manual

### Conectar no Postgres do Docker e criar duas carteiras

Wallet 1: Saldo 1000 (R$ 10,00)
Wallet 2: Saldo 0

> docker exec -it ledgerflow-postgres psql -U ledger -d ledgerflow -c "INSERT INTO wallets (balance) VALUES (1000), (0) RETURNING id, balance;"

### Verificar dinheiro

> docker exec -it ledgerflow-postgres psql -U ledger -d ledgerflow -c "SELECT id, balance FROM wallets;"

### Conectar no MongoDB para ver logs

docker exec -it ledgerflow-mongodb mongosh -u ledger -p secret123

# Dentro do shell do mongo:

use ledgerflow_audit
db.audit_logs.find()

### INFRA

Rabbitmq: http://localhost:15672
user: ledger
Pass: secret123
