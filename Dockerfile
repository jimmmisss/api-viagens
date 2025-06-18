# Etapa 1: Build
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copia os arquivos de dependências e baixa os módulos
COPY go.mod go.sum ./
RUN go mod download

# Copia o resto do código fonte
COPY . .

# Compila a aplicação
# -o /app/server: output do binário
# CGO_ENABLED=0: desabilita CGO para uma compilação estática
# -ldflags="-w -s": remove informações de debug para um binário menor
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/server ./cmd/server

# Etapa 2: Execução
FROM alpine:latest

WORKDIR /app

# Copia as migrations para a imagem final
COPY --from=builder /app/migrations ./migrations

# Copia o binário compilado da etapa de build
COPY --from=builder /app/server .

# Expõe a porta que a aplicação vai rodar
EXPOSE 8080

# Comando para executar a aplicação
CMD ["/app/server"]