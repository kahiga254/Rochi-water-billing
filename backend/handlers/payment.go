package handlers

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"waterbilling/backend/models"
	"waterbilling/backend/services"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PaymentHandler struct {
	paymentService *services.PaymentService
	billingService *services.BillingService
}

func NewPaymentHandler(paymentService *services.PaymentService, billingService *services.BillingService) *PaymentHandler {
	return &PaymentHandler{
		paymentService: paymentService,
		billingService: billingService,
	}
}

// RecordPayment handles payment recording
func (h *PaymentHandler) RecordPayment(c *gin.Context) {
	var req struct {
		BillID        string  `json:"bill_id" binding:"required"`
		MeterNumber   string  `json:"meter_number" binding:"required"`
		CustomerID    string  `json:"customer_id" binding:"required"`
		CustomerName  string  `json:"customer_name" binding:"required"`
		Amount        float64 `json:"amount" binding:"required,gt=0"`
		PaymentMethod string  `json:"payment_method" binding:"required"`
		TransactionID string  `json:"transaction_id"`
		PaymentDate   string  `json:"payment_date" binding:"required"`
		CollectedBy   string  `json:"collected_by" binding:"required"`
		Notes         string  `json:"notes"`
		Status        string  `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "Invalid payment data", err)
		return
	}

	// Parse bill ID
	billObjectID, err := primitive.ObjectIDFromHex(req.BillID)
	if err != nil {
		BadRequest(c, "Invalid bill ID", err)
		return
	}

	// Parse customer ID
	customerObjectID, err := primitive.ObjectIDFromHex(req.CustomerID)
	if err != nil {
		BadRequest(c, "Invalid customer ID", err)
		return
	}

	// Parse payment date
	paymentDate, err := time.Parse(time.RFC3339, req.PaymentDate)
	if err != nil {
		// Try parsing as date only
		paymentDate, err = time.Parse("2006-01-02", req.PaymentDate)
		if err != nil {
			BadRequest(c, "Invalid payment date", err)
			return
		}
	}

	// Create payment record
	payment := &models.Payment{
		ID:            primitive.NewObjectID(),
		BillID:        billObjectID,
		MeterNumber:   req.MeterNumber,
		CustomerID:    customerObjectID,
		CustomerName:  req.CustomerName,
		Amount:        req.Amount,
		PaymentMethod: req.PaymentMethod,
		TransactionID: req.TransactionID,
		PaymentDate:   paymentDate,
		CollectedBy:   req.CollectedBy,
		Notes:         req.Notes,
		Status:        req.Status,
		CreatedAt:     time.Now(),
	}

	// Save payment
	if err := h.paymentService.CreatePayment(payment); err != nil {
		InternalServerError(c, "Failed to save payment", err)
		return
	}

	// Update bill payment status and customer balance
	if err := h.billingService.UpdateBillPayment(req.BillID, req.Amount); err != nil {
		// Log error but don't fail the request
		fmt.Printf("Failed to update bill payment: %v\n", err)
	}

	// Generate receipt number
	receiptNumber := generateReceiptNumber()

	SuccessResponse(c, "Payment recorded successfully", gin.H{
		"id":             payment.ID.Hex(),
		"receipt_number": receiptNumber,
		"amount":         payment.Amount,
		"payment_date":   payment.PaymentDate,
		"status":         payment.Status,
	})
}

// GetPaymentsByMeter returns payment history for a specific meter
func (h *PaymentHandler) GetPaymentsByMeter(c *gin.Context) {
	meterNumber := c.Query("meter_number")
	if meterNumber == "" {
		BadRequest(c, "meter_number is required", nil)
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	payments, err := h.paymentService.GetPaymentsByMeter(meterNumber, limit)
	if err != nil {
		InternalServerError(c, "Failed to fetch payments", err)
		return
	}

	SuccessResponse(c, "Payments retrieved", payments)
}

// Helper function to generate receipt number
func generateReceiptNumber() string {
	return "RCPT-" + time.Now().Format("20060102") + "-" + randomString(6)
}

// Fixed random string generator without sleep
func randomString(n int) string {
	const letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, n)

	// Create a new random source seeded with current time
	src := rand.NewSource(time.Now().UnixNano())
	r := rand.New(src)

	for i := range result {
		result[i] = letters[r.Intn(len(letters))]
	}

	return string(result)
}
