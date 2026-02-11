package handlers

import (
	"net/http" // Add this import
	"strconv"  // Add this import
	"time"

	"waterbilling/backend/models"   // Fixed import path
	"waterbilling/backend/services" // Fixed import path

	"github.com/gin-gonic/gin"
)

type DashboardHandler struct {
	billingService  *services.BillingService
	customerService *services.CustomerService
}

func NewDashboardHandler(billingService *services.BillingService, customerService *services.CustomerService) *DashboardHandler {
	return &DashboardHandler{
		billingService:  billingService,
		customerService: customerService,
	}
}

// Helper function for not implemented endpoints
func NotImplemented(c *gin.Context, message string) {
	ErrorResponse(c, http.StatusNotImplemented, message, nil)
}

// GetDashboardStats gets dashboard statistics
func (h *DashboardHandler) GetDashboardStats(c *gin.Context) {
	// Get current month dates
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	endOfMonth := time.Date(now.Year(), now.Month()+1, 0, 23, 59, 59, 0, now.Location())

	// Get billing summary for current month
	billingSummary, err := h.billingService.GetBillingSummary(startOfMonth, endOfMonth)
	if err != nil {
		InternalServerError(c, "Failed to get billing summary", err)
		return
	}

	// Get customer statistics
	customerStats, err := h.customerService.GetCustomerStatistics()
	if err != nil {
		InternalServerError(c, "Failed to get customer statistics", err)
		return
	}

	// Get overdue bills
	overdueBills, err := h.billingService.GetOverdueBills()
	if err != nil {
		InternalServerError(c, "Failed to get overdue bills", err)
		return
	}

	// Get unpaid bills
	unpaidBills, err := h.billingService.GetUnpaidBills()
	if err != nil {
		InternalServerError(c, "Failed to get unpaid bills", err)
		return
	}

	// Calculate revenue metrics
	var totalRevenue, pendingRevenue float64
	for status, summary := range billingSummary.StatusBreakdown {
		totalRevenue += summary.TotalPaid
		if status == "pending" || status == "overdue" {
			pendingRevenue += summary.TotalAmount - summary.TotalPaid
		}
	}

	dashboardStats := DashboardStats{
		CustomerStats:  customerStats,
		BillingSummary: billingSummary,
		Revenue: RevenueStats{
			Total:          totalRevenue,
			Pending:        pendingRevenue,
			CollectionRate: calculateCollectionRate(billingSummary),
		},
		OverdueBills: OverdueStats{
			Count:  len(overdueBills),
			Amount: calculateTotalAmount(overdueBills),
		},
		UnpaidBills: UnpaidStats{
			Count:  len(unpaidBills),
			Amount: calculateTotalAmount(unpaidBills),
		},
		Month: now.Format("January 2006"),
	}

	SuccessResponse(c, "Dashboard statistics retrieved", dashboardStats)
}

// GetMonthlyReport gets monthly billing report
func (h *DashboardHandler) GetMonthlyReport(c *gin.Context) {
	yearStr := c.Param("year")
	monthStr := c.Param("month")

	year, err := strconv.Atoi(yearStr)
	if err != nil || year < 2000 || year > 2100 {
		BadRequest(c, "Invalid year", nil)
		return
	}

	month, err := strconv.Atoi(monthStr)
	if err != nil || month < 1 || month > 12 {
		BadRequest(c, "Invalid month. Must be 1-12", nil)
		return
	}

	// Calculate date range for the month
	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local)
	endDate := startDate.AddDate(0, 1, 0).Add(-time.Second)

	// Get billing summary
	billingSummary, err := h.billingService.GetBillingSummary(startDate, endDate)
	if err != nil {
		InternalServerError(c, "Failed to get billing summary", err)
		return
	}

	monthlyReport := MonthlyReport{
		Year:           year,
		Month:          month,
		Period:         startDate.Format("January 2006"),
		BillingSummary: billingSummary,
		GeneratedAt:    time.Now(),
	}

	SuccessResponse(c, "Monthly report retrieved", monthlyReport)
}

// GetZonePerformance gets performance metrics by zone
func (h *DashboardHandler) GetZonePerformance(c *gin.Context) {
	notImplemented(c, "Zone performance metrics not yet implemented")
}

// GetReaderPerformance gets performance metrics for meter readers
func (h *DashboardHandler) GetReaderPerformance(c *gin.Context) {
	notImplemented(c, "Reader performance metrics not yet implemented")
}

// Helper functions

func calculateCollectionRate(summary *services.BillingSummary) float64 {
	var totalBilled, totalPaid float64
	for _, statusSummary := range summary.StatusBreakdown {
		totalBilled += statusSummary.TotalAmount
		totalPaid += statusSummary.TotalPaid
	}

	if totalBilled == 0 {
		return 0
	}
	return (totalPaid / totalBilled) * 100
}

func calculateTotalAmount(bills []models.Bill) float64 {
	var total float64
	for _, bill := range bills {
		total += bill.Balance
	}
	return total
}

// Dashboard DTOs

type DashboardStats struct {
	CustomerStats  *services.CustomerStatistics `json:"customer_stats"`
	BillingSummary *services.BillingSummary     `json:"billing_summary"`
	Revenue        RevenueStats                 `json:"revenue"`
	OverdueBills   OverdueStats                 `json:"overdue_bills"`
	UnpaidBills    UnpaidStats                  `json:"unpaid_bills"`
	Month          string                       `json:"month"`
}

type RevenueStats struct {
	Total          float64 `json:"total"`
	Pending        float64 `json:"pending"`
	CollectionRate float64 `json:"collection_rate"`
}

type OverdueStats struct {
	Count  int     `json:"count"`
	Amount float64 `json:"amount"`
}

type UnpaidStats struct {
	Count  int     `json:"count"`
	Amount float64 `json:"amount"`
}

type MonthlyReport struct {
	Year           int                      `json:"year"`
	Month          int                      `json:"month"`
	Period         string                   `json:"period"`
	BillingSummary *services.BillingSummary `json:"billing_summary"`
	GeneratedAt    time.Time                `json:"generated_at"`
}
