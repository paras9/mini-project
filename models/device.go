package models

// Device struct represents the database schema
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
