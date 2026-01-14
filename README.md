# Product Redis API

API de gerenciamento de produtos com cache Redis integrado, desenvolvida em Go usando Clean Architecture.

## Características Principais

- **Arquitetura Limpa (Clean Architecture)**: Separação clara de responsabilidades entre domínio, aplicação e infraestrutura
- **Cache Redis Inteligente**: Write-through strategy com índices otimizados para busca rápida
- **PostgreSQL**: Banco de dados relacional com optimistic locking para concorrência
- **ULID Determinístico**: IDs únicos gerados a partir de nome + número de referência
- **Observabilidade Completa**: Logs estruturados, métricas Prometheus, health checks
- **Production-Ready**: Graceful shutdown, timeouts configuráveis, CORS, middleware de segurança

## Arquitetura

```
├── cmd/
│   └── api/           # Entry point da aplicação
├── internal/
│   ├── domain/        # Entidades e interfaces (sem dependências)
│   ├── application/   # Casos de uso (regras de negócio)
│   └── infrastructure/
│       ├── database/  # Implementação PostgreSQL
│       ├── cache/     # Implementação Redis
│       ├── http/      # Handlers, middleware, rotas
│       ├── config/    # Configuração
│       └── logger/    # Logging estruturado
```

## Pré-requisitos

- Go 1.23+
- Docker e Docker Compose
- PostgreSQL 15+
- Redis 7+

## Configuração Rápida

### 1. Clone e configure o ambiente

```bash
# Copie o arquivo de exemplo
cp .env.example .env

# Edite as variáveis se necessário
# As configurações padrão já funcionam para desenvolvimento local
```

### 2. Inicie os serviços (PostgreSQL e Redis)

```bash
docker-compose up -d
```

### 3. Instale as dependências

```bash
go mod download
```

### 4. Prepare o banco de dados

Conecte ao PostgreSQL e execute as seguintes queries para criar a tabela de produtos:

```sql
-- Criar tabela de produtos
CREATE TABLE IF NOT EXISTS products (
    id VARCHAR(26) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    reference_number VARCHAR(100) NOT NULL UNIQUE,
    category VARCHAR(100) NOT NULL,
    description TEXT,
    sku VARCHAR(100),
    brand VARCHAR(100),
    stock INTEGER NOT NULL DEFAULT 0,
    images TEXT[],
    specifications JSONB,
    version INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Índices para otimização de buscas
CREATE INDEX IF NOT EXISTS idx_products_name ON products USING GIN (to_tsvector('portuguese', name));
CREATE INDEX IF NOT EXISTS idx_products_category ON products (category);
CREATE INDEX IF NOT EXISTS idx_products_reference ON products (reference_number);
CREATE INDEX IF NOT EXISTS idx_products_created_at ON products (created_at DESC);
```

### 5. Inicie a API

```bash
go run cmd/api/main.go
```

A API estará disponível em `http://localhost:8080`

## Estrutura do Produto

Um produto possui os seguintes campos:

```json
{
  "id": "01HN8Z9QXXX...",           // ULID gerado (name + reference_number)
  "name": "Notebook Dell XPS 15",   // Nome do produto
  "reference_number": "NB-DELL-001", // Número de referência
  "category": "Electronics",         // Categoria
  "description": "High-end laptop",  // Descrição
  "sku": "DELL-XPS15-2024",         // SKU
  "brand": "Dell",                   // Marca
  "stock": 100,                      // Estoque
  "images": [                        // URLs de imagens
    "https://example.com/img1.jpg"
  ],
  "specifications": {                // Especificações customizadas
    "cpu": "Intel i7",
    "ram": "32GB"
  },
  "version": 1,                      // Versão (optimistic locking)
  "created_at": "2024-01-15T10:00:00Z",
  "updated_at": "2024-01-15T10:00:00Z"
}
```

**Nota sobre precificação**: Por design, o preço NÃO faz parte deste serviço. Em sistemas enterprise, pricing é tipicamente um serviço separado devido a complexidade de regras de negócio, mudanças frequentes e requisitos de auditoria.

## Endpoints da API

### Health Checks

```bash
# Liveness probe (Kubernetes)
GET /health/live

# Readiness probe (verifica DB + Redis)
GET /health/ready
```

### Produtos

#### Criar Produto

```bash
POST /api/v1/products
Content-Type: application/json

{
  "name": "Notebook Dell XPS 15",
  "reference_number": "NB-DELL-001",
  "category": "Electronics",
  "description": "High-end laptop for professionals",
  "sku": "DELL-XPS15-2024",
  "brand": "Dell",
  "stock": 100,
  "images": [
    "https://example.com/dell-xps15-front.jpg",
    "https://example.com/dell-xps15-side.jpg"
  ],
  "specifications": {
    "cpu": "Intel Core i7-12700H",
    "ram": "32GB DDR5",
    "storage": "1TB NVMe SSD",
    "screen": "15.6\" 4K OLED"
  }
}
```

