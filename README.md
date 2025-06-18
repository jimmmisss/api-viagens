# API de Viagens

Este projeto é uma API RESTful para gerenciamento de pedidos de viagens, desenvolvida em Go seguindo os princípios de Arquitetura Hexagonal (Ports and Adapters).

## Arquitetura

O projeto segue a Arquitetura Hexagonal, que separa claramente as responsabilidades em camadas:

### Domínio

- **Model**: Contém as entidades de negócio (TravelRequest, User) e suas regras.
- **Ports**: Define as interfaces (portas) que conectam o domínio com o mundo exterior.

### Aplicação

- **Service**: Implementa os casos de uso da aplicação, orquestrando as entidades de domínio.

### Infraestrutura

- **Adapters**: Implementa as interfaces definidas nas portas.
  - **In**: Adaptadores de entrada (HTTP, CLI).
  - **Out**: Adaptadores de saída (MySQL, Logger).
- **Config**: Configurações da aplicação.

## Funcionalidades

- Criação, consulta e atualização de pedidos de viagem
- Autenticação e autorização baseada em JWT
- Diferentes níveis de acesso (usuário comum e gerente)
- Regras de negócio para aprovação e cancelamento de viagens

## Tecnologias Utilizadas

- Go (Golang)
- MySQL para persistência de dados
- JWT para autenticação
- Arquitetura Hexagonal para organização do código

## Modificações Realizadas

Durante o desenvolvimento, foram implementadas as seguintes partes:

1. **Domínio**:
   - Definição das entidades de negócio (TravelRequest, User)
   - Definição das interfaces (portas) para repositórios e serviços

2. **Aplicação**:
   - Implementação dos serviços que orquestram as regras de negócio

3. **Infraestrutura**:
   - Implementação dos adaptadores HTTP para entrada
   - Implementação dos adaptadores MySQL para persistência
   - Implementação de autenticação JWT
   - Configuração da aplicação

4. **Correções**:
   - Correção de erros no arquivo main.go
   - Ajustes nos imports e dependências
   - Implementação de componentes faltantes

## Como Executar

1. Clone o repositório
2. Configure as variáveis de ambiente ou use os valores padrão
3. Execute o comando: `go run cmd/server/main.go`

## Estrutura de Diretórios

```
api-viagens/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── domain/
│   │   ├── model/
│   │   │   ├── request.go
│   │   │   └── user.go
│   │   └── ports/
│   │       ├── repositories.go
│   │       └── services.go
│   ├── application/
│   │   └── service/
│   │       └── travel_service.go
│   └── infrastructure/
│       ├── adpters/
│       │   ├── in/
│       │   │   └── http/
│       │   │       ├── handler.go
│       │   │       ├── middleware.go
│       │   │       └── router.go
│       │   └── out/
│       │       ├── auth/
│       │       │   └── jwt_auth.go
│       │       ├── logger/
│       │       │   └── notifier.go
│       │       └── mysql/
│       │           ├── db.go
│       │           ├── models.go
│       │           ├── repository.go
│       │           └── queries/
│       │               ├── request.sql
│       │               └── user.sql
│       └── config/
│           └── config.go
```

## Endpoints da API

- `POST /api/auth/login`: Autenticação de usuários
- `POST /api/auth/register`: Registro de novos usuários
- `POST /api/requests`: Criação de pedidos de viagem
- `GET /api/requests`: Listagem de pedidos de viagem
- `GET /api/requests/{id}`: Detalhes de um pedido de viagem
- `PUT /api/manager/requests/status`: Atualização de status de pedidos (apenas gerentes)

## Próximos Passos

- Implementar testes automatizados
- Adicionar documentação Swagger
- Implementar cache para melhorar performance
- Adicionar monitoramento e logging mais avançados