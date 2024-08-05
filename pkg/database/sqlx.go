package database

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func CreateConnection(host, port, username, password, dbname string, secure bool) (*sqlx.DB, error) {
	sslmode := "disable"
	if secure {
		sslmode = "enable"
	}

	dbpromt := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", host, port, username, password, dbname, sslmode)

	db, err := sqlx.Connect("postgres", dbpromt)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
