# ledger.ps1 - Gerenciador de Comandos do LedgerFlow
param (
    [string]$Command = "help"
)

# Configuração de Cores
$Green = "Green"
$Cyan = "Cyan"
$Yellow = "Yellow"

function Show-Help {
    Write-Host "Comandos Disponíveis:" -ForegroundColor $Cyan
    Write-Host "  setup     " -NoNewline -ForegroundColor $Green; Write-Host " -> Instala dependências (Go tools, SQLC, Migrate)"
    Write-Host "  up        " -NoNewline -ForegroundColor $Green; Write-Host " -> Sobe os containers (Docker Compose)"
    Write-Host "  down      " -NoNewline -ForegroundColor $Green; Write-Host " -> Derruba os containers"
    Write-Host "  sqlc      " -NoNewline -ForegroundColor $Green; Write-Host " -> Gera código Go a partir das queries SQL"
    Write-Host "  migrate   " -NoNewline -ForegroundColor $Green; Write-Host " -> Roda as migrações do banco de dados"
    Write-Host "  test      " -NoNewline -ForegroundColor $Green; Write-Host " -> Roda testes unitários"
    Write-Host "  test-int  " -NoNewline -ForegroundColor $Green; Write-Host " -> Roda testes de integração"
    Write-Host "  run-api   " -NoNewline -ForegroundColor $Green; Write-Host " -> Roda a API localmente"
}

switch ($Command) {
    "setup" {
        Write-Host "📦 Instalando ferramentas..." -ForegroundColor $Yellow
        go mod download
        go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
        go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
        Write-Host "✅ Setup concluído!" -ForegroundColor $Green
    }
    "up" {
        Write-Host "🚀 Subindo ambiente..." -ForegroundColor $Yellow
        docker-compose up -d
        Write-Host "✅ Ambiente online!" -ForegroundColor $Green
    }
    "down" {
        Write-Host "🛑 Parando ambiente..." -ForegroundColor $Yellow
        docker-compose down
    }
    "sqlc" {
        Write-Host "⚙️ Gerando código SQLC..." -ForegroundColor $Yellow
        sqlc generate
    }
    "migrate" {
        Write-Host "🐘 Rodando migrações..." -ForegroundColor $Yellow
        # Certifique-se que o postgres está rodando
        $DB_URL = "postgresql://ledger:secret123@localhost:5432/ledgerflow?sslmode=disable"
        migrate -path migrations -database $DB_URL up
    }
    "test" {
        Write-Host "🧪 Rodando testes unitários..." -ForegroundColor $Yellow
        go test -v -race -cover ./...
    }
    "test-int" {
        Write-Host "🧪 Rodando testes de integração..." -ForegroundColor $Yellow
        go test -v -tags=integration ./tests/integration/...
    }
    "run-api" {
        Write-Host "🔌 Iniciando API..." -ForegroundColor $Yellow
        go run cmd/api/main.go
    }
    Default {
        Show-Help
    }
}