package main

import (
	"log"
	"net/http"
	"strconv"

	"github.com/jimmmmisss/api-viagens/internal/application/service"
	httpHandler "github.com/jimmmmisss/api-viagens/internal/infrastructure/adapters/in/http"
	"github.com/jimmmmisss/api-viagens/internal/infrastructure/adapters/out/auth"
	"github.com/jimmmmisss/api-viagens/internal/infrastructure/adapters/out/logger"
	"github.com/jimmmmisss/api-viagens/internal/infrastructure/adapters/out/mysql"
	"github.com/jimmmmisss/api-viagens/internal/infrastructure/config"
)

func main() {
	cfg := config.LoadConfig()

	// 1. Conexão com o banco de dados
	db, err := mysql.NewDBConnection(cfg.Database.GetDSN())
	if err != nil {
		log.Fatalf("Erro ao conectar ao banco de dados: %v", err)
	}
	defer db.Close()

	// 2. Inicialização dos adaptadores
	sqlcRepo := mysql.NewSQLCRepository(db)
	consoleNotifier := logger.NewConsoleNotifier()
	jwtAuthService := auth.NewJWTAuthService(cfg.JWT.Secret, int(cfg.JWT.Expiration.Hours()), sqlcRepo)

	// 3. Inicialização do serviço de aplicação
	travelService := service.NewTravelService(sqlcRepo, consoleNotifier)

	// 4. Inicialização do roteador e handlers
	mainRouter := httpHandler.SetupRouter(travelService, jwtAuthService, sqlcRepo)

	// 5. Inicialização do servidor HTTP
	serverAddr := ":" + strconv.Itoa(cfg.Server.Port)
	log.Printf("🚀 Servidor da API de Viagens iniciado em http://localhost%s", serverAddr)

	server := &http.Server{
		Addr:         serverAddr,
		Handler:      mainRouter,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Não foi possível iniciar o servidor: %s\n", err)
	}
}
