# API de Viagens

## Descrição do Projeto

API de Viagens é um sistema de gerenciamento de solicitações de viagens desenvolvido em Go. O sistema permite que usuários se registrem, solicitem viagens, e gerentes aprovem ou cancelem essas solicitações, seguindo regras de negócio específicas.

## Arquitetura

O projeto segue uma arquitetura limpa (Clean Architecture) com separação clara de responsabilidades:

- **Domain**: Contém as entidades de negócio e regras de validação
- **Service**: Implementa a lógica de negócio e orquestra as operações
- **Repository**: Gerencia o acesso e persistência de dados
- **Handler**: Expõe as APIs REST e gerencia as requisições HTTP

```
api-viagens/
├── cmd/
│   └── api/           # Ponto de entrada da aplicação
├── internal/
│   ├── config/        # Configurações da aplicação
│   ├── domain/        # Entidades e regras de negócio
│   ├── handler/       # Handlers HTTP
│   ├── middleware/    # Middlewares (autenticação, etc.)
│   ├── repository/    # Implementações de acesso a dados
│   ├── service/       # Lógica de negócio
│   └── utils/         # Utilitários
├── migrations/        # Scripts de migração do banco de dados
└── docker-compose.yml # Configuração para execução em containers
```

## Tecnologias Utilizadas

- **Go**: Linguagem de programação principal
- **Gin**: Framework web para APIs REST
- **PostgreSQL**: Banco de dados relacional
- **Docker**: Containerização da aplicação
- **JWT**: Autenticação baseada em tokens

## Regras de Negócio

### Usuários
- Usuários devem se registrar com nome, email e senha
- Email deve ser único e em formato válido
- Senha deve ter no mínimo 8 caracteres

### Viagens
- Uma viagem possui destino, data de início e data de fim
- A data de fim deve ser posterior à data de início
- Uma viagem pode ter os status: solicitado, aprovado ou cancelado
- Apenas gerentes (não o solicitante) podem aprovar viagens
- Viagens aprovadas só podem ser canceladas pelo solicitante
- Viagens não podem ser canceladas se a data de início for em menos de 7 dias

### Notificações
- Usuários recebem notificações quando suas viagens são aprovadas ou canceladas

## Como Executar

### Pré-requisitos
- Docker e Docker Compose instalados

### Passos para Execução

1. Clone o repositório:
```bash
git clone https://github.com/seu-usuario/api-viagens.git
cd api-viagens
```

2. Inicie a aplicação com Docker Compose:
```bash
docker compose up -d --build
```

3. A API estará disponível em `http://localhost:8080`

> **Importante**: Ao fazer alterações no código, é necessário reconstruir os containers usando a flag `--build` para que as mudanças sejam aplicadas.

## Endpoints da API

### Autenticação
- `POST /register` - Registrar novo usuário
- `POST /login` - Autenticar usuário e obter token JWT

### Viagens
- `POST /trips` - Criar nova solicitação de viagem
- `GET /trips` - Listar viagens do usuário (com filtros opcionais)
- `GET /trips/:id` - Obter detalhes de uma viagem específica
- `PATCH /trips/:id/status` - Atualizar status da viagem (aprovar/cancelar)
- `POST /trips/:id/cancel` - Cancelar uma viagem aprovada

## Estrutura do Banco de Dados

### Tabela de Usuários
```sql
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

### Tabela de Viagens
```sql
CREATE TABLE IF NOT EXISTS trips (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    requester_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    destination VARCHAR(255) NOT NULL,
    start_date TIMESTAMPTZ NOT NULL,
    end_date TIMESTAMPTZ NOT NULL,
    status trip_status NOT NULL DEFAULT 'solicitado',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT dates_check CHECK (end_date > start_date)
);
```

## Testes

O projeto inclui testes unitários e de integração. Para executar os testes:

```bash
cd api-viagens
go test ./...
```

Para executar testes específicos:

```bash
go test ./internal/domain/... -v  # Testes de domínio
go test ./internal/service/... -v  # Testes de serviço
go test ./internal/handler/... -v  # Testes de handlers
```

## Desenvolvimento

Para desenvolvimento local sem Docker:

1. Configure um banco PostgreSQL local
2. Ajuste as variáveis de ambiente ou o arquivo de configuração
3. Execute as migrações do banco de dados
4. Inicie a aplicação:
```bash
go run cmd/api/main.go
```

## Contribuição

1. Faça um fork do projeto
2. Crie uma branch para sua feature (`git checkout -b feature/nova-feature`)
3. Commit suas mudanças (`git commit -m 'Adiciona nova feature'`)
4. Push para a branch (`git push origin feature/nova-feature`)
5. Abra um Pull Request