package database

import (
	"errors"
	"fmt"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func OpenConnection() (*gorm.DB, error) {
	connectionUrl := os.Getenv("DATABASE_URL")
	if connectionUrl == "" {
		return nil, errors.New("environment variable 'DATABASE_URL' has not been set")
	}

	db, err := gorm.Open(postgres.Open(connectionUrl))
	if err != nil {
		return nil, fmt.Errorf("failed to open connection to database: %s", err.Error())
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize sqlDB: %s", err.Error())
	}

	// Configure the database pool
	// SetMaxIdleConns sets the maximum number of connections in the idle connection pool.
	sqlDB.SetMaxIdleConns(10)

	// SetMaxOpenConns sets the maximum number of open connections to the database.
	sqlDB.SetMaxOpenConns(20)

	// SetConnMaxLifetime sets the maximum amount of time a connection may be reused.
	sqlDB.SetConnMaxLifetime(time.Hour)

	err = sqlDB.Ping()
	if err != nil {
		return nil, fmt.Errorf("failed to ping the database: %s", err.Error())
	}

	return db, nil
}
