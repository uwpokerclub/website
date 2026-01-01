// Package testutils provides utilities for testing with containerized PostgreSQL databases.
// It wraps testcontainers-go to provide easy-to-use PostgreSQL 17 containers with
// automatic Atlas migration support for isolated testing environments.
//
// The package automatically looks for Atlas configuration in the current working directory
// under atlas/atlas.hcl and applies migrations from atlas/migrations/.
//
// Example usage:
//
//	func TestMyFeature(t *testing.T) {
//		ctx := context.Background()
//		container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
//		require.NoError(t, err)
//		defer container.Close(ctx)
//
//		db := container.GetDB()
//		// Your test logic here
//	}
//
// For parallel testing with multiple containers:
//
//	func TestParallelFeatures(t *testing.T) {
//		tests := []struct {
//			name string
//			test func(*testing.T, *testutils.PostgresTestContainer)
//		}{
//			{"test_users", func(t *testing.T, container *testutils.PostgresTestContainer) {
//				// Test user functionality
//			}},
//			{"test_events", func(t *testing.T, container *testutils.PostgresTestContainer) {
//				// Test event functionality
//			}},
//		}
//
//		for _, tt := range tests {
//			tt := tt
//			t.Run(tt.name, func(t *testing.T) {
//				t.Parallel()
//				ctx := context.Background()
//				container, err := testutils.NewPostgresContainer(ctx, testutils.PostgresConfig{})
//				require.NoError(t, err)
//				defer container.Close(ctx)
//
//				// Reset database for clean state
//				require.NoError(t, container.ResetDatabase(ctx))
//				tt.test(t, container)
//			})
//		}
//	}
package testutils

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"ariga.io/atlas-go-sdk/atlasexec"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	postgresdriver "gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// PostgresTestContainer wraps a PostgreSQL test container with methods for
// database management during testing. Each container provides an isolated
// PostgreSQL 17 database with Atlas migrations automatically applied.
//
// The container supports:
//   - Automatic Atlas migration execution on startup
//   - Database reset for clean test states
//   - Proper resource cleanup
//   - Connection pooling configuration
//
// Example:
//
//	container, err := NewPostgresContainer(ctx, PostgresConfig{})
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer container.Close(ctx)
//
//	db := container.GetDB()
//	// Use db for testing
type PostgresTestContainer struct {
	container     testcontainers.Container
	connectionURL string
	db            *gorm.DB
	sqlDB         *sql.DB
	dbName        string
}

// PostgresConfig holds configuration for the PostgreSQL test container.
//
// All fields are optional and will use sensible defaults if not provided:
//   - DatabaseName: defaults to "testdb"
//   - Username: defaults to "testuser"
//   - Password: defaults to "testpass"
//
// Example:
//
//	config := PostgresConfig{
//		DatabaseName: "my_test_db",
//		Username:     "test_user",
//		Password:     "secure_pass",
//	}
type PostgresConfig struct {
	DatabaseName string
	Username     string
	Password     string
}

