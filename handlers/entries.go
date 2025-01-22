package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"mini2/database"
	"mini2/models"

	"github.com/sirupsen/logrus"
)

func GetFilteredEntriesHandler(logger *logrus.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		deviceType := r.URL.Query().Get("deviceType")
		deviceName := r.URL.Query().Get("deviceName")
		os := r.URL.Query().Get("os")
		brand := r.URL.Query().Get("brand")
		idRange := r.URL.Query().Get("idRange")

		if page < 1 {
			page = 1
		}

		limit := 100
		offset := (page - 1) * limit

		var devices []models.Device
		query := database.DB.Order("id ASC").Limit(limit).Offset(offset)

		if deviceType != "" {
			query = query.Where("device_type = ?", deviceType)
		}
		if deviceName != "" {
			query = query.Where("device_name = ?", deviceName)
		}
		if os != "" {
			query = query.Where("os = ?", os)
		}
		if brand != "" {
			query = query.Where("brand = ?", brand)
		}
		if idRange != "" {
			ids := strings.Split(idRange, "-")
			if len(ids) == 2 {
				startID, _ := strconv.Atoi(ids[0])
				endID, _ := strconv.Atoi(ids[1])
				query = query.Where("id BETWEEN ? AND ?", startID, endID)
			}
		}

		result := query.Find(&devices)
		if result.Error != nil {
			logger.Errorf("Failed to fetch records: %v", result.Error)
			http.Error(w, "Failed to fetch records", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(devices)
		if err != nil {
			logger.Errorf("Failed to encode records: %v", err)
			http.Error(w, "Failed to encode records", http.StatusInternalServerError)
		}
	}
}
