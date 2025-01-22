package handlers

import (
	"net/http"
	"os"

	"github.com/sirupsen/logrus"
)

func UploadHandler(logger *logrus.Logger, processFileFunc func(string)) http.HandlerFunc {
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
		defer file.Close()

		tempFile, err := os.Create("uploaded.csv")
		if err != nil {
			http.Error(w, "Failed to save file", http.StatusInternalServerError)
			return
		}
		defer tempFile.Close()

		_, err = tempFile.ReadFrom(file)
		if err != nil {
			http.Error(w, "Failed to save file content", http.StatusInternalServerError)
			return
		}

		go processFileFunc("uploaded.csv")

		w.WriteHeader(http.StatusAccepted)
		logger.Info("File upload started")
	}
}
