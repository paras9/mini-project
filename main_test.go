/*
package main

import (

	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

)

	func TestUploadHandler(t *testing.T) {
		// Create a temporary CSV file for testing
		tempFile, err := os.CreateTemp("", "test_upload_*.csv")
		if err != nil {
			t.Fatalf("Failed to create temporary file: %v", err)
		}
		defer os.Remove(tempFile.Name())

		_, err = tempFile.WriteString(`ID,DeviceName,DeviceType,Brand,Model,OS,OSVersion,PurchaseDate,WarrantyEnd,Status,Price\n1,TestDevice,Electronics,BrandX,ModelY,Android,11.0,2023-01-01,2024-01-01,Active,199.99`)
		if err != nil {
			t.Fatalf("Failed to write to temporary file: %v", err)
		}
		tempFile.Close()

		// Prepare a form file request
		file, err := os.Open(tempFile.Name())
		if err != nil {
			t.Fatalf("Failed to open temporary file: %v", err)
		}
		defer file.Close()

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, err := writer.CreateFormFile("file", tempFile.Name())
		if err != nil {
			t.Fatalf("Failed to create form file: %v", err)
		}
		_, err = io.Copy(part, file)
		if err != nil {
			t.Fatalf("Failed to copy file content: %v", err)
		}
		writer.Close()

		req := httptest.NewRequest(http.MethodPost, "/upload", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())

		rec := httptest.NewRecorder()
		http.HandlerFunc(uploadHandler).ServeHTTP(rec, req)

		if rec.Code != http.StatusAccepted {
			t.Errorf("Expected status %d, got %d", http.StatusAccepted, rec.Code)
		}
	}

	func TestGetFilteredEntriesHandler(t *testing.T) {
		// Seed the database with test data
		testDevices := []Device{
			{ID: 1, DeviceName: "Device1", DeviceType: "Electronics", Brand: "BrandA", Model: "Model1", OS: "Android", OSVersion: "10", PurchaseDate: "2022-01-01", WarrantyEnd: "2023-01-01", Status: "Active", Price: 299.99},
			{ID: 2, DeviceName: "Device2", DeviceType: "Accessory", Brand: "BrandB", Model: "Model2", OS: "iOS", OSVersion: "14", PurchaseDate: "2022-02-01", WarrantyEnd: "2023-02-01", Status: "Inactive", Price: 199.99},
		}
		db.Create(&testDevices)

		req := httptest.NewRequest(http.MethodGet, "/entries?page=1&deviceType=Electronics", nil)
		rec := httptest.NewRecorder()
		http.HandlerFunc(getFilteredEntriesHandler).ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
		}

		responseBody := rec.Body.String()
		if !strings.Contains(responseBody, "Device1") {
			t.Errorf("Expected response to contain Device1, got %s", responseBody)
		}
	}

	func TestProcessFile(t *testing.T) {
		// Create a temporary CSV file for testing
		tempFile, err := os.CreateTemp("", "test_process_*.csv")
		if err != nil {
			t.Fatalf("Failed to create temporary file: %v", err)
		}
		defer os.Remove(tempFile.Name())

		_, err = tempFile.WriteString(`ID,DeviceName,DeviceType,Brand,Model,OS,OSVersion,PurchaseDate,WarrantyEnd,Status,Price\n1,TestDevice,Electronics,BrandX,ModelY,Android,11.0,2023-01-01,2024-01-01,Active,199.99`)
		if err != nil {
			t.Fatalf("Failed to write to temporary file: %v", err)
		}
		tempFile.Close()

		// Call processFile
		processFile(tempFile.Name())

		// Verify the database
		var count int64
		db.Model(&Device{}).Count(&count)
		if count != 1 {
			t.Errorf("Expected 1 device in the database, got %d", count)
		}
	}
*/
package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestInitDB(t *testing.T) {
	var err error
	// Use SQLite for testing
	db, err = gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	err = db.AutoMigrate(&Device{})
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	// Cleanup
	os.Remove("test.db")
}

func TestUploadHandler(t *testing.T) {
	initDB()
	defer os.Remove("uploaded.csv")

	// Mock CSV file
	csvData := `ID,DeviceName,DeviceType,Brand,Model,OS,OSVersion,PurchaseDate,WarrantyEnd,Status,Price
1,Device1,Type1,Brand1,Model1,OS1,1.0,2022-01-01,2023-01-01,Active,1000.50
`
	req := httptest.NewRequest("POST", "/upload", bytes.NewReader([]byte(csvData)))
	req.Header.Set("Content-Type", "multipart/form-data")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(uploadHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusAccepted {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusAccepted)
	}

	if _, err := os.Stat("uploaded.csv"); os.IsNotExist(err) {
		t.Errorf("uploaded.csv file was not created")
	}
}

func TestGetFilteredEntriesHandler(t *testing.T) {
	initDB()

	// Insert mock data
	devices := []Device{
		{ID: 1, DeviceName: "Device1", DeviceType: "Type1", Brand: "Brand1", OS: "OS1", Price: 100.0},
		{ID: 2, DeviceName: "Device2", DeviceType: "Type2", Brand: "Brand2", OS: "OS2", Price: 200.0},
	}
	db.Create(&devices)

	req := httptest.NewRequest("GET", "/entries?page=1&deviceType=Type1", nil)
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(getFilteredEntriesHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var result []Device
	err := json.Unmarshal(rr.Body.Bytes(), &result)
	if err != nil {
		t.Errorf("Failed to decode response: %v", err)
	}

	if len(result) != 1 || result[0].DeviceName != "Device1" {
		t.Errorf("Filtered results are incorrect: got %v", result)
	}
}
