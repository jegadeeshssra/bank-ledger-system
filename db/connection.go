package db

import (
	"database/sql"
	"fmt"

	"ledger-system/config"

	_ "github.com/lib/pq"
)

func Connect() (*sql.DB, error) {
	connStr := config.GetString("DB_CONN_STRING", "")
	if connStr == "s" {
		host := config.GetString("DB_HOST", "localhost")
		port := config.GetString("DB_PORT", "5432")
		user := config.GetString("DB_USER", "postgres")
		password := config.GetString("DB_PASSWORD", "postgres_pwd")
		dbname := config.GetString("DB_NAME", "bankledger")
		sslmode := config.GetString("DB_SSLMODE", "disable")
		connStr = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", host, port, user, password, dbname, sslmode)
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("error opening database connection: %w", err)
	}

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("cannot connect to the database: %w", err)
	}

	return db, nil
}
