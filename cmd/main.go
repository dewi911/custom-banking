package main

import (
	"custom-banking/internal/repository"
	"custom-banking/internal/service"
	"custom-banking/internal/transport/rest"
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

	userRepo := repository.NewUsers(db)

	userService := service.NewUsers(userRepo)

	authTransport := rest.NewAuth(userService)

	g := gin.New()

	authTransport.InjectRouters(g)

	fmt.Println("Server run...")
	if err := g.Run(fmt.Sprintf(":%s", cfg.Port)); err != nil {
		logrus.Fatalf("error occured while running http server %s", err.Error())
	}
}
