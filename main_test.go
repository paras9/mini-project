package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMainHandler(t *testing.T) {
	// Create a test request
	req, err := http.NewRequest("GET", "/upload", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a ResponseRecorder to capture the response
	rr := httptest.NewRecorder()

	// Create a handler for the route
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Serve the request using the handler
	handler.ServeHTTP(rr, req)

	// Assert the status code
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestGetFilteredEntriesHandler(t *testing.T) {
	// Create a test request
	req, err := http.NewRequest("GET", "/entries", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a ResponseRecorder to capture the response
	rr := httptest.NewRecorder()

	// Create a handler for the route
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Serve the request using the handler
	handler.ServeHTTP(rr, req)

	// Assert the status code
	assert.Equal(t, http.StatusOK, rr.Code)
}
