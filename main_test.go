package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUploadRouteHandler(t *testing.T) {
	// Create a test request
	req, err := http.NewRequest("GET", "/upload", nil)
	assert.NoError(t, err)

	// Create a ResponseRecorder to capture the response
	rr := httptest.NewRecorder()

	// Use the actual handler (replace with your actual upload handler)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Upload route works"))
	})

	// Serve the request using the handler
	handler.ServeHTTP(rr, req)

	// Assert the status code
	assert.Equal(t, http.StatusOK, rr.Code)

	// Assert the response body
	assert.Equal(t, "Upload route works", rr.Body.String())
}

func TestEntriesRouteHandler(t *testing.T) {
	// Create a test request
	req, err := http.NewRequest("GET", "/entries", nil)
	assert.NoError(t, err)

	// Create a ResponseRecorder to capture the response
	rr := httptest.NewRecorder()

	// Use the actual handler (replace with your actual entries handler)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Entries route works"))
	})

	// Serve the request using the handler
	handler.ServeHTTP(rr, req)

	// Assert the status code
	assert.Equal(t, http.StatusOK, rr.Code)

	// Assert the response body
	assert.Equal(t, "Entries route works", rr.Body.String())
}
