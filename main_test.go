package main

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestInitDB(t *testing.T) {
	// Set up an in-memory SQLite database for testing
	dsn := "file::memory:?cache=shared"
	var err error
	db, err = gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to SQLite: %v", err)
	}

	err = db.AutoMigrate(&Device{})
	if err != nil {
		t.Fatalf("Failed to migrate database: %v", err)
	}

	// Verify that the Device table exists
	if !db.Migrator().HasTable(&Device{}) {
		t.Fatalf("Device table does not exist after migration")
	}
}

func TestUploadHandler(t *testing.T) {
	// Create a sample CSV file
	csvContent := "ID,DeviceName,DeviceType,Brand,Model,OS,OSVersion,PurchaseDate,WarrantyEnd,Status,Price\n" +
		"1,Device1,Type1,Brand1,Model1,OS1,1.0,2022-01-01,2023-01-01,Active,100.0\n"
	tempFile, err := os.CreateTemp("", "test_upload_*.csv")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())
	tempFile.WriteString(csvContent)
	tempFile.Close()

	// Set up a mock HTTP request
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	fileWriter, err := writer.CreateFormFile("file", tempFile.Name())
	if err != nil {
		t.Fatalf("Failed to create form file: %v", err)
	}
	file, err := os.Open(tempFile.Name())
	if err != nil {
		t.Fatalf("Failed to open temp file: %v", err)
	}
	defer file.Close()
	_, err = io.Copy(fileWriter, file)
	if err != nil {
		t.Fatalf("Failed to copy file content: %v", err)
	}
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()

	// Call the handler
	uploadHandler(rec, req)

	// Verify the response
	res := rec.Result()
	if res.StatusCode != http.StatusAccepted {
		t.Errorf("Expected status 202, got %d", res.StatusCode)
	}
}

func TestGetFilteredEntriesHandler(t *testing.T) {
	// Set up an in-memory SQLite database and insert test data
	dsn := "file::memory:?cache=shared"
	var err error
	db, err = gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to SQLite: %v", err)
	}
	db.AutoMigrate(&Device{})

	devices := []Device{
		{ID: 1, DeviceName: "Device1", DeviceType: "Type1", Brand: "Brand1", Price: 100.0},
		{ID: 2, DeviceName: "Device2", DeviceType: "Type2", Brand: "Brand2", Price: 200.0},
	}
	db.Create(&devices)

	// Set up a mock HTTP request
	req := httptest.NewRequest(http.MethodGet, "/entries?page=1&deviceType=Type1", nil)
	rec := httptest.NewRecorder()

	// Call the handler
	getFilteredEntriesHandler(rec, req)

	// Verify the response
	res := rec.Result()
	if res.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", res.StatusCode)
	}

	var response []Device
	err = json.NewDecoder(res.Body).Decode(&response)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(response) != 1 || response[0].DeviceName != "Device1" {
		t.Errorf("Unexpected response: %+v", response)
	}
}
