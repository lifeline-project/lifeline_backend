package models

import (
	"time"
)

// User - The main account (Patient or Pharmacy)
type User struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	Name         string    `json:"name"`
	Email        string    `json:"email" gorm:"unique"`
	PasswordHash string    `json:"-"` // "-" means never send password in API response
	Phone        string    `json:"phone"`
	Role         string    `json:"role"` // 'PATIENT', 'PHARMACY'
	CreatedAt    time.Time `json:"created_at"`
}

// PharmacyProfile - Location data (Linked to User)
type PharmacyProfile struct {
	UserID       uint      `json:"user_id" gorm:"primaryKey"`
	PharmacyName string    `json:"pharmacy_name"`
	Address      string    `json:"address"`
	Latitude     float64   `json:"latitude"`
	Longitude    float64   `json:"longitude"`
	// We don't put the 'Location' geography type here. 
	// We handle that with raw SQL queries in the repository layer.
}

// EmergencyRequest - The core feature
type EmergencyRequest struct {
	ID                    uint      `json:"id" gorm:"primaryKey"`
	PatientID             uint      `json:"patient_id"`
	PrescriptionImageURL  string    `json:"prescription_image_url"`
	Note                  string    `json:"note"`
	Status                string    `json:"status"` // 'OPEN', 'FULFILLED', 'CANCELLED'
	FulfilledByPharmacyID *uint     `json:"fulfilled_by_pharmacy_id"` 
	CreatedAt             time.Time `json:"created_at"`
}