**Lógica de Negócio**:
1. Gera ULID a partir de `name + reference_number`
2. Verifica se já existe no Redis
3. Se existe e é idêntico, ignora (retorna o existente)
4. Se existe e é diferente, retorna erro 409
5. Se não existe, salva no PostgreSQL
6. Se salvamento OK, atualiza cache Redis e índices

#### Atualizar Produto

```bash
PUT /api/v1/products/{id}
Content-Type: application/json

{
  "name": "Notebook Dell XPS 15",
  "category": "Electronics",
  "description": "Updated description",
  "sku": "DELL-XPS15-2024",
  "brand": "Dell",
  "stock": 95,
  "images": [...],
  "specifications": {...}
}
```

**Lógica de Negócio**:
1. Busca produto atual (cache ou DB)
2. Compara propriedades
3. Se idêntico, ignora (retorna o existente)
4. Se diferente, atualiza no PostgreSQL com optimistic locking
5. Se atualização OK, atualiza cache e índices (se categoria/nome mudou)

#### Deletar Produto

```bash
DELETE /api/v1/products/{id}
```

**Lógica de Negócio**:
1. Busca produto para obter metadados
2. Deleta do PostgreSQL
3. Remove do cache Redis e de todos os índices

#### Buscar por ID

```bash
GET /api/v1/products/{id}
```

**Lógica de Negócio**:
1. Busca no Redis primeiro
2. Se não encontrar, busca no PostgreSQL
3. Se encontrou no PostgreSQL, popula o cache

#### Listar Todos

```bash
GET /api/v1/products?limit=50&offset=0
```

**Lógica de Negócio**:
1. Tenta buscar IDs do set `all_products` no Redis
2. Busca produtos do cache usando os IDs
3. Se cache miss ou parcial, busca do PostgreSQL
4. Popula cache assincronamente se veio do DB

#### Buscar por Nome (Busca Preditiva)

```bash
GET /api/v1/products/search/name?q=dell&limit=20&offset=0
```

**Lógica de Negócio**:
1. Busca IDs no set `product_by_name_dell` do Redis
2. Busca produtos do cache usando os IDs
3. Se cache miss, busca do PostgreSQL com `LIKE`
4. Popula cache assincronamente

#### Buscar por Categoria

```bash
GET /api/v1/products/search/category?q=electronics&limit=20&offset=0
```

**Lógica de Negócio**:
1. Busca IDs no set `product_by_category_electronics` do Redis
2. Busca produtos do cache usando os IDs
3. Se cache miss, busca do PostgreSQL
4. Popula cache assincronamente

## Estratégia de Cache Redis

### Estrutura de Chaves

```
product_{ulid}                     # Produto individual (JSON)
all_products                       # Set com todos os IDs
product_by_name_{name}             # Set com IDs por nome
product_by_category_{category}     # Set com IDs por categoria
```

### Write-Through sem TTL

- Cache é atualizado simultaneamente com o banco
- Sem expiração automática (TTL = 0)
- Invalidação manual em updates/deletes
- Mais consistente, ideal quando CPU de DB é mais caro que memória Redis

### Resilência

- Falhas no Redis NÃO matam operações
- Sistema continua funcionando via PostgreSQL
- Logs de warning para falhas de cache
- Cache é sempre best-effort, nunca crítico

## Optimistic Locking

Para prevenir conflitos de concorrência:

```go
// Cada produto tem um campo version
type Product struct {
    Version int `json:"version"`
    // ...
}

// No update, verificamos a versão
UPDATE products
SET ... version = version + 1
WHERE id = $1 AND version = $2
```

Se a versão não bate, retorna erro 409 (Conflict).

## Observabilidade

### Logs Estruturados

```bash
# Logs em formato JSON (production) ou colorido (development)
{"level":"info","msg":"http request","method":"GET","path":"/api/v1/products","status":200}
```

### Log Level Dinâmico

O nível de log pode ser alterado em tempo de execução sem restart:

```bash
# Consultar nível atual
GET /log/level

# Alterar para debug
PUT /log/level
Content-Type: application/json
{"level": "debug"}

# Níveis disponíveis: debug, info, warn, error, dpanic, panic, fatal
```

Útil para troubleshooting em produção sem necessidade de restart.

### Métricas Prometheus

```bash
# Endpoint de métricas
GET /metrics

# Exemplo de métricas disponíveis:
# - go_goroutines
# - http_request_duration_seconds
# - http_requests_total
# etc.
```

### Health Checks

```bash
# Verifica se a API está viva
curl http://localhost:8080/health/live

# Verifica se dependências estão OK
curl http://localhost:8080/health/ready
# Retorna 200 se tudo OK, 503 se algum serviço está down
```

## Desenvolvimento

### Rodando Testes

