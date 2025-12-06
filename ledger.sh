#!/bin/bash

# Cores
GREEN='\033[0;32m'
CYAN='\033[0;36m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

show_help() {
    echo -e "${CYAN}Comandos Disponíveis:${NC}"
    echo -e "  ${GREEN}setup${NC}     -> Instala dependências (Go tools, SQLC, Migrate)"
    echo -e "  ${GREEN}up${NC}        -> Sobe os containers (Docker Compose)"
    echo -e "  ${GREEN}down${NC}      -> Derruba os containers"
    echo -e "  ${GREEN}sqlc${NC}      -> Gera código Go a partir das queries SQL"
    echo -e "  ${GREEN}migrate${NC}   -> Roda as migrações do banco de dados"
    echo -e "  ${GREEN}test${NC}      -> Roda testes unitários"
    echo -e "  ${GREEN}test-int${NC}  -> Roda testes de integração"
    echo -e "  ${GREEN}run-api${NC}   -> Roda a API localmente"
}

case "$1" in
    setup)
        echo -e "${YELLOW}📦 Instalando ferramentas...${NC}"
        go mod download
        go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
        go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
        echo -e "${GREEN}✅ Setup concluído!${NC}"
        ;;
    up)
        echo -e "${YELLOW}🚀 Subindo ambiente...${NC}"
        docker-compose up -d
        echo -e "${GREEN}✅ Ambiente online!${NC}"
        ;;
    down)
        echo -e "${YELLOW}🛑 Parando ambiente...${NC}"
        docker-compose down
        ;;
    sqlc)
        echo -e "${YELLOW}⚙️ Gerando código SQLC...${NC}"
        sqlc generate
        ;;
    migrate)
        echo -e "${YELLOW}🐘 Rodando migrações...${NC}"
        DB_URL="postgresql://ledger:secret123@localhost:5432/ledgerflow?sslmode=disable"
        migrate -path ./migrations -database "$DB_URL" up
        ;;
    test)
        echo -e "${YELLOW}🧪 Rodando testes unitários...${NC}"
        go test -v -race -cover ./...
        ;;
    test-int)
        echo -e "${YELLOW}🧪 Rodando testes de integração...${NC}"
        go test -v -tags=integration ./tests/integration/...
        ;;
    run-api)
        echo -e "${YELLOW}🔌 Iniciando API...${NC}"
        go run cmd/api/main.go
        ;;
    *)
        show_help
        ;;
esac