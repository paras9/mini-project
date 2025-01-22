package handlers

import (
	"encoding/csv"
	"log"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"mini2/database"
	"mini2/models"
)

var (
	totalRows     int64
	processedRows int64
)

func ProcessFile(filePath string) {
	start := time.Now()
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	_, err = reader.Read() // Skip header
	if err != nil {
		log.Fatalf("Failed to read CSV header: %v", err)
	}

	chunks := make(chan []string, 1000)
	wg := sync.WaitGroup{}
	numWorkers := 25

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go processChunk(chunks, &wg)
	}

	go func() {
		for {
			record, err := reader.Read()
			if err != nil {
				break
			}
			atomic.AddInt64(&totalRows, 1)
			chunks <- record
		}
		close(chunks)
	}()

	wg.Wait()
	log.Printf("Processing completed in %v", time.Since(start))
}

func processChunk(chunks chan []string, wg *sync.WaitGroup) {
	defer wg.Done()

	var devices []models.Device
	batchSize := 500

	for record := range chunks {
		price, _ := strconv.ParseFloat(record[10], 64)
		device := models.Device{
			ID:           parseInt(record[0]),
			DeviceName:   record[1],
			DeviceType:   record[2],
			Brand:        record[3],
			Model:        record[4],
			OS:           record[5],
			OSVersion:    record[6],
			PurchaseDate: record[7],
			WarrantyEnd:  record[8],
			Status:       record[9],
			Price:        price,
		}

		devices = append(devices, device)

		if len(devices) >= batchSize {
			insertBatch(devices)
			devices = devices[:0]
		}
	}

	if len(devices) > 0 {
		insertBatch(devices)
	}
}

func insertBatch(devices []models.Device) {
	err := database.DB.Create(&devices).Error
	if err != nil {
		log.Printf("Failed to insert batch: %v", err)
	}
}

func parseInt(s string) int {
	v, _ := strconv.Atoi(s)
	return v
}
