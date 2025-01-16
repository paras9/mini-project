package main

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Device struct {
	ID           int `gorm:"primaryKey"`
	DeviceName   string
	DeviceType   string
	Brand        string
	Model        string
	OS           string
	OSVersion    string
	PurchaseDate string
	WarrantyEnd  string
	Status       string
	Price        float64
}

var (
	db            *gorm.DB
	totalRows     int64
	processedRows int64
	mu            sync.Mutex
)

func initDB() {
	var err error
	dsn := "host=db user=postgres password=yourpassword dbname=devices port=5432 sslmode=disable"
	maxRetries := 5
	retryInterval := 3 * time.Second

	for retries := 0; retries < maxRetries; retries++ {
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err == nil {
			break
		}
		log.Printf("Failed to connect to database, retrying... (%d/%d)", retries+1, maxRetries)
		time.Sleep(retryInterval)
		retryInterval *= 3 // Exponential backoff
	}

	if err != nil {
		log.Fatalf("Failed to connect to database after %d retries: %v", maxRetries, err)
	}

	err = db.AutoMigrate(&Device{})
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to get DB from GORM: %v", err)
	}
	sqlDB.SetMaxOpenConns(200)
	sqlDB.SetMaxIdleConns(100)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
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

	_, err = bufio.NewReader(file).WriteTo(tempFile)
	if err != nil {
		http.Error(w, "Failed to save file content", http.StatusInternalServerError)
		return
	}

	go processFile("uploaded.csv")

	w.WriteHeader(http.StatusAccepted)
	fmt.Fprintln(w, "File upload started.")
}

func processFile(filePath string) {
	start := time.Now()
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(bufio.NewReader(file))
	_, err = reader.Read() // Skip header
	if err != nil {
		log.Fatalf("Failed to read CSV header: %v", err)
	}

	chunks := make(chan []string, 1000) // Buffered channel for efficiency
	wg := sync.WaitGroup{}
	numWorkers := 25 // Adjust based on available CPU cores

	// Create a pool of workers for processing chunks
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go processChunk(chunks, &wg)
	}

	// Concurrently read the file and send chunks to workers
	go func() {
		for {
			record, err := reader.Read()
			if err != nil {
				if err.Error() == "EOF" {
					break
				}
				log.Printf("Skipping record due to error: %v", err)
				continue
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

	var devices []Device
	batchSize := 500 // Reduce batch size to avoid overloading the database

	for record := range chunks {
		price, _ := strconv.ParseFloat(record[10], 64)
		device := Device{
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

		// Insert batch when the batch size is reached
		if len(devices) >= batchSize {
			insertBatch(devices)
			devices = devices[:0]
		}
	}

	// Insert any remaining devices
	if len(devices) > 0 {
		insertBatch(devices)
	}
}

func insertBatch(devices []Device) {
	maxRetries := 3
	retryInterval := 2 * time.Second

	for retries := 0; retries < maxRetries; retries++ {
		err := db.Create(&devices).Error
		if err == nil {
			atomic.AddInt64(&processedRows, int64(len(devices)))
			fmt.Printf("\rRows processed: %d/%d", processedRows, totalRows)
			return
		}

		log.Printf("Failed to insert batch, retrying... (%d/%d)", retries+1, maxRetries)
		time.Sleep(retryInterval)
		retryInterval *= 2 // Exponential backoff
	}

	log.Printf("Failed to insert batch after %d retries", maxRetries)
}

func parseInt(s string) int {
	v, _ := strconv.Atoi(s)
	return v
}

func getFilteredEntriesHandler(w http.ResponseWriter, r *http.Request) {
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

	var devices []Device
	query := db.Order("id ASC").Limit(limit).Offset(offset)

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
		http.Error(w, "Failed to fetch records", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(devices)
	if err != nil {
		http.Error(w, "Failed to encode records", http.StatusInternalServerError)
		return
	}
}

func main() {
	initDB()

	http.HandleFunc("/upload", uploadHandler)
	http.HandleFunc("/entries", getFilteredEntriesHandler)
	http.Handle("/", http.FileServer(http.Dir("./frontend")))

	fmt.Println("Server started at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
