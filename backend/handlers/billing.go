package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"waterbilling/backend/models"
	"waterbilling/backend/services"
	"waterbilling/backend/utils"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type BillingHandler struct {
	billingService *services.BillingService
}

func NewBillingHandler(billingService *services.BillingService) *BillingHandler {
	return &BillingHandler{
		billingService: billingService,
	}
}

// SubmitMeterReading submits a new meter reading
func (h *BillingHandler) SubmitMeterReading(c *gin.Context) {
	var req MeterReadingRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "Invalid meter reading data", err)
		return
	}

	// Validate required fields
	if req.MeterNumber == "" {
		BadRequest(c, "Meter number is required", nil)
		return
	}

	if req.CurrentReading <= 0 {
		BadRequest(c, "Current reading must be greater than 0", nil)
		return
	}

	// Set default values
	if req.ReadingDate.IsZero() {
		req.ReadingDate = time.Now()
	}

	if req.ReadingType == "" {
		req.ReadingType = "manual"
	}

	if req.ReadingMethod == "" {
		req.ReadingMethod = "mobile_app"
	}

	// Create meter reading model
	reading := &models.MeterReading{
		MeterNumber:    req.MeterNumber,
		CurrentReading: req.CurrentReading,
		ReadingDate:    req.ReadingDate,
		ReadingType:    req.ReadingType,
		ReadingMethod:  req.ReadingMethod,
		ReaderID:       req.ReaderID,
		ReaderName:     req.ReaderName,
		Location:       req.Location,
		MeterPhotoURL:  req.MeterPhotoURL,
		MeterCondition: req.MeterCondition,
		Notes:          req.Notes,
	}

	// Submit reading and generate bill
	bill, err := h.billingService.SubmitMeterReading(reading)
	if err != nil {
		if strings.Contains(err.Error(), "customer with meter number") {
			NotFound(c, "Customer not found")
		} else if strings.Contains(err.Error(), "current reading cannot be less than previous reading") {
			BadRequest(c, "Current reading cannot be less than previous reading", err)
		} else {
			InternalServerError(c, "Failed to submit meter reading", err)
		}
		return
	}

	CreatedResponse(c, "Meter reading submitted and bill generated successfully", bill)
}

// GetCustomerBills gets all bills for a customer
func (h *BillingHandler) GetCustomerBills(c *gin.Context) {
	meterNumber := c.Param("meterNumber")
	if meterNumber == "" {
		BadRequest(c, "Meter number is required", nil)
		return
	}

	status := c.Query("status")
	limit := c.DefaultQuery("limit", "50")

	var limitInt int64 = 50
	if limit != "" {
		if l, err := strconv.ParseInt(limit, 10, 64); err == nil && l > 0 {
			limitInt = l
		}
	}

	bills, err := h.billingService.GetCustomerBills(meterNumber, status, limitInt)
	if err != nil {
		InternalServerError(c, "Failed to fetch customer bills", err)
		return
	}

	SuccessResponse(c, "Customer bills retrieved", bills)
}

// GetCustomerReadingHistory gets reading history for a customer
func (h *BillingHandler) GetCustomerReadingHistory(c *gin.Context) {
	meterNumber := c.Param("meterNumber")
	if meterNumber == "" {
		BadRequest(c, "Meter number is required", nil)
		return
	}

	limit := c.DefaultQuery("limit", "12")
	limitInt, err := strconv.ParseInt(limit, 10, 64)
	if err != nil {
		limitInt = 12
	}

	readings, err := h.billingService.GetCustomerReadingHistory(meterNumber, limitInt)
	if err != nil {
		InternalServerError(c, "Failed to fetch reading history", err)
		return
	}

	SuccessResponse(c, "Reading history retrieved", readings)
}

