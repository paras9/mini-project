package handlers_test

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"mini2/handlers"

	"github.com/sirupsen/logrus"
)

func TestUploadHandler(t *testing.T) {
	logger := logrus.New()
	logger.SetOutput(io.Discard) // Discard logs during tests

	processFileFunc := func(filePath string) {
		// Mock processing logic, if needed
	}

	t.Run("Successful upload", func(t *testing.T) {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		fileWriter, err := writer.CreateFormFile("file", "test.csv")
		if err != nil {
			t.Fatalf("Failed to create form file: %v", err)
		}

		_, err = fileWriter.Write([]byte("test,data"))
		if err != nil {
			t.Fatalf("Failed to write to form file: %v", err)
		}
		writer.Close()

		req := httptest.NewRequest(http.MethodPost, "/upload", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())

		rec := httptest.NewRecorder()
		handler := handlers.UploadHandler(logger, processFileFunc)
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusAccepted {
			t.Errorf("Expected status %d, got %d", http.StatusAccepted, rec.Code)
		}

		if _, err := os.Stat("uploaded.csv"); os.IsNotExist(err) {
			t.Errorf("Expected file 'uploaded.csv' to exist, but it does not")
		}

		// Cleanup
		os.Remove("uploaded.csv")
	})

	t.Run("Invalid request method", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/upload", nil)
		rec := httptest.NewRecorder()
		handler := handlers.UploadHandler(logger, processFileFunc)
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusMethodNotAllowed {
			t.Errorf("Expected status %d, got %d", http.StatusMethodNotAllowed, rec.Code)
		}
	})

	t.Run("Missing file in request", func(t *testing.T) {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		writer.Close()

		req := httptest.NewRequest(http.MethodPost, "/upload", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())

		rec := httptest.NewRecorder()
		handler := handlers.UploadHandler(logger, processFileFunc)
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, rec.Code)
		}
	})

	t.Run("Error creating file on server", func(t *testing.T) {
		// Simulate error by setting a directory with no write permission
		os.Chmod(".", 0555)
		defer os.Chmod(".", 0755) // Reset permissions after test

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		fileWriter, err := writer.CreateFormFile("file", "test.csv")
		if err != nil {
			t.Fatalf("Failed to create form file: %v", err)
		}

		_, err = fileWriter.Write([]byte("test,data"))
		if err != nil {
			t.Fatalf("Failed to write to form file: %v", err)
		}
		writer.Close()

		req := httptest.NewRequest(http.MethodPost, "/upload", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())

		rec := httptest.NewRecorder()
		handler := handlers.UploadHandler(logger, processFileFunc)
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, rec.Code)
		}
	})
}
