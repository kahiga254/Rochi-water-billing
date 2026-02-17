package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"

	"waterbilling/backend/database"
	"waterbilling/backend/handlers"
	"waterbilling/backend/middleware"
	"waterbilling/backend/services"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Connect to MongoDB
	if err := database.Connect(); err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}
	defer database.Disconnect()

	// Initialize collections
	collections := initializeCollections()

	// Initialize services
	services := initializeServices(collections)

	// Initialize handlers
	handlers := initializeHandlers(services)

	// Initialize Gin router with middleware
	router := setupRouter(handlers, services.JWT)

	// Start server
	startServer(router)
}

// Collections holds all MongoDB collections
type Collections struct {
	Customers *mongo.Collection
	Readings  *mongo.Collection
	Bills     *mongo.Collection
	Payments  *mongo.Collection
	Users     *mongo.Collection
	SMSLogs   *mongo.Collection
	Tariffs   *mongo.Collection
	Templates *mongo.Collection
}

func initializeCollections() *Collections {
	db := database.DB

	return &Collections{
		Customers: db.Collection("customers"),
		Readings:  db.Collection("meter_readings"),
		Bills:     db.Collection("bills"),
		Payments:  db.Collection("payments"),
		Users:     db.Collection("users"),
		SMSLogs:   db.Collection("sms_logs"),
		Tariffs:   db.Collection("tariffs"),
		Templates: db.Collection("notification_templates"),
	}
}

// Services holds all business logic services
type Services struct {
	Customer *services.CustomerService
	Billing  *services.BillingService
	User     *services.UserService
	JWT      *services.JWTService
	SMS      *services.SMSService
	Payment  *services.PaymentService
}

func initializeServices(collections *Collections) *Services {
	// JWT Service
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "your-secret-key-change-in-production"
		log.Println("WARNING: Using default JWT secret. Set JWT_SECRET in .env for production!")
	}

	tokenDuration := 24 * time.Hour // Tokens valid for 24 hours
	jwtService := services.NewJWTService(jwtSecret, tokenDuration)

	// Customer Service
	customerService := services.NewCustomerService(collections.Customers, collections.Tariffs)

	// SMS Service - Initialize FIRST so it can be passed to other services
	smsService, err := services.NewSMSService(database.DB)
	if err != nil {
		log.Printf("Warning: SMS service initialization failed: %v", err)
		log.Println("SMS functionality will be disabled. Set TWILIO credentials in .env to enable.")
	}

	// Billing Service - NOW WITH SMS SERVICE INCLUDED
	billingService := services.NewBillingService(
		collections.Customers,
		collections.Readings,
		collections.Bills,
		collections.Payments,
		collections.Tariffs,
		smsService,
	)

	// User Service
	userService := services.NewUserService(collections.Users)
	paymentService := services.NewPaymentService(collections.Payments)

	return &Services{
		Customer: customerService,
		Billing:  billingService,
		User:     userService,
		JWT:      jwtService,
		SMS:      smsService,
		Payment:  paymentService,
	}
}

// Handlers holds all HTTP handlers
type Handlers struct {
	Customer  *handlers.CustomerHandler
	Billing   *handlers.BillingHandler
	SMS       *handlers.SMSHandler
	Dashboard *handlers.DashboardHandler
	Auth      *handlers.AuthHandler
	Payment   *handlers.PaymentHandler
}

func initializeHandlers(svc *Services) *Handlers {
	return &Handlers{
		Customer: handlers.NewCustomerHandler(svc.Customer),
		// ✅ Updated: Pass both Billing and User services to BillingHandler
		Billing:   handlers.NewBillingHandler(svc.Billing, svc.User),
		SMS:       handlers.NewSMSHandler(svc.Billing, svc.SMS),
		Dashboard: handlers.NewDashboardHandler(svc.Billing, svc.Customer),
		Auth:      handlers.NewAuthHandler(svc.User, svc.JWT),
		Payment:   handlers.NewPaymentHandler(svc.Payment, svc.Billing),
	}
}

