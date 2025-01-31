package handlers

import (
	"io"
	"net/http"

	"github.com/sirupsen/logrus"
)

func UploadHandler(logger *logrus.Logger, processFileFunc func(io.ReadCloser)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}

		file, _, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "Failed to read file", http.StatusBadRequest)
			return
		}

		// Pass the file to processFileFunc and defer closure within it
		go func() {
			defer file.Close() // Close the file after processing
			processFileFunc(file)
		}()

		w.WriteHeader(http.StatusAccepted)
		logger.Info("File upload started")
	}
}
