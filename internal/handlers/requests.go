package handlers

import (
	"net/http"
	"strconv"
	"time"

	"lifeline_backend/internal/models"
	"lifeline_backend/internal/repository"
	wshub "lifeline_backend/internal/websocket"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type RequestHandler struct {
	DB           *gorm.DB
	PharmacyRepo *repository.PharmacyRepository
	Hub          *wshub.Hub
}

func NewRequestHandler(db *gorm.DB, hub *wshub.Hub) *RequestHandler {
	return &RequestHandler{
		DB:           db,
		PharmacyRepo: repository.NewPharmacyRepository(db),
		Hub:          hub,
	}
}

type CreateRequestInput struct {
	Note                 string  `json:"note"`
	PrescriptionImageURL string  `json:"prescription_image_url"`
	Latitude             float64 `json:"latitude" binding:"required"`
	Longitude            float64 `json:"longitude" binding:"required"`
}

type CloseRequestInput struct {
	Status                string `json:"status" binding:"required,oneof=FULFILLED CANCELLED"`
	FulfilledByPharmacyID *uint  `json:"fulfilled_by_pharmacy_id"`
}

// CreateRequest handles new emergency request submissions from patients
func (h *RequestHandler) CreateRequest(c *gin.Context) {
	userIDVal, _ := c.Get("userID")
	userID, ok := userIDVal.(uint)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req CreateRequestInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	emergencyReq := models.EmergencyRequest{
		PatientID:            userID,
		Note:                 req.Note,
		PrescriptionImageURL: req.PrescriptionImageURL,
		Status:               "OPEN",
		RequestLatitude:      req.Latitude,
		RequestLongitude:     req.Longitude,
		CreatedAt:            time.Now(),
	}

	if err := h.DB.Create(&emergencyReq).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save emergency request"})
		return
	}

	// Query nearby pharmacies (default 3km radius = 3000 meters)
	nearby, err := h.PharmacyRepo.FindNearby(req.Latitude, req.Longitude, 3000.0)
	if err != nil {
		// Log error, but proceed with empty nearby and return request
		nearby = []models.PharmacyProfile{}
	}

	// Broadcast to nearby pharmacies in the background
	go h.Hub.BroadcastToNearby(emergencyReq, nearby)

	c.JSON(http.StatusCreated, gin.H{
		"message":        "Emergency request broadcasted successfully",
		"request":        emergencyReq,
		"nearby_count":   len(nearby),
		"notified_stores": nearby,
	})
}

// CloseRequest handles closing an existing request (marking as FULFILLED or CANCELLED)
func (h *RequestHandler) CloseRequest(c *gin.Context) {
	userIDVal, _ := c.Get("userID")
	userID, ok := userIDVal.(uint)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	idParam := c.Param("id")
	reqID, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request ID"})
		return
	}

	var input CloseRequestInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var emergencyReq models.EmergencyRequest
	if err := h.DB.First(&emergencyReq, reqID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Request not found"})
		return
	}

	// Only the patient who created the request can close it
	if emergencyReq.PatientID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access forbidden: you do not own this request"})
		return
	}

	emergencyReq.Status = input.Status
	if input.FulfilledByPharmacyID != nil && *input.FulfilledByPharmacyID > 0 {
		emergencyReq.FulfilledByPharmacyID = input.FulfilledByPharmacyID
	}
	if err := h.DB.Save(&emergencyReq).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to close request"})
		return
	}

	// Notify the pharmacy via WebSocket
	if emergencyReq.FulfilledByPharmacyID != nil {
		go h.Hub.SendToUser(*emergencyReq.FulfilledByPharmacyID, "REQUEST_FULFILLED", emergencyReq)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Request closed successfully",
		"request": emergencyReq,
	})
}

// ListPatientRequests lists all requests submitted by the logged-in patient
func (h *RequestHandler) ListPatientRequests(c *gin.Context) {
	userIDVal, _ := c.Get("userID")
	userID, ok := userIDVal.(uint)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var requests []models.EmergencyRequest
	if err := h.DB.Where("patient_id = ?", userID).Order("created_at desc").Find(&requests).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve requests"})
		return
	}

	c.JSON(http.StatusOK, requests)
}

type IncomingRequestForPharmacy struct {
	ID                   uint      `json:"id"`
	PatientID            uint      `json:"patient_id"`
	PatientName          string    `json:"patient_name"`
	Note                 string    `json:"note"`
	PrescriptionImageURL string    `json:"prescription_image_url"`
	Status               string    `json:"status"`
	RequestLatitude      float64   `json:"request_latitude"`
	RequestLongitude     float64   `json:"request_longitude"`
	CreatedAt            time.Time `json:"created_at"`
}

// ListPharmacyIncomingRequests lists open emergency requests for logged-in pharmacy
func (h *RequestHandler) ListPharmacyIncomingRequests(c *gin.Context) {
	userIDVal, _ := c.Get("userID")
	_, ok := userIDVal.(uint)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var requests []models.EmergencyRequest
	if err := h.DB.Where("status = ?", "OPEN").Order("created_at desc").Limit(20).Find(&requests).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch incoming requests"})
		return
	}

	var result []IncomingRequestForPharmacy
	for _, req := range requests {
		var patient models.User
		_ = h.DB.Select("name").First(&patient, req.PatientID)
		pName := patient.Name
		if pName == "" {
			pName = "Patient #" + strconv.Itoa(int(req.PatientID))
		}
		result = append(result, IncomingRequestForPharmacy{
			ID:                   req.ID,
			PatientID:            req.PatientID,
			PatientName:          pName,
			Note:                 req.Note,
			PrescriptionImageURL: req.PrescriptionImageURL,
			Status:               req.Status,
			RequestLatitude:      req.RequestLatitude,
			RequestLongitude:     req.RequestLongitude,
			CreatedAt:            req.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, result)
}

