#!/bin/bash

# Cores para logs bonitos
GREEN='\033[0;32m'
CYAN='\033[0;36m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# 1. Descobre onde este script está e onde é a raiz do projeto
# Isso garante que funcione rodando ./scripts/ledger.sh ou cd scripts && ./ledger.sh
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_ROOT="$SCRIPT_DIR/.."

# Ajuste do DB_URL (garantindo que não tenha espaços extras)
DB_URL="postgresql://ledger:secret123@localhost:5432/ledgerflow?sslmode=disable"

show_help() {
    echo -e "${CYAN}--- LedgerFlow CLI ---${NC}"
    echo -e "${CYAN}Uso: ./scripts/ledger.sh [comando]${NC}"
    echo ""
    echo -e "  ${GREEN}setup${NC}     -> Instala dependências (Go tools, SQLC, Migrate)"
    echo -e "  ${GREEN}up${NC}        -> Sobe o ambiente (Docker Compose)"
    echo -e "  ${GREEN}down${NC}      -> Derruba o ambiente"
    echo -e "  ${GREEN}sqlc${NC}      -> Gera código Go (SQLC)"
    echo -e "  ${GREEN}migrate${NC}   -> Roda as migrações (Up)"
    echo -e "  ${GREEN}migrate-down${NC} -> Desfaz a última migração"
    echo -e "  ${GREEN}test${NC}      -> Roda testes unitários"
    echo -e "  ${GREEN}test-int${NC}  -> Roda testes de integração"
    echo -e "  ${GREEN}run-api${NC}   -> Roda a API localmente"
}

case "$1" in
    setup)
        echo -e "${YELLOW}📦 Instalando ferramentas...${NC}"
        go mod download
        go get github.com/sqlc-dev/sqlc/cmd/sqlc@latest
        go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
        echo -e "${GREEN}✅ Setup concluído!${NC}"
        ;;
    up)
        echo -e "${YELLOW}🚀 Subindo ambiente...${NC}"
        # Usa -f para garantir que ele ache o docker-compose na raiz
        docker-compose -f "$PROJECT_ROOT/docker-compose.yml" up -d
        echo -e "${GREEN}✅ Ambiente online!${NC}"
        ;;
    down)
        echo -e "${YELLOW}🛑 Parando ambiente...${NC}"
        docker-compose -f "$PROJECT_ROOT/docker-compose.yml" down
        ;;
    sqlc)
        echo -e "${YELLOW}⚙️ Gerando código SQLC...${NC}"
        # Entra na raiz para rodar o sqlc, pois ele busca o sqlc.yaml
        cd "$PROJECT_ROOT" && sqlc generate
        ;;
    migrate)
        echo -e "${YELLOW}🐘 Rodando migrações (UP)...${NC}"
        # O prefixo file:// é obrigatório/recomendado em alguns ambientes
        # Usamos o caminho absoluto $PROJECT_ROOT/migrations
        migrate -path "$PROJECT_ROOT/migrations" -database "$DB_URL" up
        ;;
    migrate-down)
        echo -e "${YELLOW}🐘 Revertendo migração (DOWN)...${NC}"
        migrate -path "$PROJECT_ROOT/migrations" -database "$DB_URL" down 1
        ;;
    test)
        echo -e "${YELLOW}🧪 Rodando testes unitários...${NC}"
        cd "$PROJECT_ROOT" && go test -v -race -cover ./...
        ;;
    test-int)
        echo -e "${YELLOW}🧪 Rodando testes de integração...${NC}"
        cd "$PROJECT_ROOT" && go test -v -tags=integration ./tests/integration/...
        ;;
    run-api)
        echo -e "${YELLOW}🔌 Iniciando API...${NC}"
        cd "$PROJECT_ROOT" && go run cmd/api/main.go
        ;;
    *)
        show_help
        ;;
esac