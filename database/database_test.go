package database_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"mini2/database"
)

type TestModel struct {
	ID   uint   `gorm:"primaryKey"`
	Name string `gorm:"size:255"`
}

func TestInitDB(t *testing.T) {
	dsn := "host=localhost user=postgres password=postgres dbname=testdb port=5432 sslmode=disable"

	// Initialize the database
	database.InitDB(dsn)

	// Check if the DB instance is initialized
	assert.NotNil(t, database.DB, "DB instance should not be nil")

	// Verify connection settings
	sqlDB, err := database.DB.DB()
	assert.NoError(t, err, "Should not error while getting sql.DB")
	assert.Equal(t, 200, sqlDB.Stats().MaxOpenConnections, "MaxOpenConns should be 200")
	assert.Equal(t, 100, sqlDB.Stats().Idle, "MaxIdleConns should be 100")
}

func TestMigrate(t *testing.T) {
	dsn := "host=localhost user=postgres password=postgres dbname=testdb port=5432 sslmode=disable"
	database.InitDB(dsn)

	// Migrate the TestModel
	database.Migrate(&TestModel{})

	// Verify migration by checking if the table exists
	var exists bool
	err := database.DB.Raw("SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = ?)", "test_models").Scan(&exists).Error
	assert.NoError(t, err, "Should not error while checking table existence")
	assert.True(t, exists, "Table 'test_models' should exist after migration")
}
