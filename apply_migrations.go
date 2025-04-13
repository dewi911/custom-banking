package main

import (
	"custom-banking/pkg/database/migrations"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"os"
)

func main() {
	logrus.SetLevel(logrus.InfoLevel)
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	logrus.Info("Starting migration script")

	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		connStr = "postgres://postgres:postgres@localhost:5432/banking?sslmode=disable"
		logrus.Info("Using default database connection string")
	}

	logrus.Info("Connecting to database...")
	db, err := sqlx.Connect("postgres", connStr)
	if err != nil {
		logrus.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()
	logrus.Info("Successfully connected to database")

	logrus.Info("Applying migrations...")
	err = migrations.RunMigrations(db)
	if err != nil {
		logrus.Fatalf("Failed to apply migrations: %v", err)
	}

	var tableExists bool
	err = db.QueryRow("SELECT EXISTS (SELECT FROM pg_tables WHERE schemaname = 'public' AND tablename = 'loan_payments')").Scan(&tableExists)
	if err != nil {
		logrus.Fatalf("Failed to check if loan_payments table exists: %v", err)
	}

	if tableExists {
		logrus.Info("loan_payments table exists!")
	} else {
		logrus.Error("loan_payments table does not exist!")
		logrus.Info("Creating loan_payments table directly...")
		_, err = db.Exec(`
			CREATE TABLE IF NOT EXISTS loan_payments (
				id serial PRIMARY KEY,
				loan_id integer REFERENCES loans(id),
				amount numeric,
				date timestamp,
				status varchar(20),
				payment_method varchar(50),
				transaction_id integer
			);
			CREATE INDEX IF NOT EXISTS idx_loan_payments_loan_id ON loan_payments (loan_id);
		`)
		if err != nil {
			logrus.Fatalf("Failed to create loan_payments table: %v", err)
		}
		logrus.Info("loan_payments table created successfully")
	}

	logrus.Info("Migration process completed successfully")
}
