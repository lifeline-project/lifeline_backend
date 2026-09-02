package repository

import (
	"lifeline_backend/internal/models"

	"gorm.io/gorm"
)

type PharmacyRepository struct {
	db *gorm.DB
}

func NewPharmacyRepository(db *gorm.DB) *PharmacyRepository {
	return &PharmacyRepository{db: db}
}

// FindNearby retrieves pharmacies within the specified radius (in meters) of the given coordinates.
// It uses the SQL Haversine formula for database-agnostic spherical distance calculations.
func (r *PharmacyRepository) FindNearby(lat, lon float64, radiusMeters float64) ([]models.PharmacyProfile, error) {
	var profiles []models.PharmacyProfile

	// SQL Haversine query (earth radius approx 6,371,000 meters)
	err := r.db.Raw(`
		SELECT user_id, pharmacy_name, address, latitude, longitude
		FROM pharmacy_profiles
		WHERE (6371000 * acos(
			cos(radians(?)) * cos(radians(latitude)) * cos(radians(longitude) - radians(?)) +
			sin(radians(?)) * sin(radians(latitude))
		)) <= ?
	`, lat, lon, lat, radiusMeters).Scan(&profiles).Error

	return profiles, err
}
