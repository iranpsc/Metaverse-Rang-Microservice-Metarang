package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"

	"metargb/social-service/internal/client"
	"metargb/social-service/internal/handler"
	"metargb/social-service/internal/repository"
	"metargb/social-service/internal/service"
)

func main() {
	configPaths := []string{
		"config.env",
		"./config.env",
		"../config.env",
		"../../config.env",
		"services/social-service/config.env",
	}
	var configLoaded bool
	for _, configPath := range configPaths {
		if err := godotenv.Load(configPath); err == nil {
			configLoaded = true
			break
		}
	}
	if !configLoaded {
		log.Printf("Warning: config.env not found, using environment variables only")
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4&collation=utf8mb4_unicode_ci",
		getEnv("DB_USER", "root"),
		getEnv("DB_PASSWORD", ""),
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_PORT", "3306"),
		getEnv("DB_DATABASE", "metargb_db"),
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("Successfully connected to database")

	challengeRepo := repository.NewChallengeRepository(db)
	followRepo := repository.NewFollowRepository(db)
	userRepo := repository.NewUserRepository(db)

	var commercialClient client.CommercialClient
	commercialAddr := getEnv("COMMERCIAL_SERVICE_ADDR", "")
	if commercialAddr != "" {
		c, cerr := client.NewCommercialClient(commercialAddr)
		if cerr != nil {
			log.Printf("Warning: could not connect commercial client (PSC prizes disabled): %v", cerr)
		} else {
			commercialClient = c
			defer func() { _ = c.Close() }()
		}
	} else {
		log.Printf("Warning: COMMERCIAL_SERVICE_ADDR not set — challenge prize credits will be skipped")
	}

	challengeSvc := service.NewChallengeService(challengeRepo, commercialClient)
	followSvc := service.NewFollowService(followRepo, userRepo)

	grpcServer := grpc.NewServer()
	handler.RegisterChallengeHandler(grpcServer, challengeSvc)
	handler.RegisterFollowHandler(grpcServer, followSvc)

	port := getEnv("GRPC_PORT", "50061")
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", port, err)
	}

	log.Printf("Social service listening on port %s", port)

	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	grpcServer.GracefulStop()
	log.Println("Server stopped")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