// NewPostgresContainer creates and starts a new PostgreSQL 17 test container
// with the specified configuration. The container will be ready to accept connections
// and will have Atlas migrations automatically applied.
//
// Each container is isolated and can be used in parallel testing scenarios.
// Remember to call Close() on the returned container to clean up resources.
//
// Example:
//
//	ctx := context.Background()
//	container, err := NewPostgresContainer(ctx, PostgresConfig{
//		DatabaseName: "integration_test",
//	})
//	if err != nil {
//		return fmt.Errorf("failed to create test container: %w", err)
//	}
//	defer container.Close(ctx)
//
//	db := container.GetDB()
//	// Use db for testing...
//
// For tests that need a fresh database state:
//
//	// Reset database between test cases
//	err = container.ResetDatabase(ctx)
//	if err != nil {
//		return fmt.Errorf("failed to reset database: %w", err)
//	}
func NewPostgresContainer(ctx context.Context, config PostgresConfig) (*PostgresTestContainer, error) {
	// Set defaults
	if config.DatabaseName == "" {
		config.DatabaseName = "testdb"
	}
	if config.Username == "" {
		config.Username = "testuser"
	}
	if config.Password == "" {
		config.Password = "testpass"
	}

	// Create PostgreSQL container
	postgresContainer, err := postgres.Run(ctx,
		"postgres:17",
		postgres.WithDatabase(config.DatabaseName),
		postgres.WithUsername(config.Username),
		postgres.WithPassword(config.Password),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to start postgres container: %w", err)
	}

	// Get connection URL
	connectionURL, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		postgresContainer.Terminate(ctx)
		return nil, fmt.Errorf("failed to get connection string: %w", err)
	}

	// Create GORM DB connection
	db, err := gorm.Open(postgresdriver.Open(connectionURL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		postgresContainer.Terminate(ctx)
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Get underlying SQL DB for atlas migrations
	sqlDB, err := db.DB()
	if err != nil {
		postgresContainer.Terminate(ctx)
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Configure connection pool for testing
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Test connection
	if err := sqlDB.Ping(); err != nil {
		postgresContainer.Terminate(ctx)
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	container := &PostgresTestContainer{
		container:     postgresContainer,
		connectionURL: connectionURL,
		db:            db,
		sqlDB:         sqlDB,
		dbName:        config.DatabaseName,
	}

	// Run atlas migrations
	if err := container.runAtlasMigrations(ctx); err != nil {
		container.Close(ctx)
		return nil, fmt.Errorf("failed to run atlas migrations: %w", err)
	}

	return container, nil
}

// GetDB returns the GORM database connection for this container.
// The connection is ready to use and has been configured with appropriate
// connection pooling settings for testing.
//
// Example:
//
//	db := container.GetDB()
//	var count int64
//	err := db.Table("users").Count(&count).Error
func (p *PostgresTestContainer) GetDB() *gorm.DB {
	return p.db
}

// GetConnectionURL returns the database connection URL for this container.
// This can be useful for tools that need a direct connection string.
//
// Example:
//
//	url := container.GetConnectionURL()
//	// url is something like: "postgres://testuser:testpass@localhost:32768/testdb?sslmode=disable"
func (p *PostgresTestContainer) GetConnectionURL() string {
	return p.connectionURL
}

// GetSQLDB returns the underlying sql.DB connection.
// This provides access to the standard library database interface
// for cases where GORM is not needed.
//
// Example:
//
//	sqlDB := container.GetSQLDB()
//	err := sqlDB.Ping()
func (p *PostgresTestContainer) GetSQLDB() *sql.DB {
	return p.sqlDB
}

// ResetDatabase truncates all tables to provide a clean state for tests.
// This is more efficient than dropping and recreating the database or container.
//
// This ensures a completely clean database state while preserving the schema
// and is safe to call between test cases.
//
// Example:
//
//	// Between test cases
//	err := container.ResetDatabase(ctx)
//	if err != nil {
//		t.Fatalf("Failed to reset database: %v", err)
//	}
//	// Database is now in a clean state for the next test
func (p *PostgresTestContainer) ResetDatabase(ctx context.Context) error {
	// Get all table names from the public schema
	var tableNames []string
	err := p.db.Raw(`
		SELECT tablename 
		FROM pg_tables 
		WHERE schemaname = 'public'
	`).Scan(&tableNames).Error
	if err != nil {
		return fmt.Errorf("failed to get table names: %w", err)
	}

	if len(tableNames) == 0 {
		return nil // No tables to reset
	}

	// Disable foreign key checks temporarily
	if err := p.db.Exec("SET session_replication_role = replica").Error; err != nil {
		return fmt.Errorf("failed to disable foreign key checks: %w", err)
	}

	// Truncate all tables
	for _, tableName := range tableNames {
		if err := p.db.Exec(fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", tableName)).Error; err != nil {
			return fmt.Errorf("failed to truncate table %s: %w", tableName, err)
		}
	}

	// Re-enable foreign key checks
	if err := p.db.Exec("SET session_replication_role = DEFAULT").Error; err != nil {
		return fmt.Errorf("failed to re-enable foreign key checks: %w", err)
	}

	return nil
}

// Close terminates the container and cleans up resources.
// This should always be called when done with a container, typically
// in a defer statement right after container creation.
//
// Example:
//
//	container, err := NewPostgresContainer(ctx, config)
//	if err != nil {
//		return err
//	}
//	defer func() {
//		if err := container.Close(ctx); err != nil {
//			log.Printf("Failed to close container: %v", err)
//		}
//	}()
func (p *PostgresTestContainer) Close(ctx context.Context) error {
	if p.sqlDB != nil {
		if err := p.sqlDB.Close(); err != nil {
			return fmt.Errorf("failed to close database connection: %w", err)
		}
	}

	if p.container != nil {
		if err := p.container.Terminate(ctx); err != nil {
			return fmt.Errorf("failed to terminate container: %w", err)
		}
	}

	return nil
}

// runAtlasMigrations applies atlas migrations to the test database
func (p *PostgresTestContainer) runAtlasMigrations(ctx context.Context) error {
	// Find the server module root by looking for go.mod
	moduleRoot, err := findModuleRoot()
	if err != nil {
		return fmt.Errorf("failed to find module root: %w", err)
	}

	// Create Atlas client with module root directory
	client, err := atlasexec.NewClient(moduleRoot, "atlas")
	if err != nil {
		return fmt.Errorf("failed to initialize atlas client: %w", err)
	}

	// Construct absolute paths for atlas configuration and migrations
	configPath := filepath.Join(moduleRoot, "atlas", "atlas.hcl")
	migrationsPath := filepath.Join(moduleRoot, "atlas", "migrations")

	// Verify that the configuration file exists
	if _, err := os.Stat(configPath); err != nil {
		return fmt.Errorf("atlas config file not found at %s: %w", configPath, err)
	}

	// Apply pending migrations with custom environment
	_, err = client.MigrateApply(ctx, &atlasexec.MigrateApplyParams{
		Env:       "gorm",
		ConfigURL: fmt.Sprintf("file://%s", configPath),
		DirURL:    fmt.Sprintf("file://%s", migrationsPath),
		URL:       p.connectionURL,
	})
	if err != nil {
		return fmt.Errorf("failed to apply atlas migrations: %w", err)
	}

	return nil
}

// findModuleRoot walks up the directory tree to find the go.mod file
func findModuleRoot() (string, error) {
	// Start from the current file's directory
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("failed to get current file path")
	}

	dir := filepath.Dir(currentFile)

	// Walk up the directory tree
	for {
		goModPath := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached the root directory
			break
		}
		dir = parent
	}

	return "", fmt.Errorf("go.mod not found in any parent directory")
}
