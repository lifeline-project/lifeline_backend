package handlers

import (
	"net/http"
	"strconv"
	"time"

	"lifeline_backend/internal/models"
	wshub "lifeline_backend/internal/websocket"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ResponseHandler struct {
	DB  *gorm.DB
	Hub *wshub.Hub
}

func NewResponseHandler(db *gorm.DB, hub *wshub.Hub) *ResponseHandler {
	return &ResponseHandler{
		DB:  db,
		Hub: hub,
	}
}

type CreateResponseInput struct {
	RequestID           uint    `json:"request_id" binding:"required"`
	Availability        string  `json:"availability" binding:"required,oneof=YES NO"`
	Price               float64 `json:"price"`
	EstimatedPickupTime string  `json:"estimated_pickup_time"`
	Notes               string  `json:"notes"`
}

// CreateResponse handles a pharmacy's availability response to a request
func (h *ResponseHandler) CreateResponse(c *gin.Context) {
	userIDVal, _ := c.Get("userID")
	userID, ok := userIDVal.(uint)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req CreateResponseInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify request exists and is OPEN
	var emergencyReq models.EmergencyRequest
	if err := h.DB.First(&emergencyReq, req.RequestID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Emergency request not found"})
		return
	}

	if emergencyReq.Status != "OPEN" {
		c.JSON(http.StatusConflict, gin.H{"error": "Cannot respond to a closed or fulfilled request"})
		return
	}

	// Fetch pharmacy profile details
	var pharmacy models.PharmacyProfile
	if err := h.DB.Where("user_id = ?", userID).First(&pharmacy).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Pharmacy profile not found"})
		return
	}

	response := models.PharmacyResponse{
		PharmacyID:          userID,
		RequestID:           req.RequestID,
		Availability:        req.Availability,
		Price:               req.Price,
		EstimatedPickupTime: req.EstimatedPickupTime,
		Notes:               req.Notes,
		RespondedAt:         time.Now(),
	}

	// Upsert response (if pharmacy already responded, update fields)
	var existingResponse models.PharmacyResponse
	err := h.DB.Where("pharmacy_id = ? AND request_id = ?", userID, req.RequestID).First(&existingResponse).Error
	if err == nil {
		existingResponse.Availability = req.Availability
		if req.Price > 0 {
			existingResponse.Price = req.Price
		}
		if req.EstimatedPickupTime != "" {
			existingResponse.EstimatedPickupTime = req.EstimatedPickupTime
		}
		if req.Notes != "" {
			existingResponse.Notes = req.Notes
		}
		existingResponse.RespondedAt = time.Now()
		if err := h.DB.Save(&existingResponse).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update response"})
			return
		}
		response = existingResponse
	} else {
		if err := h.DB.Create(&response).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to record response"})
			return
		}
	}

	// Relay the response to the patient in real time over WebSocket
	go h.Hub.SendToUser(emergencyReq.PatientID, "NEW_PHARMACY_RESPONSE", map[string]interface{}{
		"response": response,
		"pharmacy": pharmacy,
	})

	c.JSON(http.StatusCreated, gin.H{
		"message":  "Response submitted successfully",
		"response": response,
	})
}

// ListResponsesForRequest retrieves all pharmacy responses for a given request
func (h *ResponseHandler) ListResponsesForRequest(c *gin.Context) {
	idParam := c.Param("id")
	reqID, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request ID"})
		return
	}

	// Verify request exists
	var emergencyReq models.EmergencyRequest
	if err := h.DB.First(&emergencyReq, reqID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Emergency request not found"})
		return
	}

	// Fetch responses and join with pharmacy profiles
	type ResponseDetail struct {
		models.PharmacyResponse
		PharmacyName string  `json:"pharmacy_name"`
		Address      string  `json:"address"`
		Latitude     float64 `json:"latitude"`
		Longitude    float64 `json:"longitude"`
	}

	var details []ResponseDetail
	err = h.DB.Table("pharmacy_responses").
		Select("pharmacy_responses.*, pharmacy_profiles.pharmacy_name, pharmacy_profiles.address, pharmacy_profiles.latitude, pharmacy_profiles.longitude").
		Joins("JOIN pharmacy_profiles ON pharmacy_profiles.user_id = pharmacy_responses.pharmacy_id").
		Where("pharmacy_responses.request_id = ?", reqID).
		Scan(&details).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve responses"})
		return
	}

	c.JSON(http.StatusOK, details)
}
