package database

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"ariga.io/atlas-go-sdk/atlasexec"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func OpenConnection(runMigrations bool) (*gorm.DB, error) {
	connectionUrl := os.Getenv("DATABASE_URL")
	if connectionUrl == "" {
		return nil, errors.New("environment variable 'DATABASE_URL' has not been set")
	}

	tlsParameters := os.Getenv("DATABASE_TLS_PARAMETERS")

	connectionUrl = connectionUrl + tlsParameters

	db, err := gorm.Open(postgres.Open(connectionUrl), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})

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

	// Run Atlas migrations
	if runMigrations {
		workingDir := "."
		client, err := atlasexec.NewClient(workingDir, "atlas")
		if err != nil {
			return nil, fmt.Errorf("failed to initialize atlas client: %s", err.Error())
		}

		_, err = client.MigrateApply(context.Background(), &atlasexec.MigrateApplyParams{
			Env:       "gorm",
			ConfigURL: "file://atlas/atlas.hcl",
			URL:       connectionUrl,
			DirURL:    "file://atlas/migrations",
		})
		if err != nil {
			return nil, fmt.Errorf("failed to apply atlas migrations: %s", err.Error())
		}
	}

	return db, nil
}
