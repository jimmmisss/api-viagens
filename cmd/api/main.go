package main

import (
	"context"
	"fmt"
	middleware "github.com/jimmmmisss/api-viagens/internal/midleware"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jimmmmisss/api-viagens/internal/config"
	"github.com/jimmmmisss/api-viagens/internal/handler"
	"github.com/jimmmmisss/api-viagens/internal/repository"
	"github.com/jimmmmisss/api-viagens/internal/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("could not load config: %v", err)
	}

	dbpool, err := connectDB(cfg)
	if err != nil {
		log.Fatalf("could not connect to database: %v", err)
	}
	defer dbpool.Close()

	// Setup dependencies
	userRepo := repository.NewPostgresUserRepository(dbpool)
	tripRepo := repository.NewPostgresTripRepository(dbpool)
	notificationSvc := service.NewLogNotificationService()
	userSvc := service.NewUserService(userRepo)
	tripSvc := service.NewTripService(tripRepo, userRepo, notificationSvc)
	h := handler.NewHandler(userSvc, tripSvc)

	// Setup Gin router
	router := setupRouter(h, cfg.JWTSecretKey)

	// Setup server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.APIPort),
		Handler: router,
	}

	// Graceful shutdown
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exiting")
}

func connectDB(cfg *config.Config) (*pgxpool.Pool, error) {
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName, cfg.DBSSLMode)

	dbpool, err := pgxpool.New(context.Background(), connStr)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	if err := dbpool.Ping(context.Background()); err != nil {
		dbpool.Close()
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}

	log.Println("Database connection successful")
	return dbpool, nil
}

func setupRouter(h *handler.Handler, jwtSecret string) *gin.Engine {
	r := gin.Default()

	// Public routes
	r.POST("/register", h.RegisterUser)
	r.POST("/login", h.LoginUser)

	// Authenticated routes
	authRoutes := r.Group("/")
	authMiddleware := middleware.AuthMiddleware(jwtSecret)
	authRoutes.Use(authMiddleware)
	{
		authRoutes.POST("/trips", h.CreateTrip)
		authRoutes.GET("/trips", h.ListTrips)
		authRoutes.GET("/trips/:id", h.GetTripByID)
		authRoutes.PATCH("/trips/:id/status", h.UpdateTripStatus)
		authRoutes.POST("/trips/:id/cancel", h.CancelApprovedTrip)
	}

	return r
}
