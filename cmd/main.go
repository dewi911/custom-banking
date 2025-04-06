package main

import (
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
	"os"
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

	logrus.Info("Initializing services...")
	accessControl := service.NewAccessControl(rolesRepository)
	usersService := service.NewUsers(userRepo, tokensRepository, rolesRepository, eventRepository)
	accountService := service.NewAccount(accountRepository, transactionRepository, eventRepository, randomGenerator)
	transactionService := service.NewTransaction(transactionRepository, accountRepository)
	cardService := service.NewCard(cardRepository, userRepo, accountRepository, eventRepository, randomGenerator)
	eventService := service.NewEvent(eventRepository)

	logrus.Info("Initializing transport layer...")
	authTransport := rest.NewAuth(usersService)
	accountTransport := rest.NewAccount(accountService)
	transactionTransport := rest.NewTransaction(transactionService)
	cardTransport := rest.NewCard(cardService)
	eventTransport := rest.NewEvent(eventService)

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

	logrus.Infof("Starting server on port %s...", cfg.Port)
	if err := g.Run(fmt.Sprintf(":%s", cfg.Port)); err != nil {
		logrus.Fatalf("error occurred while running http server: %s", err.Error())
	}
}
