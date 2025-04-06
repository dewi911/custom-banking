package migrations

import (
	"embed"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"io/fs"
	"path"
	"strings"
)

//go:embed schema.sql
var migrationsFS embed.FS

func RunMigrations(db *sqlx.DB) error {
	if err := db.Ping(); err != nil {
		return fmt.Errorf("unable to ping database: %w", err)
	}

	var pgVersion string
	if err := db.QueryRow("SELECT version()").Scan(&pgVersion); err != nil {
		return fmt.Errorf("error getting PostgreSQL version: %w", err)
	}
	logrus.Infof("Connected to PostgreSQL: %s", pgVersion)

	_, err := db.Exec("CREATE SCHEMA IF NOT EXISTS public")
	if err != nil {
		return fmt.Errorf("error creating public schema: %w", err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS migrations (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL UNIQUE,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("error creating migrations table: %w", err)
	}

	schemaSQL, err := migrationsFS.ReadFile("schema.sql")
	if err != nil {
		return fmt.Errorf("error reading schema file: %w", err)
	}

	logrus.Info("Applying database schema migrations...")

	commands := splitSQLScript(string(schemaSQL))
	for i, cmd := range commands {
		if strings.TrimSpace(cmd) == "" {
			continue
		}

		logrus.Debugf("Executing SQL command %d/%d: %s", i+1, len(commands), getShortCommandForLog(cmd))
		_, err = db.Exec(cmd)
		if err != nil {
			return fmt.Errorf("error applying migration command: %w\nCommand: %s", err, getShortCommandForLog(cmd))
		}
	}

	err = verifyTables(db)
	if err != nil {
		return fmt.Errorf("database table verification failed: %w", err)
	}

	logrus.Info("Database migrations applied successfully")
	return nil
}

func splitSQLScript(script string) []string {
	var commands []string
	var currentCommand strings.Builder
	inFunction := false
	inBlock := false
	dollarQuote := false

	lines := strings.Split(script, "\n")
	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		if trimmedLine == "" || strings.HasPrefix(trimmedLine, "--") {
			continue
		}

		if strings.Contains(trimmedLine, "DO $$") || strings.Contains(trimmedLine, "BEGIN") {
			inFunction = true
		}

		if strings.Contains(trimmedLine, "$$") {
			dollarQuote = !dollarQuote
		}

		if strings.Contains(trimmedLine, "BEGIN") {
			inBlock = true
		}
		if strings.Contains(trimmedLine, "END") {
			inBlock = false
		}

		currentCommand.WriteString(line)
		currentCommand.WriteString("\n")

		isEndOfCommand := strings.HasSuffix(trimmedLine, ";") && !inFunction && !inBlock && !dollarQuote
		isEndOfPlPgSQL := strings.Contains(trimmedLine, "END") && strings.Contains(trimmedLine, "$$") && inFunction

		if isEndOfCommand || isEndOfPlPgSQL {
			commands = append(commands, currentCommand.String())
			currentCommand.Reset()

			if isEndOfPlPgSQL {
				inFunction = false
			}
		}
	}

	if currentCommand.Len() > 0 {
		commands = append(commands, currentCommand.String())
	}

	return commands
}

func getShortCommandForLog(cmd string) string {
	const maxLength = 100
	cmd = strings.ReplaceAll(cmd, "\n", " ")
	cmd = strings.TrimSpace(cmd)

	if len(cmd) > maxLength {
		return cmd[:maxLength] + "..."
	}
	return cmd
}

func verifyTables(db *sqlx.DB) error {
	essentialTables := []string{
		"roles", "users", "accounts", "currency", "transactions",
		"refresh_tokens", "cards", "event",
	}

	logrus.Info("Verifying essential database tables...")
	for _, table := range essentialTables {
		var exists bool
		query := `
			SELECT EXISTS (
				SELECT FROM information_schema.tables 
				WHERE table_schema = 'public' 
				AND table_name = $1
			)
		`
		err := db.QueryRow(query, table).Scan(&exists)
		if err != nil {
			return fmt.Errorf("error checking table %s: %w", table, err)
		}

		if !exists {
			return fmt.Errorf("essential table %s is missing after migrations", table)
		}
	}
	logrus.Info("All essential tables are present")

	essentialTypes := []string{
		"transaction_status", "event_type", "card_type",
	}

	logrus.Info("Verifying essential ENUM types...")
	for _, typeName := range essentialTypes {
		var exists bool
		query := `
			SELECT EXISTS (
				SELECT 1 FROM pg_type WHERE typname = $1
			)
		`
		err := db.QueryRow(query, typeName).Scan(&exists)
		if err != nil {
			return fmt.Errorf("error checking type %s: %w", typeName, err)
		}

		if !exists {
			return fmt.Errorf("essential ENUM type %s is missing after migrations", typeName)
		}
	}
	logrus.Info("All essential ENUM types are present")

	var roleCount int
	err := db.QueryRow("SELECT COUNT(*) FROM roles").Scan(&roleCount)
	if err != nil {
		return fmt.Errorf("error checking roles data: %w", err)
	}
	if roleCount < 3 {
		logrus.Warn("Default roles may not be fully populated")
	} else {
		logrus.Info("Default roles are properly populated")
	}

	var currencyCount int
	err = db.QueryRow("SELECT COUNT(*) FROM currency").Scan(&currencyCount)
	if err != nil {
		return fmt.Errorf("error checking currency data: %w", err)
	}
	if currencyCount < 5 {
		logrus.Warn("Default currencies may not be fully populated")
	} else {
		logrus.Info("Default currencies are properly populated")
	}

	return nil
}

func ListMigrationFiles() ([]string, error) {
	var files []string
	err := fs.WalkDir(migrationsFS, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && path.Ext(p) == ".sql" {
			files = append(files, p)
		}
		return nil
	})

	return files, err
}
