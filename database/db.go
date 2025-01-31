package database

import (
	"time"

	"mini2/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB(dsn string) {
	// Get the logger instance
	logger := config.GetLogger()

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		// Log the error using the logger
		logger.Fatalf("Failed to connect to database: %v", err)
	}

	sqlDB, _ := DB.DB()
	sqlDB.SetMaxOpenConns(200)
	sqlDB.SetMaxIdleConns(100)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)

	// Log the successful database connection
	logger.Infof("Successfully connected to the database.")
}

func Migrate(models ...interface{}) {
	// Get the logger instance
	logger := config.GetLogger()

	err := DB.AutoMigrate(models...)
	if err != nil {
		// Log the error using the logger
		logger.Fatalf("Failed to migrate database: %v", err)
	}

	// Log the successful migration
	logger.Infof("Database migration completed successfully.")
}
