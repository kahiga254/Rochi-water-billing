package handlers

import (
	"net/http"
	"strconv"

	"waterbilling/backend/models"
	"waterbilling/backend/services"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Helper function for not implemented endpoints
func notImplemented(c *gin.Context, message string) {
	ErrorResponse(c, http.StatusNotImplemented, message, nil)
}

type SMSHandler struct {
	billingService *services.BillingService
	smsService     *services.SMSService
}

func NewSMSHandler(billingService *services.BillingService, smsService *services.SMSService) *SMSHandler {
	return &SMSHandler{
		billingService: billingService,
		smsService:     smsService,
	}
}

// SendBillNotification sends SMS notification for a specific bill
func (h *SMSHandler) SendBillNotification(c *gin.Context) {
	billID := c.Param("billID")
	if billID == "" {
		BadRequest(c, "Bill ID is required", nil)
		return
	}

	_, err := primitive.ObjectIDFromHex(billID)
	if err != nil {
		BadRequest(c, "Invalid bill ID", err)
		return
	}

	// SMS service is not implemented yet
	notImplemented(c, "SMS service is not yet implemented. Configure Twilio or Africa's Talking in .env file.")
}

// BulkSendBillNotifications sends SMS notifications for multiple bills
func (h *SMSHandler) BulkSendBillNotifications(c *gin.Context) {
	var req BulkSMSRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "Invalid request data", err)
		return
	}

	// Check if SMS service is enabled
	if !h.smsService.IsEnabled() {
		ErrorResponse(c, http.StatusServiceUnavailable,
			"SMS service is not configured", nil)
		return
	}

	// Validate bill IDs
	if len(req.BillIDs) == 0 && !req.SendToUnpaid {
		BadRequest(c, "Either provide bill IDs or set send_to_unpaid to true", nil)
		return
	}

	if len(req.BillIDs) > 100 {
		BadRequest(c, "Maximum 100 bills per batch", nil)
		return
	}

	var bills []models.Bill
	var err error

	if req.SendToUnpaid {
		// Get all unpaid bills
		bills, err = h.billingService.GetUnpaidBills()
		if err != nil {
			InternalServerError(c, "Failed to fetch unpaid bills", err)
			return
		}
	} else {
		// TODO: Fetch specific bills by IDs
		notImplemented(c, "Fetching specific bills by IDs not yet implemented")
		return
	}

	// TODO: Implement bulk SMS sending
	SuccessResponse(c, "Bulk SMS endpoint ready", gin.H{
		"total_bills":     len(bills),
		"action_required": "Implement bulk SMS sending logic",
	})
}

// SendPaymentConfirmation sends payment confirmation SMS
func (h *SMSHandler) SendPaymentConfirmation(c *gin.Context) {
	var req PaymentConfirmationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "Invalid request data", err)
		return
	}

	if req.MeterNumber == "" && req.BillID == "" {
		BadRequest(c, "Either meter number or bill ID is required", nil)
		return
	}

	if req.Amount <= 0 {
		BadRequest(c, "Payment amount must be greater than 0", nil)
		return
	}

	// Check if SMS service is enabled
	if !h.smsService.IsEnabled() {
		ErrorResponse(c, http.StatusServiceUnavailable,
			"SMS service is not configured", nil)
		return
	}

	// TODO: Implement payment confirmation SMS
	SuccessResponse(c, "Payment confirmation endpoint ready", gin.H{
		"status":         "ready",
		"next_steps":     "Implement payment confirmation logic",
		"amount":         req.Amount,
		"transaction_id": req.TransactionID,
	})
}

// GetSMSLogs gets SMS sending history
func (h *SMSHandler) GetSMSLogs(c *gin.Context) {
	// Build filter from query parameters
	filter := bson.M{}

	if meterNumber := c.Query("meterNumber"); meterNumber != "" {
		filter["meter_number"] = meterNumber
	}

	if messageType := c.Query("messageType"); messageType != "" {
		filter["message_type"] = messageType
	}

	// Parse date filters (simplified)
	if startDate := c.Query("startDate"); startDate != "" {
		filter["sent_at"] = bson.M{"$gte": startDate}
	}

	if endDate := c.Query("endDate"); endDate != "" {
		if startFilter, ok := filter["sent_at"].(bson.M); ok {
			startFilter["$lte"] = endDate
		} else {
			filter["sent_at"] = bson.M{"$lte": endDate}
		}
	}

	limitStr := c.DefaultQuery("limit", "50")
	limit := int64(50)
	if l, err := strconv.ParseInt(limitStr, 10, 64); err == nil && l > 0 {
		limit = l
	}

	// Check if SMS service is available
	if h.smsService == nil {
		InternalServerError(c, "SMS service not initialized", nil)
		return
	}

	logs, err := h.smsService.GetSMSLogs(filter, limit)
	if err != nil {
		InternalServerError(c, "Failed to fetch SMS logs", err)
		return
	}

	SuccessResponse(c, "SMS logs retrieved", logs)
}

// SendDisconnectionWarning sends disconnection warning SMS
func (h *SMSHandler) SendDisconnectionWarning(c *gin.Context) {
	// Get overdue bills
	bills, err := h.billingService.GetOverdueBills()
	if err != nil {
		InternalServerError(c, "Failed to fetch overdue bills", err)
		return
	}

	// Check if SMS service is enabled
	if !h.smsService.IsEnabled() {
		SuccessResponse(c, "SMS service not configured", gin.H{
			"message":             "SMS service is not configured. No messages were sent.",
			"overdue_bills_count": len(bills),
			"config_required":     "Configure Twilio or Africa's Talking in .env file",
		})
		return
	}

	// TODO: Implement actual disconnection warning sending
	// This would involve fetching customers for each bill and sending SMS

	response := gin.H{
		"message":             "Disconnection warning process would be initiated",
		"overdue_bills_count": len(bills),
		"action_required":     "Implement customer fetching and SMS sending logic",
		"sms_service_status":  "enabled",
	}

	SuccessResponse(c, "Disconnection warning endpoint ready", response)
}

// Request/Response DTOs
type BulkSMSRequest struct {
	BillIDs      []string `json:"bill_ids,omitempty"`
	SendToUnpaid bool     `json:"send_to_unpaid"`
	TemplateID   string   `json:"template_id,omitempty"`
}

type PaymentConfirmationRequest struct {
	MeterNumber   string  `json:"meter_number,omitempty"`
	BillID        string  `json:"bill_id,omitempty"`
	Amount        float64 `json:"amount" binding:"required"`
	TransactionID string  `json:"transaction_id"`
	ReceiptNumber string  `json:"receipt_number"`
}
