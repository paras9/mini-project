package database

import (
	"mini2/models"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// Mocking gorm.DB
type MockDB struct {
	mock.Mock
}

func (m *MockDB) DB() *gorm.DB {
	args := m.Called()
	return args.Get(0).(*gorm.DB)
}

func TestInitDB(t *testing.T) {
	// Mocking the database connection
	mockDB := new(MockDB)
	mockDB.On("DB").Return(&gorm.DB{})

	// Initialize the DB
	InitDB("mock_dsn")

	// Validate if the DB connection was initialized
	assert.NotNil(t, DB, "DB should not be nil")

	// Validate that the mock method was called
	mockDB.AssertExpectations(t)
}

func TestMigrate(t *testing.T) {
	// Mocking DB auto-migration
	mockDB := new(MockDB)
	mockDB.On("DB").Return(&gorm.DB{})

	// Calling Migrate function
	Migrate(&models.Device{})

	// Assert migration was successful
	mockDB.AssertExpectations(t)
}
