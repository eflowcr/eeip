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
	"github.com/eprac/eeip-backend/internal/infrastructure/auth"
	"github.com/eprac/eeip-backend/internal/infrastructure/database"
	"github.com/eprac/eeip-backend/internal/interfaces/api/handlers"
	"github.com/eprac/eeip-backend/internal/interfaces/api/middleware"

	"github.com/emersion/go-message"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"golang.org/x/net/html/charset"
)

func init() {
	message.CharsetReader = charset.NewReaderLabel
}

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
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	router.Use(cors.New(config))
	
	// Repositories
	sqlxDB := sqlx.NewDb(db, "postgres")
	emailRepo := database.NewEmailRepository(sqlxDB)
	accountRepo := database.NewAccountRepository(sqlxDB)
	stakeholderRepo := database.NewStakeholderRepository(sqlxDB)
	userRepo := database.NewUserRepository(sqlxDB)

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
	mailerSvc := services.NewMailerService()
	emailCollector := services.NewEmailCollector(emailRepo, aiEngine, stakeholderRepo, telegramSvc, mailerSvc)
	summaryEngine := services.NewSummaryEngine(openAIKey)
	
	cronService := services.NewCronService(emailRepo, accountRepo, stakeholderRepo, telegramSvc, mailerSvc)
	cronService.Start()

	// Handlers
	emailHandler := handlers.NewEmailHandler(emailRepo, summaryEngine)
	accountHandler := handlers.NewAccountHandler(accountRepo, emailCollector)
	stakeholderHandler := handlers.NewStakeholderHandler(stakeholderRepo)
	jwtSecret := getEnv("JWT_SECRET", "super-secret-key")
	tokenManager := auth.NewJWTManager(jwtSecret, 24*time.Hour)
	authHandler := handlers.NewAuthHandler(userRepo, accountRepo, tokenManager, emailCollector)
	userHandler := handlers.NewUserHandler(userRepo)

	api := router.Group("/api/v1")
	{
		api.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})
		
		api.POST("/auth/login", authHandler.Login)
		api.POST("/auth/register", authHandler.Register)

		protected := api.Group("/")
		protected.Use(middleware.AuthMiddleware(tokenManager))
		{
			// Everyone logged in can access these, but handlers will filter by Role
			protected.GET("/emails/important", emailHandler.GetImportantEmails)
			protected.GET("/emails/all", emailHandler.GetGlobalInbox)
			protected.PUT("/emails/:emailId/status", emailHandler.UpdateEmailStatus)
			protected.POST("/emails/:emailId/summary", emailHandler.GenerateSummary)
			protected.GET("/accounts/:accountId/emails", emailHandler.GetEmailsByAccount)
			protected.GET("/accounts", accountHandler.GetAccounts)

			// Only Admin / Auditor can access (Auditor might be read-only, but for now we keep it simple)
			adminOrAuditor := protected.Group("/")
			adminOrAuditor.Use(middleware.RoleMiddleware("Admin", "Auditor"))
			{
				adminOrAuditor.GET("/stakeholders", stakeholderHandler.GetStakeholders)
			}

			// Only Admin can modify accounts and stakeholders
			adminOnly := protected.Group("/")
			adminOnly.Use(middleware.RoleMiddleware("Admin"))
			{
				adminOnly.POST("/accounts", accountHandler.CreateAccount)
				adminOnly.PUT("/accounts/:accountId", accountHandler.UpdateAccount)
				adminOnly.DELETE("/accounts/:accountId", accountHandler.DeleteAccount)
				adminOnly.POST("/accounts/test", accountHandler.TestConnection)
				adminOnly.POST("/accounts/:accountId/test", accountHandler.TestExistingConnection)
				adminOnly.POST("/accounts/:accountId/sync", accountHandler.SyncAccount)
				
				adminOnly.POST("/stakeholders", stakeholderHandler.CreateStakeholder)
				adminOnly.DELETE("/stakeholders/:id", stakeholderHandler.DeleteStakeholder)
				
				// Admin only user management
				adminOnly.GET("/users", userHandler.GetUsers)
				adminOnly.PUT("/users/:userId/role", userHandler.UpdateRole)
			}
		}
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

	cronService.Stop()
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
