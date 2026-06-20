package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/eprac/eeip-backend/internal/application/services"
	"github.com/eprac/eeip-backend/internal/infrastructure/database"
	"github.com/eprac/eeip-backend/internal/interfaces/api/handlers"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	migrateOnly := flag.Bool("migrate-only", false, "Run migrations and exit")
	flag.Parse()

	dbHost := getEnv("DATABASE_HOST", "localhost")
	dbPort := getEnv("DATABASE_PORT", "5432")
	dbUser := getEnv("DATABASE_USER", "usreprac")
	dbPass := getEnv("DATABASE_PASSWORD", "password")
	dbName := getEnv("DATABASE_NAME", "eeip")

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", dbUser, dbPass, dbHost, dbPort, dbName)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	runMigrations(db)

	if *migrateOnly {
		log.Println("Migrations complete. Exiting due to -migrate-only flag.")
		return
	}

	router := gin.Default()
	router.Use(cors.Default())
	
	// Repositories
	sqlxDB := sqlx.NewDb(db, "postgres")
	emailRepo := database.NewEmailRepository(sqlxDB)
	accountRepo := database.NewAccountRepository(sqlxDB)
	stakeholderRepo := database.NewStakeholderRepository(sqlxDB)

	// Services
	openAIKey := getEnv("OPENAI_API_KEY", "mock-key")
	var aiEngine services.AIClassificationEngine
	if openAIKey == "mock-key" || openAIKey == "tu_openai_key_aqui" || openAIKey == "" {
		aiEngine = services.NewMockAIEngine()
	} else {
		aiEngine = services.NewAIClassificationEngine(openAIKey)
	}
	telegramToken := getEnv("TELEGRAM_BOT_TOKEN", "")
	telegramSvc := services.NewTelegramService(telegramToken)
	emailCollector := services.NewEmailCollector(emailRepo, aiEngine, stakeholderRepo, telegramSvc)
	summaryEngine := services.NewSummaryEngine(openAIKey)

	// Handlers
	emailHandler := handlers.NewEmailHandler(emailRepo, summaryEngine)
	accountHandler := handlers.NewAccountHandler(accountRepo, emailCollector)
	stakeholderHandler := handlers.NewStakeholderHandler(stakeholderRepo)

	api := router.Group("/api/v1")
	{
		api.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})
		
		api.GET("/emails/important", emailHandler.GetImportantEmails)
		api.GET("/emails/all", emailHandler.GetGlobalInbox)
		api.PUT("/emails/:emailId/status", emailHandler.UpdateEmailStatus)
		api.POST("/emails/:emailId/summary", emailHandler.GenerateSummary)
		api.GET("/accounts/:accountId/emails", emailHandler.GetEmailsByAccount)
		api.POST("/accounts", accountHandler.CreateAccount)
		api.GET("/accounts", accountHandler.GetAccounts)
		api.PUT("/accounts/:accountId", accountHandler.UpdateAccount)
		api.DELETE("/accounts/:accountId", accountHandler.DeleteAccount)
		api.POST("/accounts/test", accountHandler.TestConnection)
		api.POST("/accounts/:accountId/test", accountHandler.TestExistingConnection)
		api.POST("/accounts/:accountId/sync", accountHandler.SyncAccount)
		
		api.POST("/stakeholders", stakeholderHandler.CreateStakeholder)
		api.GET("/stakeholders", stakeholderHandler.GetStakeholders)
		api.DELETE("/stakeholders/:id", stakeholderHandler.DeleteStakeholder)
	}

	srv := &http.Server{
		Addr:    ":10000",
		Handler: router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	log.Println("Server listening on :10000")

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

func runMigrations(db *sql.DB) {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Fatalf("Failed to create migration driver: %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"postgres", driver)
	if err != nil {
		log.Fatalf("Failed to create migration instance: %v", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("Failed to run migrations: %v", err)
	}
	log.Println("Migrations ran successfully")
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
