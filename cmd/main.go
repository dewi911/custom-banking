package main

import (
	"custom-banking/internal/repository"
	"custom-banking/internal/service"
	"custom-banking/internal/transport/rest"
	"custom-banking/pkg"
	"custom-banking/pkg/config"
	"custom-banking/pkg/database"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func main() {
	cfg, err := config.Parse()
	if err != nil {
		logrus.WithError(err).Fatalf("error parsing config from env variables: %s", err.Error())
	}

	fmt.Printf("%+v\n", cfg)

	db, err := database.CreateConnection(cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPass, cfg.DBName, cfg.SSLMode)
	if err != nil {
		logrus.WithError(err).Fatalf("error connecting to database: %s", err.Error())
	}
	defer db.Close()

	randomGenerator := pkg.NewGenerator("BY", "123456")

	userRepo := repository.NewUsers(db)
	tokensRepository := repository.NewTokens(db)
	rolesRepository := repository.NewRoles(db)
	accountRepository := repository.NewAccount(db)
	transactionRepository := repository.NewTransactions(db)
	cardRepository := repository.NewCard(db)
	eventRepository := repository.NewEvent(db)

	accessControl := service.NewAccessControl(rolesRepository)

	usersService := service.NewUsers(userRepo, tokensRepository, rolesRepository, eventRepository)
	accountService := service.NewAccount(accountRepository, transactionRepository, eventRepository, randomGenerator)
	transactionService := service.NewTransaction(transactionRepository, accountRepository)
	cardService := service.NewCard(cardRepository, userRepo, accountRepository, eventRepository, randomGenerator)
	eventService := service.NewEvent(eventRepository)

	authTransport := rest.NewAuth(usersService)
	accountTransport := rest.NewAccount(accountService)
	transactionTransport := rest.NewTransaction(transactionService)
	cardTransport := rest.NewCard(cardService)
	eventTransport := rest.NewEvent(eventService)

	accessControlMiddleware := rest.AccessControlMiddleware(accessControl, rolesRepository)

	g := gin.New()

	g.Use(rest.LoggingMiddleware())
	authTransport.InjectRouters(g, accessControlMiddleware)
	accountTransport.InjectRoutes(g, authTransport.AuthMiddleware(), accessControlMiddleware)
	transactionTransport.InjectRoutes(g, authTransport.AuthMiddleware(), accessControlMiddleware)
	cardTransport.InjectRoutes(g, authTransport.AuthMiddleware(), accessControlMiddleware)
	eventTransport.InjectRoutes(g, authTransport.AuthMiddleware(), accessControlMiddleware)

	fmt.Println("Server run...")
	if err := g.Run(fmt.Sprintf(":%s", cfg.Port)); err != nil {
		logrus.Fatalf("error occured while running http server %s", err.Error())
	}
}
