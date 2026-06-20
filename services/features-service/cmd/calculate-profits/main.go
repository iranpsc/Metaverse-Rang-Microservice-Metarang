package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	"metargb/features-service/internal/repository"
	"metargb/features-service/internal/service"
	"metargb/shared/pkg/logger"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	log := logger.NewLogger("calculate-hourly-profits")

	dbDSN := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4&collation=utf8mb4_unicode_ci",
		getEnv("DB_USER", "metargb_user"),
		getEnv("DB_PASSWORD", "metargb_password"),
		getEnv("DB_HOST", "mysql"),
		getEnv("DB_PORT", "3306"),
		getEnv("DB_DATABASE", "metargb_db"),
	)

	database, err := sql.Open("mysql", dbDSN)
	if err != nil {
		log.Fatal("Failed to connect to database", "error", err)
	}
	defer database.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	if err := database.PingContext(ctx); err != nil {
		log.Fatal("Failed to ping database", "error", err)
	}

	hourlyProfitRepo := repository.NewHourlyProfitRepository(database)
	profitService := service.NewProfitService(
		hourlyProfitRepo,
		nil,
		nil,
		nil,
		nil,
		database,
		log,
	)

	updated, err := profitService.RunHourlyProfitCalculation(ctx)
	if err != nil {
		log.Fatal("Hourly profit calculation failed", "error", err)
	}

	log.Info("Hourly profit calculation finished", "updated_records", updated)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
