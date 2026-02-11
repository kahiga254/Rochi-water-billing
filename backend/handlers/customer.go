package handlers

import (
	"fmt"
	"net/http"

	"waterbilling/backend/models"
	"waterbilling/backend/services"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CustomerHandler struct {
	customerService *services.CustomerService
}

func NewCustomerHandler(customerService *services.CustomerService) *CustomerHandler {
	return &CustomerHandler{
		customerService: customerService,
	}
}

// CreateCustomer handles customer creation
// @Summary Create a new customer
// @Description Create a new water billing customer
// @Tags Customers
// @Accept json
// @Produce json
// @Param customer body models.Customer true "Customer data"
// @Success 201 {object} Response "Customer created successfully"
// @Failure 400 {object} Response "Invalid input"
// @Failure 409 {object} Response "Customer already exists"
// @Failure 500 {object} Response "Internal server error"
// @Router /customers [post]
func (h *CustomerHandler) CreateCustomer(c *gin.Context) {
	var customer models.Customer

	if err := c.ShouldBindJSON(&customer); err != nil {
		BadRequest(c, "Invalid customer data", err)
		return
	}

	// Validate required fields
	if customer.MeterNumber == "" {
		BadRequest(c, "Meter number is required", nil)
		return
	}

	if customer.FirstName == "" || customer.LastName == "" {
		BadRequest(c, "First name and last name are required", nil)
		return
	}

	if customer.PhoneNumber == "" {
		BadRequest(c, "Phone number is required", nil)
		return
	}

	if customer.Address.StreetAddress == "" || customer.Address.City == "" {
		BadRequest(c, "Address is required", nil)
		return
	}

	if customer.Zone == "" {
		BadRequest(c, "Zone is required", nil)
		return
	}

	// Create customer
	if err := h.customerService.CreateCustomer(&customer); err != nil {
		if err.Error() == "customer with meter number "+customer.MeterNumber+" already exists" {
			ErrorResponse(c, http.StatusConflict, "Customer already exists", err)
		} else {
			InternalServerError(c, "Failed to create customer", err)
		}
		return
	}

	CreatedResponse(c, "Customer created successfully", customer)
}

// GetCustomerByMeterNumber retrieves a customer by meter number
// @Summary Get customer by meter number
// @Description Get customer details using meter number
// @Tags Customers
// @Accept json
// @Produce json
// @Param meterNumber path string true "Meter Number"
// @Success 200 {object} Response "Customer found"
// @Failure 404 {object} Response "Customer not found"
// @Failure 500 {object} Response "Internal server error"
// @Router /customers/meter/{meterNumber} [get]
func (h *CustomerHandler) GetCustomerByMeterNumber(c *gin.Context) {
	meterNumber := c.Param("meterNumber")
	if meterNumber == "" {
		BadRequest(c, "Meter number is required", nil)
		return
	}

	customer, err := h.customerService.GetCustomerByMeterNumber(meterNumber)
	if err != nil {
		InternalServerError(c, "Failed to fetch customer", err)
		return
	}

	if customer == nil {
		NotFound(c, "Customer not found")
		return
	}

	SuccessResponse(c, "Customer found", customer)
}

// GetCustomerByID retrieves a customer by ID
// @Summary Get customer by ID
// @Description Get customer details using customer ID
// @Tags Customers
// @Accept json
// @Produce json
// @Param id path string true "Customer ID"
// @Success 200 {object} Response "Customer found"
// @Failure 404 {object} Response "Customer not found"
// @Failure 500 {object} Response "Internal server error"
// @Router /customers/{id} [get]
func (h *CustomerHandler) GetCustomerByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		BadRequest(c, "Customer ID is required", nil)
		return
	}

	_, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		BadRequest(c, "Invalid customer ID", err)
		return
	}

	BadRequest(c, "Please use meter number endpoint: /api/v1/customers/meter/{meterNumber}", nil)
}

// UpdateCustomer updates customer information
// @Summary Update customer
// @Description Update customer information
// @Tags Customers
// @Accept json
// @Produce json
// @Param meterNumber path string true "Meter Number"
// @Param updates body map[string]interface{} true "Customer updates"
// @Success 200 {object} Response "Customer updated successfully"
// @Failure 400 {object} Response "Invalid input"
// @Failure 404 {object} Response "Customer not found"
// @Failure 500 {object} Response "Internal server error"
// @Router /customers/meter/{meterNumber} [put]
func (h *CustomerHandler) UpdateCustomer(c *gin.Context) {
	meterNumber := c.Param("meterNumber")
	if meterNumber == "" {
		BadRequest(c, "Meter number is required", nil)
		return
	}

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		BadRequest(c, "Invalid update data", err)
		return
	}

	if err := h.customerService.UpdateCustomer(meterNumber, updates); err != nil {
		if err.Error() == "customer with meter number "+meterNumber+" not found" {
			NotFound(c, "Customer not found")
		} else {
			InternalServerError(c, "Failed to update customer", err)
		}
		return
	}

	SuccessResponse(c, "Customer updated successfully", nil)
}

// SearchCustomers searches for customers
// @Summary Search customers
// @Description Search customers by various criteria
// @Tags Customers
// @Accept json
// @Produce json
// @Param search query string false "Search term"
// @Param zone query string false "Zone"
// @Param status query string false "Status"
// @Param customerType query string false "Customer Type"
// @Param limit query int false "Limit results" default(50)
// @Success 200 {object} Response "Customers found"
// @Failure 500 {object} Response "Internal server error"
// @Router /customers/search [get]
func (h *CustomerHandler) SearchCustomers(c *gin.Context) {
	searchTerm := c.Query("search")
	zone := c.Query("zone")
	status := c.Query("status")
	customerType := c.Query("customerType")
	limit := c.DefaultQuery("limit", "50")

	var limitInt int64 = 50
	if limit != "" {
		if l, err := parseInt64(limit); err == nil && l > 0 {
			limitInt = l
		}
	}

	customers, err := h.customerService.SearchCustomers(searchTerm, zone, status, customerType, limitInt)
	if err != nil {
		InternalServerError(c, "Failed to search customers", err)
		return
	}

	SuccessResponse(c, "Customers found", customers)
}