// ProcessPayment processes a payment for a bill
func (h *BillingHandler) ProcessPayment(c *gin.Context) {
	billID := c.Param("billID")
	if billID == "" {
		BadRequest(c, "Bill ID is required", nil)
		return
	}

	var req PaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "Invalid payment data", err)
		return
	}

	// Validate required fields
	if req.Amount <= 0 {
		BadRequest(c, "Payment amount must be greater than 0", nil)
		return
	}

	if req.PaymentMethod == "" {
		BadRequest(c, "Payment method is required", nil)
		return
	}

	// Convert bill ID to ObjectID
	objectID, err := primitive.ObjectIDFromHex(billID)
	if err != nil {
		BadRequest(c, "Invalid bill ID", err)
		return
	}

	// Get bill details first to include in payment record
	// We'll need to fetch the bill to get customer details
	// For now, we'll create payment with minimal info
	payment := &models.Payment{
		BillID:        objectID,
		Amount:        req.Amount,
		PaymentMethod: req.PaymentMethod,
		TransactionID: req.TransactionID,
		ReceiptNumber: req.ReceiptNumber,
		PayerName:     req.PayerName,
		PayerPhone:    req.PayerPhone,
		CollectedBy:   req.CollectedBy,
		Notes:         req.Notes,
	}

	// Process payment
	if err := h.billingService.ProcessPayment(payment); err != nil {
		if strings.Contains(err.Error(), "bill not found") {
			NotFound(c, "Bill not found")
		} else if strings.Contains(err.Error(), "payment amount must be greater than 0") {
			BadRequest(c, "Payment amount must be greater than 0", err)
		} else {
			InternalServerError(c, "Failed to process payment", err)
		}
		return
	}

	SuccessResponse(c, "Payment processed successfully", payment)
}

// GetBillDetails gets details of a specific bill
func (h *BillingHandler) GetBillDetails(c *gin.Context) {
	// This would query the bills collection directly
	// For now, we'll implement a simple version
	// We'll need to inject the bills collection or expand the service

}

// GetOverdueBills gets all overdue bills
func (h *BillingHandler) GetOverdueBills(c *gin.Context) {
	bills, err := h.billingService.GetOverdueBills()
	if err != nil {
		InternalServerError(c, "Failed to fetch overdue bills", err)
		return
	}

	SuccessResponse(c, "Overdue bills retrieved", bills)
}

// GetUnpaidBills gets all unpaid bills (pending and overdue)
func (h *BillingHandler) GetUnpaidBills(c *gin.Context) {
	bills, err := h.billingService.GetUnpaidBills()
	if err != nil {
		InternalServerError(c, "Failed to fetch unpaid bills", err)
		return
	}

	SuccessResponse(c, "Unpaid bills retrieved", bills)
}

// GetBillingSummary gets billing summary for a period
func (h *BillingHandler) GetBillingSummary(c *gin.Context) {
	startDateStr := c.Query("start")
	endDateStr := c.Query("end")

	var startDate, endDate time.Time
	var err error

	// Default to current month if no dates provided
	if startDateStr == "" {
		now := time.Now()
		startDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	} else {
		startDate, err = utils.ParseDateString(startDateStr)
		if err != nil {
			BadRequest(c, "Invalid start date format. Use YYYY-MM-DD", err)
			return
		}
	}

	if endDateStr == "" {
		now := time.Now()
		endDate = time.Date(now.Year(), now.Month()+1, 0, 23, 59, 59, 0, now.Location())
	} else {
		endDate, err = utils.ParseDateString(endDateStr)
		if err != nil {
			BadRequest(c, "Invalid end date format. Use YYYY-MM-DD", err)
			return
		}
	}

	// Ensure start date is before end date
	if startDate.After(endDate) {
		BadRequest(c, "Start date must be before end date", nil)
		return
	}

	summary, err := h.billingService.GetBillingSummary(startDate, endDate)
	if err != nil {
		InternalServerError(c, "Failed to get billing summary", err)
		return
	}

	SuccessResponse(c, "Billing summary retrieved", summary)
}