func setupRouter(h *Handlers, jwtService *services.JWTService) *gin.Engine {
	// Set Gin mode
	if os.Getenv("ENV") == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	router := gin.New()

	// Global middleware
	router.Use(middleware.CORSMiddleware())
	router.Use(middleware.LoggingMiddleware())
	router.Use(gin.Recovery()) // Recovery from panics

	// API Routes
	api := router.Group("/api/v1")
	{
		// Public routes (no authentication required)
		public := api.Group("/auth")
		{
			public.POST("/login", h.Auth.Login)
			public.POST("/refresh-token", h.Auth.RefreshToken)
			public.POST("/register", h.Auth.Register)
			public.POST("/setup-admin", setupInitialAdmin)
		}

		// Protected routes (require authentication)
		protected := api.Group("")
		protected.Use(middleware.AuthMiddleware(jwtService))
		{
			// Customer routes
			customers := protected.Group("/customers")
			{
				customers.GET("", middleware.RoleMiddleware("admin", "manager"), h.Customer.GetCustomers)
				customers.POST("", middleware.RoleMiddleware("admin", "manager"), h.Customer.CreateCustomer)
				customers.GET("/meter/:meterNumber", h.Customer.GetCustomerByMeterNumber)
				customers.GET("/search", h.Customer.SearchCustomers)
				customers.GET("/zone/:zone", h.Customer.GetCustomersByZone)
				customers.PUT("/meter/:meterNumber", middleware.RoleMiddleware("admin", "manager", "customer_service"), h.Customer.UpdateCustomer)
				customers.PUT("/meter/:meterNumber/status", middleware.RoleMiddleware("admin", "manager"), h.Customer.UpdateCustomerStatus)
				customers.GET("/statistics", middleware.RoleMiddleware("admin", "manager"), h.Customer.GetCustomerStatistics)
				customers.POST("/bulk", middleware.RoleMiddleware("admin"), h.Customer.BulkCreateCustomers)
			}

			// Billing routes
			billing := protected.Group("/billing")
			{
				// Meter readings
				billing.POST("/readings", middleware.RoleMiddleware("admin", "reader", "manager"), h.Billing.SubmitMeterReading)
				billing.POST("/readings/bulk", middleware.RoleMiddleware("admin", "reader", "manager"), h.Billing.BulkSubmitReadings)

				// Customer billing info
				billing.GET("/customers/:meterNumber/bills", h.Billing.GetCustomerBills)
				billing.GET("/customers/:meterNumber/readings", h.Billing.GetCustomerReadingHistory)

				// Bill management
				billing.GET("/bills/overdue", middleware.RoleMiddleware("admin", "manager", "cashier"), h.Billing.GetOverdueBills)
				billing.GET("/bills/unpaid", middleware.RoleMiddleware("admin", "manager", "cashier"), h.Billing.GetUnpaidBills)
				billing.POST("/bills/:billID/pay", middleware.RoleMiddleware("admin", "cashier"), h.Billing.ProcessPayment)
				// ✅ Added my-readings endpoint
				billing.GET("/readings/my-readings", middleware.RoleMiddleware("reader"), h.Billing.GetMyReadings)

				// Summary and reports
				billing.GET("/summary", middleware.RoleMiddleware("admin", "manager"), h.Billing.GetBillingSummary)
			}

			// Payment routes
			payments := protected.Group("/payments")
			{
				payments.GET("", middleware.RoleMiddleware("admin", "customer_service"), h.Payment.GetPaymentsByMeter)
				payments.POST("", middleware.RoleMiddleware("admin", "cashier"), h.Payment.RecordPayment)
			}

			// SMS routes
			sms := protected.Group("/sms")
			sms.Use(middleware.RoleMiddleware("admin", "manager"))
			{
				sms.POST("/bills/:billID/notify", h.SMS.SendBillNotification)
				sms.POST("/bills/bulk-notify", h.SMS.BulkSendBillNotifications)
				sms.POST("/payments/confirm", h.SMS.SendPaymentConfirmation)
				sms.POST("/disconnection-warnings", h.SMS.SendDisconnectionWarning)
				sms.GET("/logs", h.SMS.GetSMSLogs)
			}

			// Dashboard routes
			dashboard := protected.Group("/dashboard")
			{
				dashboard.GET("/stats", h.Dashboard.GetDashboardStats)
				dashboard.GET("/reports/:year/:month", middleware.RoleMiddleware("admin", "manager"), h.Dashboard.GetMonthlyReport)
				dashboard.GET("/zones/performance", middleware.RoleMiddleware("admin", "manager"), h.Dashboard.GetZonePerformance)
				dashboard.GET("/readers/performance", middleware.RoleMiddleware("admin", "manager"), h.Dashboard.GetReaderPerformance)
			}

			// User management routes
			users := protected.Group("/users")
			users.Use(middleware.RoleMiddleware("admin"))
			{
				users.POST("", h.Auth.Register)
				users.GET("", h.Auth.GetUsers)
			}

			// Profile routes (authenticated users)
			profile := protected.Group("/profile")
			{
				profile.GET("", h.Auth.GetProfile)
				profile.PUT("", h.Auth.UpdateProfile)
				profile.POST("/change-password", h.Auth.ChangePassword)
				profile.POST("/logout", h.Auth.Logout)
			}
		}

		// Webhook routes (public but with secret validation)
		webhooks := api.Group("/webhooks")
		{
			webhooks.POST("/sms-delivery", handleSMSDeliveryWebhook)
			webhooks.POST("/mpesa-callback", handleMpesaWebhook)
		}
	}

	// Health check and info endpoints (public)
	router.GET("/health", healthCheck)
	router.GET("/", rootHandler)
	router.GET("/info", systemInfo)

	return router
}