// GetCustomersByZone gets customers in a zone
// @Summary Get customers by zone
// @Description Get all customers in a specific zone
// @Tags Customers
// @Accept json
// @Produce json
// @Param zone path string true "Zone"
// @Success 200 {object} Response "Customers found"
// @Failure 500 {object} Response "Internal server error"
// @Router /customers/zone/{zone} [get]
func (h *CustomerHandler) GetCustomersByZone(c *gin.Context) {
	zone := c.Param("zone")
	if zone == "" {
		BadRequest(c, "Zone is required", nil)
		return
	}

	customers, err := h.customerService.GetCustomersByZone(zone)
	if err != nil {
		InternalServerError(c, "Failed to fetch customers by zone", err)
		return
	}

	SuccessResponse(c, "Customers found", customers)
}

// UpdateCustomerStatus updates customer status
// @Summary Update customer status
// @Description Update customer status (active, inactive, disconnected, etc.)
// @Tags Customers
// @Accept json
// @Produce json
// @Param meterNumber path string true "Meter Number"
// @Param request body UpdateStatusRequest true "Status update"
// @Success 200 {object} Response "Status updated successfully"
// @Failure 400 {object} Response "Invalid input"
// @Failure 404 {object} Response "Customer not found"
// @Failure 500 {object} Response "Internal server error"
// @Router /customers/meter/{meterNumber}/status [put]
func (h *CustomerHandler) UpdateCustomerStatus(c *gin.Context) {
	meterNumber := c.Param("meterNumber")
	if meterNumber == "" {
		BadRequest(c, "Meter number is required", nil)
		return
	}

	var req UpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "Invalid request data", err)
		return
	}

	if req.Status == "" {
		BadRequest(c, "Status is required", nil)
		return
	}

	// Validate status
	validStatuses := map[string]bool{
		"active":       true,
		"inactive":     true,
		"disconnected": true,
		"pending":      true,
		"suspended":    true,
	}

	if !validStatuses[req.Status] {
		BadRequest(c, "Invalid status value", nil)
		return
	}

	if err := h.customerService.UpdateCustomerStatus(meterNumber, req.Status, req.Reason); err != nil {
		if err.Error() == "customer with meter number "+meterNumber+" not found" {
			NotFound(c, "Customer not found")
		} else {
			InternalServerError(c, "Failed to update customer status", err)
		}
		return
	}

	SuccessResponse(c, "Customer status updated successfully", nil)
}

// GetCustomerStatistics gets customer statistics
// @Summary Get customer statistics
// @Description Get statistics about customers
// @Tags Customers
// @Accept json
// @Produce json
// @Success 200 {object} Response "Statistics retrieved"
// @Failure 500 {object} Response "Internal server error"
// @Router /customers/statistics [get]
func (h *CustomerHandler) GetCustomerStatistics(c *gin.Context) {
	stats, err := h.customerService.GetCustomerStatistics()
	if err != nil {
		InternalServerError(c, "Failed to get customer statistics", err)
		return
	}

	SuccessResponse(c, "Customer statistics", stats)
}

// BulkCreateCustomers creates multiple customers from CSV/Google Sheets
// @Summary Bulk create customers
// @Description Create multiple customers from CSV data
// @Tags Customers
// @Accept json
// @Produce json
// @Param customers body []models.Customer true "Array of customers"
// @Success 201 {object} Response "Customers created successfully"
// @Failure 400 {object} Response "Invalid input"
// @Failure 500 {object} Response "Internal server error"
// @Router /customers/bulk [post]
func (h *CustomerHandler) BulkCreateCustomers(c *gin.Context) {
	var customers []models.Customer

	if err := c.ShouldBindJSON(&customers); err != nil {
		BadRequest(c, "Invalid customer data", err)
		return
	}

	if len(customers) == 0 {
		BadRequest(c, "No customers provided", nil)
		return
	}

	if len(customers) > 1000 {
		BadRequest(c, "Maximum 1000 customers per batch", nil)
		return
	}

	var results []BulkCreateResult
	var errors []BulkCreateError

	for i, customer := range customers {
		if err := h.customerService.CreateCustomer(&customer); err != nil {
			errors = append(errors, BulkCreateError{
				Index: i,
				Meter: customer.MeterNumber,
				Error: err.Error(),
			})
		} else {
			results = append(results, BulkCreateResult{
				Meter: customer.MeterNumber,
				Name:  customer.FullName(),
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
		ErrorResponse(c, http.StatusBadRequest, "All customers failed to create", nil)
		return
	}

	CreatedResponse(c, "Bulk create completed", response)
}

// UpdateStatusRequest represents status update request
type UpdateStatusRequest struct {
	Status string `json:"status" binding:"required"`
	Reason string `json:"reason,omitempty"`
}

// BulkCreateResult represents successful bulk create
type BulkCreateResult struct {
	Meter string `json:"meter"`
	Name  string `json:"name"`
}

// BulkCreateError represents failed bulk create
type BulkCreateError struct {
	Index int    `json:"index"`
	Meter string `json:"meter"`
	Error string `json:"error"`
}

// Helper function to parse string to int64
func parseInt64(s string) (int64, error) {
	var n int64
	_, err := fmt.Sscanf(s, "%d", &n)
	return n, err
}
