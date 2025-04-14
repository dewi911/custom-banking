package main

import (
	"context"
	"custom-banking/internal/repository"
	"custom-banking/internal/service"
	"custom-banking/internal/transport/rest"
	"custom-banking/pkg"
	"custom-banking/pkg/config"
	"custom-banking/pkg/database"
	"custom-banking/pkg/database/migrations"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func init() {
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	if os.Getenv("DEBUG") == "true" {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-signalChan
		logrus.Infof("Received shutdown signal: %v", sig)
		cancel()
	}()

	cfg, err := config.Parse()
	if err != nil {
		logrus.WithError(err).Fatalf("error parsing config from env variables: %s", err.Error())
	}

	logrus.Infof("Starting application with configuration: %+v", cfg)

	logrus.Info("Connecting to database...")
	db, err := database.CreateConnection(cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPass, cfg.DBName, cfg.SSLMode)
	if err != nil {
		logrus.WithError(err).Fatalf("error connecting to database: %s", err.Error())
	}

	defer func() {
		logrus.Info("Closing database connection...")
		if err := db.Close(); err != nil {
			logrus.WithError(err).Error("Error closing database connection")
		} else {
			logrus.Info("Database connection closed successfully")
		}
	}()

	logrus.Info("Database connection established successfully")

	logrus.Info("Listing available migration files...")
	migrationFiles, err := migrations.ListMigrationFiles()
	if err != nil {
		logrus.WithError(err).Warn("Failed to list migration files")
	} else {
		for _, file := range migrationFiles {
			logrus.Infof("Found migration file: %s", file)
		}
	}

	logrus.Info("Starting database migrations...")
	if err := migrations.RunMigrations(db); err != nil {
		logrus.WithError(err).Fatalf("error running database migrations: %s", err.Error())
	}
	logrus.Info("Database migrations completed successfully")

	randomGenerator := pkg.NewGenerator("BY", "123456")

	logrus.Info("Initializing repositories...")
	userRepo := repository.NewUsers(db)
	tokensRepository := repository.NewTokens(db)
	rolesRepository := repository.NewRoles(db)
	accountRepository := repository.NewAccount(db)
	transactionRepository := repository.NewTransactions(db)
	cardRepository := repository.NewCard(db)
	eventRepository := repository.NewEvent(db)

	loansRepository := repository.NewLoans(db)
	stakingRepository := repository.NewStaking(db)
	cardTransfersRepository := repository.NewCardTransfers(db)
	insuranceRepository := repository.NewInsurance(db)
	cashbackRepository := repository.NewCashbackRepo(db)
	invoiceRepository := repository.NewInvoiceRepo(db)

	logrus.Info("Initializing services...")
	accessControl := service.NewAccessControl(rolesRepository)
	usersService := service.NewUsers(userRepo, tokensRepository, rolesRepository, eventRepository)
	accountService := service.NewAccount(accountRepository, transactionRepository, eventRepository, randomGenerator)
	transactionService := service.NewTransaction(transactionRepository, accountRepository)
	cardService := service.NewCard(cardRepository, userRepo, accountRepository, eventRepository, randomGenerator)
	eventService := service.NewEvent(eventRepository)

	loanService := service.NewLoanService(loansRepository, accountRepository)
	stakingService := service.NewStakingService(stakingRepository, accountRepository)
	cardTransfersService := service.NewCardTransfersService(cardTransfersRepository)
	insuranceService := service.NewInsuranceService(insuranceRepository, accountRepository, transactionRepository)
	cashbackService := service.NewCashbackService(cashbackRepository, accountRepository, transactionRepository)
	invoiceService := service.NewInvoiceService(invoiceRepository, accountRepository, cardRepository, userRepo)

	logrus.Info("Initializing transport layer...")
	authTransport := rest.NewAuth(usersService)
	accountTransport := rest.NewAccount(accountService)
	transactionTransport := rest.NewTransaction(transactionService)
	cardTransport := rest.NewCard(cardService)
	eventTransport := rest.NewEvent(eventService)

	loanHandler := rest.NewLoanHandler(loanService)
	stakingHandler := rest.NewStakingHandler(stakingService)
	cardTransfersHandler := rest.NewCardTransfersHandler(cardTransfersService)
	insuranceHandler := rest.NewInsuranceHandler(insuranceService)
	cashbackHandler := rest.NewCashbackHandler(cashbackService)
	invoiceHandler := rest.NewInvoiceHandler(invoiceService)

	accessControlMiddleware := rest.AccessControlMiddleware(accessControl, rolesRepository)

	logrus.Info("Configuring HTTP server...")
	g := gin.New()
	g.Use(rest.LoggingMiddleware())

	logrus.Info("Registering API routes...")
	authTransport.InjectRouters(g, accessControlMiddleware)
	accountTransport.InjectRoutes(g, authTransport.AuthMiddleware(), accessControlMiddleware)
	transactionTransport.InjectRoutes(g, authTransport.AuthMiddleware(), accessControlMiddleware)
	cardTransport.InjectRoutes(g, authTransport.AuthMiddleware(), accessControlMiddleware)
	eventTransport.InjectRoutes(g, authTransport.AuthMiddleware(), accessControlMiddleware)
	loanHandler.Register(g, authTransport.AuthMiddleware(), accessControlMiddleware)
	stakingHandler.Register(g, authTransport.AuthMiddleware(), accessControlMiddleware)
	cardTransfersHandler.Register(g, authTransport.AuthMiddleware(), accessControlMiddleware)
	v1 := g.Group("/api/v1")
	v1.Use(authTransport.AuthMiddleware())

	logrus.Info("Registering new API routes for loans, staking, and card transfers...")
	//stakingHandler.Register(v1)
	//cardTransfersHandler.Register(v1)
	insuranceHandler.InitRoutes(v1)
	cashbackHandler.InitRoutes(v1)
	invoiceHandler.InitRoutes(v1)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.Port),
		Handler: g,
	}

	go func() {
		logrus.Infof("Starting server on port %s...", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.Fatalf("Error starting server: %v", err)
		}
	}()

	<-ctx.Done()
	logrus.Info("Shutdown signal received, initiating graceful shutdown...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logrus.WithError(err).Error("Server shutdown error")
	} else {
		logrus.Info("Server gracefully stopped")
	}

	logrus.Info("Application shutdown complete")
}
