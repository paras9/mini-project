package handlers

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"mini2/config"
	"mini2/database"
	"mini2/models"

	"github.com/sirupsen/logrus"
)

var (
	totalRows     int64
	processedRows int64
)

func ProcessFile(file io.ReadCloser) {
	// Get the logger instance
	logger := config.GetLogger()

	defer file.Close()

	start := time.Now()
	reader := csv.NewReader(file)

	// Read the header
	_, err := reader.Read()
	if err != nil {
		logger.Fatalf("Failed to read CSV header: %v", err)
	}

	chunks := make(chan []string, 1000)
	wg := sync.WaitGroup{}
	numWorkers := 20

	// Start worker goroutines
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go processChunk(chunks, &wg, logger)
	}

	// Read the CSV file and send records to chunks channel
	go func() {
		defer close(chunks)
		for {
			record, err := reader.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				logger.Printf("Error reading record: %v", err)
				continue
			}
			atomic.AddInt64(&totalRows, 1)
			chunks <- record
		}
	}()

	//ensures worker finish before counting
	wg.Wait()
	logger.Infof("Processing completed in %v", time.Since(start))
	logger.Infof("Total rows: %d, Processed rows: %d", totalRows, processedRows)
}

func processChunk(chunks chan []string, wg *sync.WaitGroup, logger *logrus.Logger) {
	defer wg.Done()

	var devices []models.Device
	batchSize := 500

	for record := range chunks {
		price, err := strconv.ParseFloat(record[10], 64)
		if err != nil {
			logger.Printf("Invalid price in record: %v", record)
			continue
		}

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
			insertBatch(devices, logger)
			devices = devices[:0]
		}
	}

	if len(devices) > 0 {
		insertBatch(devices, logger)
	}
}

func insertBatch(devices []models.Device, logger *logrus.Logger) {
	err := database.DB.Create(&devices).Error
	if err != nil {
		logger.Printf("Failed to insert batch: %v", err)
		return
	}
	atomic.AddInt64(&processedRows, int64(len(devices)))
}

func parseInt(s string) int {
	v, err := strconv.Atoi(s)
	if err != nil {
		fmt.Errorf("Invalid integer value: %s", s)
		return 0
	}
	return v
}
