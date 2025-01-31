package handlers

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestUploadHandler(t *testing.T) {
	logger := logrus.New()
	processed := make(chan bool, 1) // Use a channel to track function execution

	processFileFunc := func(file io.ReadCloser) {
		defer file.Close()
		processed <- true // Send true when function is executed
	}

	handler := UploadHandler(logger, processFileFunc)

	// Create a fake file
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, _ := writer.CreateFormFile("file", "test.csv")
	part.Write([]byte("test data"))
	writer.Close()

	r := httptest.NewRequest(http.MethodPost, "/upload", &buf)
	r.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, r)

	resp := w.Result()
	assert.Equal(t, http.StatusAccepted, resp.StatusCode)

	// Wait for the goroutine to process the file
	select {
	case <-processed:
		// Success, processFileFunc was called
	case <-time.After(1 * time.Second):
		t.Error("processFileFunc was not called")
	}
}