func startServer(router *gin.Engine) {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	host := os.Getenv("HOST")
	if host == "" {
		host = "0.0.0.0"
	}

	address := host + ":" + port

	log.Printf("🚀 Water Billing System API starting on %s", address)
	log.Printf("📚 API Documentation available at http://%s/api/v1/docs", address)
	log.Printf("🔧 Environment: %s", os.Getenv("ENV"))

	if err := router.Run(address); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

// Health check endpoint
func healthCheck(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := database.Client.Ping(ctx, nil)
	dbStatus := "connected"
	if err != nil {
		dbStatus = "disconnected"
		log.Printf("Database health check failed: %v", err)
	}

	c.JSON(200, gin.H{
		"status":    "ok",
		"service":   "water-billing-api",
		"version":   "1.0.0",
		"database":  dbStatus,
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

// Root handler
func rootHandler(c *gin.Context) {
	c.JSON(200, gin.H{
		"service": "Water Billing System API",
		"version": "1.0.0",
		"endpoints": map[string]string{
			"api":    "/api/v1",
			"docs":   "/api/v1/docs",
			"health": "/health",
			"info":   "/info",
		},
		"description": "API for water company billing system with customer management, meter readings, billing, and SMS notifications",
	})
}

// System info endpoint
func systemInfo(c *gin.Context) {
	c.JSON(200, gin.H{
		"service":     "Water Billing System",
		"version":     "1.0.0",
		"environment": os.Getenv("ENV"),
		"go_version":  "1.21+",
		"features": []string{
			"Customer Management",
			"Meter Reading Submission",
			"Automated Billing",
			"Payment Processing",
			"SMS Notifications",
			"Dashboard & Reports",
			"Role-based Access Control",
		},
		"database":     "MongoDB",
		"sms_provider": getSMSProviderInfo(),
		"uptime":       time.Since(startTime).String(),
	})
}

// SMS delivery webhook handler
func handleSMSDeliveryWebhook(c *gin.Context) {
	var payload struct {
		MessageID string `json:"message_id"`
		Status    string `json:"status"`
		Timestamp string `json:"timestamp"`
		Error     string `json:"error,omitempty"`
	}

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(400, gin.H{"error": "Invalid payload"})
		return
	}

	secret := c.GetHeader("X-Webhook-Secret")
	expectedSecret := os.Getenv("WEBHOOK_SECRET")

	if expectedSecret != "" && secret != expectedSecret {
		c.JSON(401, gin.H{"error": "Invalid webhook secret"})
		return
	}

	c.JSON(200, gin.H{"status": "processed"})
}

// M-Pesa webhook handler
func handleMpesaWebhook(c *gin.Context) {
	c.JSON(200, gin.H{"status": "received"})
}

// Helper function to get SMS provider info
func getSMSProviderInfo() string {
	if os.Getenv("TWILIO_ACCOUNT_SID") != "" {
		return "Twilio"
	}
	if os.Getenv("AFRICASTALKING_API_KEY") != "" {
		return "Africa's Talking"
	}
	return "Not configured"
}

// Add this function to main.go
func setupInitialAdmin(c *gin.Context) {
	var req struct {
		Username  string `json:"username" binding:"required"`
		Email     string `json:"email" binding:"required,email"`
		Password  string `json:"password" binding:"required,min=6"`
		FirstName string `json:"first_name" binding:"required"`
		LastName  string `json:"last_name" binding:"required"`
		Phone     string `json:"phone" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"success": false,
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}

	collections := initializeCollections()

	count, err := collections.Users.CountDocuments(c.Request.Context(), gin.H{})
	if err != nil {
		c.JSON(500, gin.H{
			"success": false,
			"error":   "database_error",
			"message": "Failed to check existing users",
		})
		return
	}

	if count > 0 {
		c.JSON(403, gin.H{
			"success": false,
			"error":   "setup_complete",
			"message": "System already has users. Please contact an administrator.",
		})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(500, gin.H{
			"success": false,
			"error":   "password_hash_failed",
			"message": "Failed to hash password",
		})
		return
	}

	now := time.Now()
	user := bson.M{
		"_id":           primitive.NewObjectID(),
		"first_name":    req.FirstName,
		"last_name":     req.LastName,
		"email":         req.Email,
		"phone_number":  req.Phone,
		"username":      req.Username,
		"password":      string(hashedPassword),
		"role":          "admin",
		"department":    "Administration",
		"employee_id":   "ADMIN001",
		"assigned_zone": nil,
		"permissions":   []string{"*"},
		"is_active":     true,
		"last_login":    nil,
		"created_at":    now,
		"updated_at":    now,
	}

	result, err := collections.Users.InsertOne(c.Request.Context(), user)
	if err != nil {
		c.JSON(500, gin.H{
			"success": false,
			"error":   "creation_failed",
			"message": "Failed to create user: " + err.Error(),
		})
		return
	}

	c.JSON(201, gin.H{
		"success": true,
		"message": "Initial admin user created successfully",
		"data": gin.H{
			"id":           result.InsertedID.(primitive.ObjectID).Hex(),
			"username":     req.Username,
			"email":        req.Email,
			"first_name":   req.FirstName,
			"last_name":    req.LastName,
			"phone_number": req.Phone,
			"role":         "admin",
			"is_active":    true,
		},
	})
}

var startTime = time.Now()
