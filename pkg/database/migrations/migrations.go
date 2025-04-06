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

//go:embed schema.sql constraints.sql
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

	logrus.Info("Checking for blocking sessions...")
	var blockingCount int
	err = db.QueryRow(`
		SELECT COUNT(*) FROM pg_stat_activity 
		WHERE waiting AND NOT pid = pg_backend_pid()
	`).Scan(&blockingCount)
	if err != nil {
		logrus.Warnf("Unable to check for blocking sessions: %v", err)
	} else if blockingCount > 0 {
		logrus.Warnf("Detected %d blocking sessions that might affect migrations", blockingCount)
	}

	if err := applyMigrationFile(db, "schema.sql"); err != nil {
		return err
	}

	if err := applyMigrationFile(db, "constraints.sql"); err != nil {
		return err
	}

	err = verifyTables(db)
	if err != nil {
		return fmt.Errorf("database table verification failed: %w", err)
	}

	logrus.Info("Database migrations applied successfully")
	return nil
}

func applyMigrationFile(db *sqlx.DB, filename string) error {
	logrus.Infof("Applying migration file: %s", filename)

	sqlContent, err := migrationsFS.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("error reading %s file: %w", filename, err)
	}

	commands := splitSQLScript(string(sqlContent))

	var migrationExists bool
	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM migrations WHERE name = $1)", filename).Scan(&migrationExists)
	if err != nil {
		logrus.Warnf("Error checking migration status: %v", err)
	} else if migrationExists {
		logrus.Infof("Migration %s has already been applied, checking integrity", filename)
	}

	var tx *sqlx.Tx
	if filename == "schema.sql" {
		tx, err = db.Beginx()
		if err != nil {
			return fmt.Errorf("error starting transaction for %s: %w", filename, err)
		}
		defer func() {
			if err != nil {
				tx.Rollback()
				logrus.Warnf("Transaction rolled back for %s", filename)
			}
		}()
	}

	for i, cmd := range commands {
		if strings.TrimSpace(cmd) == "" {
			continue
		}

		logrus.Debugf("Executing SQL command %d/%d from %s: %s", i+1, len(commands), filename, getShortCommandForLog(cmd))

		var execErr error
		if tx != nil {
			_, execErr = tx.Exec(cmd)
		} else {
			_, execErr = db.Exec(cmd)
		}

		if execErr != nil {
			if filename == "constraints.sql" && strings.Contains(execErr.Error(), "уже существует") {
				logrus.Warnf("Constraint already exists, continuing: %v", execErr)
				continue
			}

			logrus.Errorf("Error executing command: %s", cmd)
			return fmt.Errorf("error applying migration command from %s: %w\nCommand: %s",
				filename, execErr, getShortCommandForLog(cmd))
		}

		if i < len(commands)-1 && strings.Contains(cmd, "DO $$") {
			logrus.Debug("Completed executing DO block, proceeding to next command")
		}
	}

	if tx != nil {
		if err = tx.Commit(); err != nil {
			return fmt.Errorf("error committing transaction for %s: %w", filename, err)
		}
		logrus.Debug("Transaction committed successfully")
	}

	if !migrationExists {
		_, err = db.Exec("INSERT INTO migrations (name) VALUES ($1) ON CONFLICT DO NOTHING", filename)
		if err != nil {
			logrus.Warnf("Error recording migration status: %v", err)
		}
	}

	logrus.Infof("Successfully applied migration file: %s", filename)
	return nil
}

func splitSQLScript(script string) []string {
	var commands []string
	var currentCommand strings.Builder
	dollarQuoteDepth := 0

	lines := strings.Split(script, "\n")
	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		if trimmedLine == "" || strings.HasPrefix(trimmedLine, "--") {
			continue
		}

		dollarCount := strings.Count(trimmedLine, "$$")
		if dollarCount > 0 {
			dollarQuoteDepth += dollarCount % 2
		}

		currentCommand.WriteString(line)
		currentCommand.WriteString("\n")

		if (strings.HasSuffix(trimmedLine, ";") && dollarQuoteDepth == 0) ||
			(strings.Contains(trimmedLine, "END") && strings.Contains(trimmedLine, "$$") && dollarQuoteDepth == 0) {
			commands = append(commands, currentCommand.String())
			currentCommand.Reset()
		}
	}

	if currentCommand.Len() > 0 {
		commands = append(commands, currentCommand.String())
	}

	var refinedCommands []string
	for _, cmd := range commands {
		if strings.Count(cmd, "DO $$") > 1 {
			parts := splitDOBlocks(cmd)
			refinedCommands = append(refinedCommands, parts...)
		} else {
			refinedCommands = append(refinedCommands, cmd)
		}
	}

	return refinedCommands
}

func splitDOBlocks(script string) []string {
	var blocks []string
	var currentBlock strings.Builder
	dollarQuoteDepth := 0

	lines := strings.Split(script, "\n")
	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		if strings.Contains(trimmedLine, "DO $$") && dollarQuoteDepth == 0 && currentBlock.Len() > 0 {
			blocks = append(blocks, currentBlock.String())
			currentBlock.Reset()
		}

		dollarCount := strings.Count(trimmedLine, "$$")
		if dollarCount > 0 {
			dollarQuoteDepth += dollarCount % 2
		}

		currentBlock.WriteString(line)
		currentBlock.WriteString("\n")

		if strings.Contains(trimmedLine, "END") && strings.Contains(trimmedLine, "$$") && dollarQuoteDepth == 0 {
			blocks = append(blocks, currentBlock.String())
			currentBlock.Reset()
		}
	}

	if currentBlock.Len() > 0 {
		blocks = append(blocks, currentBlock.String())
	}

	return blocks
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

	logrus.Info("Verifying essential foreign key constraints...")
	essentialConstraints := []struct {
		name  string
		table string
	}{
		{"FK_accounts_currency_1", "accounts"},
		{"FK_refresh_tokens_user_1", "refresh_tokens"},
		{"FK_users_role_1", "users"},
	}

	for _, constraint := range essentialConstraints {
		var exists bool
		query := `
			SELECT EXISTS (
				SELECT 1 FROM information_schema.table_constraints 
				WHERE constraint_name = $1 AND table_name = $2
			)
		`
		err := db.QueryRow(query, constraint.name, constraint.table).Scan(&exists)
		if err != nil {
			return fmt.Errorf("error checking constraint %s: %w", constraint.name, err)
		}

		if !exists {
			logrus.Warnf("Foreign key constraint %s on table %s is missing", constraint.name, constraint.table)
		}
	}

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