```bash
# Todos os testes
go test ./...

# Com cobertura
go test -cover ./...

# Gerar relatório de cobertura HTML
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

### Estrutura de Testes

```
internal/
├── domain/
│   └── entity/
│       └── product_test.go        # Testes unitários de entidade
├── application/
│   └── usecase/
│       └── create_product_test.go # Testes de casos de uso
└── infrastructure/
    ├── database/
    │   └── postgres_repository_test.go
    └── cache/
        └── redis_repository_test.go
```

## Exemplo de Uso Completo

```bash
# 1. Criar um produto
curl -X POST http://localhost:8080/api/v1/products \
  -H "Content-Type: application/json" \
  -d '{
    "name": "iPhone 15 Pro",
    "reference_number": "APL-IP15P-001",
    "category": "Smartphones",
    "description": "Latest iPhone with A17 Pro chip",
    "sku": "APPLE-IP15P-256-TIT",
    "brand": "Apple",
    "stock": 50,
    "images": ["https://example.com/iphone15pro.jpg"],
    "specifications": {
      "storage": "256GB",
      "color": "Titanium",
      "chip": "A17 Pro"
    }
  }'

# Resposta (201 Created):
# {
#   "id": "01HN8Z9QXX...",
#   "name": "iPhone 15 Pro",
#   "version": 1,
#   ...
# }

# 2. Buscar o produto
curl http://localhost:8080/api/v1/products/01HN8Z9QXX...

# 3. Atualizar estoque
curl -X PUT http://localhost:8080/api/v1/products/01HN8Z9QXX... \
  -H "Content-Type: application/json" \
  -d '{
    "name": "iPhone 15 Pro",
    "category": "Smartphones",
    "description": "Latest iPhone with A17 Pro chip",
    "sku": "APPLE-IP15P-256-TIT",
    "brand": "Apple",
    "stock": 45,
    "images": ["https://example.com/iphone15pro.jpg"],
    "specifications": {
      "storage": "256GB",
      "color": "Titanium",
      "chip": "A17 Pro"
    }
  }'

# 4. Buscar por categoria
curl "http://localhost:8080/api/v1/products/search/category?q=smartphones&limit=10"

# 5. Buscar por nome (busca preditiva)
curl "http://localhost:8080/api/v1/products/search/name?q=iphone&limit=10"

# 6. Listar todos
curl "http://localhost:8080/api/v1/products?limit=50&offset=0"

# 7. Deletar produto
curl -X DELETE http://localhost:8080/api/v1/products/01HN8Z9QXX...
```

## Variáveis de Ambiente

Veja [.env.example](.env.example) para todas as variáveis disponíveis.

Principais:

```bash
# Server
SERVER_PORT=8080

# PostgreSQL
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=@Pass2025
DB_NAME=products_db

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=@Pass2025
REDIS_DB=0

# Application
LOG_LEVEL=info
ENVIRONMENT=development
```

## Deployment

### Docker

```dockerfile
# Exemplo de Dockerfile
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 go build -o api cmd/api/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/api .
EXPOSE 8080
CMD ["./api"]
```

### Kubernetes

```yaml
# Health checks configurados
livenessProbe:
  httpGet:
    path: /health/live
    port: 8080
  initialDelaySeconds: 10
  periodSeconds: 10

readinessProbe:
  httpGet:
    path: /health/ready
    port: 8080
  initialDelaySeconds: 5
  periodSeconds: 5
```

## Segurança

- Sem preço no modelo de produto (segregação de responsabilidades)
- Validação de entrada em todos os endpoints
- Logs estruturados sem informações sensíveis
- Timeouts configuráveis para prevenir ataques
- CORS configurável
- Graceful shutdown para não perder requisições

## Performance

- Cache Redis reduz carga no PostgreSQL
- Índices otimizados no PostgreSQL (GIN para full-text, B-tree para categorias)
- Connection pooling configurável
- Paginação em todos os endpoints de listagem
- Pipeline Redis para operações em batch

## Limitações Conhecidas

- Busca por nome usa `LIKE` no PostgreSQL (não é full-text search avançado)
- Cache Redis não tem TTL (requer mais memória)
- Sem autenticação/autorização (deve ser adicionado via API Gateway)
- Sem rate limiting (deve ser adicionado via API Gateway)

## Próximos Passos Sugeridos

1. Adicionar testes de integração
2. Implementar full-text search com Elasticsearch
3. Adicionar CI/CD pipeline
4. Implementar cache warming strategy
5. Adicionar OpenAPI/Swagger documentation
6. Implementar event sourcing para auditoria
7. Adicionar GraphQL endpoint
8. Implementar sharding de cache por categoria

## Contribuindo

1. Fork o projeto
2. Crie uma branch para sua feature (`git checkout -b feature/AmazingFeature`)
3. Commit suas mudanças (`git commit -m 'Add some AmazingFeature'`)
4. Push para a branch (`git push origin feature/AmazingFeature`)
5. Abra um Pull Request

## Licença

MIT License

## Suporte

Para dúvidas ou problemas, abra uma issue no GitHub.