// BulkSubmitReadings submits multiple meter readings
func (h *BillingHandler) BulkSubmitReadings(c *gin.Context) {
	var readings []MeterReadingRequest

	if err := c.ShouldBindJSON(&readings); err != nil {
		BadRequest(c, "Invalid meter reading data", err)
		return
	}

	if len(readings) == 0 {
		BadRequest(c, "No readings provided", nil)
		return
	}

	if len(readings) > 100 {
		BadRequest(c, "Maximum 100 readings per batch", nil)
		return
	}

	var results []BulkReadingResult
	var errors []BulkReadingError

	for i, req := range readings {
		// Validate required fields
		if req.MeterNumber == "" {
			errors = append(errors, BulkReadingError{
				Index: i,
				Meter: req.MeterNumber,
				Error: "Meter number is required",
			})
			continue
		}

		if req.CurrentReading <= 0 {
			errors = append(errors, BulkReadingError{
				Index: i,
				Meter: req.MeterNumber,
				Error: "Current reading must be greater than 0",
			})
			continue
		}

		// Set default values
		if req.ReadingDate.IsZero() {
			req.ReadingDate = time.Now()
		}

		reading := &models.MeterReading{
			MeterNumber:    req.MeterNumber,
			CurrentReading: req.CurrentReading,
			ReadingDate:    req.ReadingDate,
			ReadingType:    req.ReadingType,
			ReadingMethod:  req.ReadingMethod,
			ReaderName:     req.ReaderName,
			Notes:          req.Notes,
		}

		bill, err := h.billingService.SubmitMeterReading(reading)
		if err != nil {
			errors = append(errors, BulkReadingError{
				Index: i,
				Meter: req.MeterNumber,
				Error: err.Error(),
			})
		} else {
			results = append(results, BulkReadingResult{
				Meter:      req.MeterNumber,
				BillNumber: bill.BillNumber,
				Amount:     bill.TotalAmount,
			})
		}
	}

	response := gin.H{
		"success": len(results),
		"failed":  len(errors),
		"results": results,
		"errors":  errors,
	}

	if len(errors) > 0 && len(results) == 0 {
		ErrorResponse(c, http.StatusBadRequest, "All readings failed to process", nil)
		return
	}

	CreatedResponse(c, "Bulk readings processed", response)
}

// Request/Response DTOs

type MeterReadingRequest struct {
	MeterNumber    string             `json:"meter_number" binding:"required"`
	CurrentReading float64            `json:"current_reading" binding:"required"`
	ReadingDate    time.Time          `json:"reading_date"`
	ReadingType    string             `json:"reading_type"`   // "manual", "estimated", "actual"
	ReadingMethod  string             `json:"reading_method"` // "mobile_app", "field_agent", "customer"
	ReaderID       primitive.ObjectID `json:"reader_id,omitempty"`
	ReaderName     string             `json:"reader_name,omitempty"`
	Location       models.GeoLocation `json:"location,omitempty"`
	MeterPhotoURL  string             `json:"meter_photo_url,omitempty"`
	MeterCondition string             `json:"meter_condition,omitempty"`
	Notes          string             `json:"notes,omitempty"`
}

type PaymentRequest struct {
	Amount        float64 `json:"amount" binding:"required"`
	PaymentMethod string  `json:"payment_method" binding:"required"`
	TransactionID string  `json:"transaction_id"`
	ReceiptNumber string  `json:"receipt_number"`
	PayerName     string  `json:"payer_name,omitempty"`
	PayerPhone    string  `json:"payer_phone,omitempty"`
	CollectedBy   string  `json:"collected_by" binding:"required"`
	Notes         string  `json:"notes,omitempty"`
}

type BulkReadingResult struct {
	Meter      string  `json:"meter"`
	BillNumber string  `json:"bill_number"`
	Amount     float64 `json:"amount"`
}

type BulkReadingError struct {
	Index int    `json:"index"`
	Meter string `json:"meter"`
	Error string `json:"error"`
}
