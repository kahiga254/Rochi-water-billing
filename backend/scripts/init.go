package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"waterbilling/backend/database"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	// Connect to database
	if err := database.Connect(); err != nil {
		log.Fatal(err)
	}
	defer database.Disconnect()

	fmt.Println("=== Database Initialization Started ===")

	// Create collections with validation
	createCollections()

	// Create indexes
	createIndexes()

	// Create admin user
	createAdminUser()

	// Create default tariff
	createDefaultTariff()

	// Create notification templates
	createNotificationTemplates()

	fmt.Println("=== Database Initialization Completed Successfully ===")
}

func createCollections() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Customer collection validation
	customerValidator := bson.M{
		"$jsonSchema": bson.M{
			"bsonType": "object",
			"required": []string{"meter_number", "first_name", "last_name", "phone_number", "address", "zone", "tariff_code"},
			"properties": bson.M{
				"meter_number": bson.M{
					"bsonType":    "string",
					"description": "must be a string and is required",
				},
				"account_number": bson.M{
					"bsonType": "string",
				},
				"first_name": bson.M{
					"bsonType":    "string",
					"description": "must be a string and is required",
				},
				"last_name": bson.M{
					"bsonType":    "string",
					"description": "must be a string and is required",
				},
				"phone_number": bson.M{
					"bsonType":    "string",
					"description": "must be a string and is required",
				},
				"status": bson.M{
					"bsonType": "string",
					"enum":     []string{"active", "inactive", "disconnected", "pending", "suspended"},
				},
			},
		},
	}

	// Create or update collection with validation
	db := database.DB
	collections, _ := db.ListCollectionNames(ctx, bson.M{})

	// Create customers collection if not exists
	if !contains(collections, "customers") {
		opts := options.CreateCollection().SetValidator(customerValidator)
		db.CreateCollection(ctx, "customers", opts)
		fmt.Println("✓ Created 'customers' collection with validation")
	}

	// Create other collections if they don't exist
	collectionsToCreate := []string{
		"meter_readings",
		"bills",
		"payments",
		"users",
		"sms_logs",
		"notification_templates",
		"tariffs",
	}

	for _, collName := range collectionsToCreate {
		if !contains(collections, collName) {
			db.CreateCollection(ctx, collName)
			fmt.Printf("✓ Created '%s' collection\n", collName)
		}
	}
}

func createIndexes() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	fmt.Println("\n=== Creating Database Indexes ===")

	// 1. CUSTOMERS COLLECTION INDEXES
	customerIndexes := []mongo.IndexModel{
		// Meter number as unique primary identifier
		{
			Keys:    bson.D{{Key: "meter_number", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("meter_number_unique"),
		},
		// Account number as unique alternative identifier
		{
			Keys:    bson.D{{Key: "account_number", Value: 1}},
			Options: options.Index().SetUnique(true).SetSparse(true).SetName("account_number_unique"),
		},
		// Phone number should be unique
		{
			Keys:    bson.D{{Key: "phone_number", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("phone_number_unique"),
		},
		// Zone index for geographic queries
		{
			Keys:    bson.D{{Key: "zone", Value: 1}},
			Options: options.Index().SetName("zone_index"),
		},
		// Status index for filtering active/inactive customers
		{
			Keys:    bson.D{{Key: "status", Value: 1}},
			Options: options.Index().SetName("status_index"),
		},
		// Combined index for customer search
		{
			Keys: bson.D{
				{Key: "first_name", Value: "text"},
				{Key: "last_name", Value: "text"},
				{Key: "meter_number", Value: "text"},
				{Key: "phone_number", Value: "text"},
			},
			Options: options.Index().SetName("customer_search_text").SetWeights(bson.D{
				{Key: "meter_number", Value: 10},
				{Key: "first_name", Value: 5},
				{Key: "last_name", Value: 5},
				{Key: "phone_number", Value: 3},
			}),
		},
		// Index for customer type and zone queries
		{
			Keys: bson.D{
				{Key: "customer_type", Value: 1},
				{Key: "zone", Value: 1},
			},
			Options: options.Index().SetName("customer_type_zone"),
		},
	}

	// 2. METER READINGS COLLECTION INDEXES
	readingIndexes := []mongo.IndexModel{
		// Primary index on meter number and reading date
		{
			Keys: bson.D{
				{Key: "meter_number", Value: 1},
				{Key: "reading_date", Value: -1},
			},
			Options: options.Index().SetName("meter_reading_date"),
		},
		// Index for billing period queries
		{
			Keys: bson.D{
				{Key: "meter_number", Value: 1},
				{Key: "month", Value: 1},
				{Key: "year", Value: 1},
			},
			Options: options.Index().SetUnique(true).SetName("meter_month_year_unique"),
		},
		// Index for reader assignments
		{
			Keys: bson.D{
				{Key: "reader_id", Value: 1},
				{Key: "reading_date", Value: -1},
			},
			Options: options.Index().SetName("reader_readings"),
		},
		// Zone-based reading queries
		{
			Keys: bson.D{
				{Key: "zone", Value: 1},
				{Key: "reading_date", Value: -1},
			},
			Options: options.Index().SetName("zone_readings"),
		},
		// Status index for workflow management
		{
			Keys:    bson.D{{Key: "status", Value: 1}},
			Options: options.Index().SetName("reading_status"),
		},
	}

	// 3. BILLS COLLECTION INDEXES
	billIndexes := []mongo.IndexModel{
		// Bill number as unique identifier
		{
			Keys:    bson.D{{Key: "bill_number", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("bill_number_unique"),
		},
		// Meter number and status for customer bill queries
		{
			Keys: bson.D{
				{Key: "meter_number", Value: 1},
				{Key: "status", Value: 1},
			},
			Options: options.Index().SetName("meter_bill_status"),
		},
		// Due date index for overdue bills
		{
			Keys:    bson.D{{Key: "due_date", Value: 1}},
			Options: options.Index().SetName("due_date_index"),
		},
		// Billing period index for reports
		{
			Keys:    bson.D{{Key: "billing_period", Value: 1}},
			Options: options.Index().SetName("billing_period_index"),
		},
		// Combined index for payment tracking
		{
			Keys: bson.D{
				{Key: "meter_number", Value: 1},
				{Key: "status", Value: 1},
				{Key: "due_date", Value: 1},
			},
			Options: options.Index().SetName("customer_payment_status"),
		},
		// Index for SMS notification tracking
		{
			Keys: bson.D{
				{Key: "sms_sent", Value: 1},
				{Key: "due_date", Value: 1},
			},
			Options: options.Index().SetName("sms_notification_tracking"),
		},
	}

	// 4. PAYMENTS COLLECTION INDEXES
	paymentIndexes := []mongo.IndexModel{
		// Transaction ID should be unique
		{
			Keys:    bson.D{{Key: "transaction_id", Value: 1}},
			Options: options.Index().SetUnique(true).SetSparse(true).SetName("transaction_id_unique"),
		},
		// Receipt number should be unique
		{
			Keys:    bson.D{{Key: "receipt_number", Value: 1}},
			Options: options.Index().SetUnique(true).SetSparse(true).SetName("receipt_number_unique"),
		},
		// Meter number and date for customer payment history
		{
			Keys: bson.D{
				{Key: "meter_number", Value: 1},
				{Key: "payment_date", Value: -1},
			},
			Options: options.Index().SetName("customer_payments"),
		},
		// Payment date for financial reporting
		{
			Keys:    bson.D{{Key: "payment_date", Value: -1}},
			Options: options.Index().SetName("payment_date_index"),
		},
		// Collected by index for cashier performance
		{
			Keys:    bson.D{{Key: "collected_by", Value: 1}},
			Options: options.Index().SetName("collected_by_index"),
		},
		// Payment method index for analysis
		{
			Keys:    bson.D{{Key: "payment_method", Value: 1}},
			Options: options.Index().SetName("payment_method_index"),
		},
	}

	// 5. USERS COLLECTION INDEXES
	userIndexes := []mongo.IndexModel{
		// Username and email should be unique
		{
			Keys:    bson.D{{Key: "username", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("username_unique"),
		},
		{
			Keys:    bson.D{{Key: "email", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("email_unique"),
		},
		// Employee ID should be unique
		{
			Keys:    bson.D{{Key: "employee_id", Value: 1}},
			Options: options.Index().SetUnique(true).SetSparse(true).SetName("employee_id_unique"),
		},
		// Role and zone for permission queries
		{
			Keys: bson.D{
				{Key: "role", Value: 1},
				{Key: "assigned_zone", Value: 1},
			},
			Options: options.Index().SetName("role_zone_index"),
		},
		// Active status index
		{
			Keys:    bson.D{{Key: "is_active", Value: 1}},
			Options: options.Index().SetName("user_active_status"),
		},
	}

	// 6. SMS LOGS COLLECTION INDEXES
	smsLogIndexes := []mongo.IndexModel{
		// Sent date for chronological queries
		{
			Keys:    bson.D{{Key: "sent_at", Value: -1}},
			Options: options.Index().SetName("sms_sent_date"),
		},
		// Meter number for customer SMS history
		{
			Keys:    bson.D{{Key: "meter_number", Value: 1}},
			Options: options.Index().SetName("customer_sms_history"),
		},
		// Message type for analytics
		{
			Keys:    bson.D{{Key: "message_type", Value: 1}},
			Options: options.Index().SetName("sms_message_type"),
		},
		// Status for delivery tracking
		{
			Keys:    bson.D{{Key: "status", Value: 1}},
			Options: options.Index().SetName("sms_status"),
		},
	}

	// 7. TARIFFS COLLECTION INDEXES
	tariffIndexes := []mongo.IndexModel{
		// Tariff code should be unique
		{
			Keys:    bson.D{{Key: "code", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("tariff_code_unique"),
		},
		// Effective date for current tariff queries
		{
			Keys:    bson.D{{Key: "effective_date", Value: -1}},
			Options: options.Index().SetName("tariff_effective_date"),
		},
		// Active status for filtering
		{
			Keys:    bson.D{{Key: "is_active", Value: 1}},
			Options: options.Index().SetName("tariff_active_status"),
		},
	}

	// Create all indexes
	collections := map[string][]mongo.IndexModel{
		"customers":      customerIndexes,
		"meter_readings": readingIndexes,
		"bills":          billIndexes,
		"payments":       paymentIndexes,
		"users":          userIndexes,
		"sms_logs":       smsLogIndexes,
		"tariffs":        tariffIndexes,
	}

	for collectionName, indexes := range collections {
		collection := database.DB.Collection(collectionName)
		_, err := collection.Indexes().CreateMany(ctx, indexes)
		if err != nil {
			log.Printf("Error creating indexes for %s: %v", collectionName, err)
		} else {
			fmt.Printf("✓ Created %d indexes for '%s' collection\n", len(indexes), collectionName)
		}
	}
}

func createAdminUser() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := database.DB.Collection("users")

	// Check if admin already exists
	count, err := collection.CountDocuments(ctx, bson.M{"username": "admin"})
	if err != nil {
		log.Printf("Error checking admin user: %v", err)
		return
	}

	if count > 0 {
		fmt.Println("✓ Admin user already exists")
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("Admin@2024"), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		return
	}

	adminUser := bson.M{
		"first_name":   "System",
		"last_name":    "Administrator",
		"email":        "admin@watercompany.com",
		"phone_number": "+254700000000",
		"username":     "admin",
		"password":     string(hashedPassword),
		"role":         "admin",
		"employee_id":  "EMP001",
		"department":   "Administration",
		"permissions":  []string{"all"},
		"is_active":    true,
		"created_at":   time.Now(),
		"updated_at":   time.Now(),
	}

	_, err = collection.InsertOne(ctx, adminUser)
	if err != nil {
		log.Printf("Error creating admin user: %v", err)
		return
	}

	fmt.Println("✓ Admin user created successfully")
	fmt.Println("  Username: admin")
	fmt.Println("  Password: Admin@2024")
	fmt.Println("  Please change the password after first login!")
}

func createDefaultTariff() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := database.DB.Collection("tariffs")

	// Check if default tariff exists
	count, err := collection.CountDocuments(ctx, bson.M{"code": "RES-BASIC"})
	if err != nil {
		log.Printf("Error checking default tariff: %v", err)
		return
	}

	if count > 0 {
		fmt.Println("✓ Default tariff already exists")
		return
	}

	defaultTariff := bson.M{
		"code":           "RES-BASIC",
		"name":           "Residential Basic",
		"customer_type":  "residential",
		"description":    "Default residential tariff for water billing",
		"base_rate":      100.0,
		"fixed_charge":   150.0,
		"effective_date": time.Now(),
		"is_active":      true,
		"created_at":     time.Now(),
		"updated_at":     time.Now(),
	}

	_, err = collection.InsertOne(ctx, defaultTariff)
	if err != nil {
		log.Printf("Error creating default tariff: %v", err)
		return
	}

	fmt.Println("✓ Default tariff created: RES-BASIC (Residential Basic)")
}

func createNotificationTemplates() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := database.DB.Collection("notification_templates")

	templates := []bson.M{
		{
			"template_type": "sms",
			"name":          "Bill Notification",
			"body":          "Dear {customer_name},\nYour water bill {bill_number} is ready.\nMeter: {meter_number}\nConsumption: {consumption} m³\nAmount Due: Ksh {amount}\nDue Date: {due_date}\nPay via M-Pesa: Paybill 123456 Account: {meter_number}\nThank you!",
			"variables":     []string{"{customer_name}", "{bill_number}", "{meter_number}", "{consumption}", "{amount}", "{due_date}"},
			"language":      "en",
			"is_active":     true,
			"created_at":    time.Now(),
			"updated_at":    time.Now(),
		},
		{
			"template_type": "sms",
			"name":          "Payment Confirmation",
			"body":          "Dear {customer_name},\nPayment of Ksh {amount} received for bill {bill_number}.\nReceipt: {receipt_number}\nTransaction: {transaction_id}\nBalance: Ksh {balance}\nThank you for paying on time!",
			"variables":     []string{"{customer_name}", "{amount}", "{bill_number}", "{receipt_number}", "{transaction_id}", "{balance}"},
			"language":      "en",
			"is_active":     true,
			"created_at":    time.Now(),
			"updated_at":    time.Now(),
		},
		{
			"template_type": "sms",
			"name":          "Disconnection Warning",
			"body":          "Dear {customer_name},\nYour water account {meter_number} has overdue balance of Ksh {amount}.\nPay before {final_date} to avoid disconnection.\nPay via M-Pesa: Paybill 123456 Account: {meter_number}",
			"variables":     []string{"{customer_name}", "{meter_number}", "{amount}", "{final_date}"},
			"language":      "en",
			"is_active":     true,
			"created_at":    time.Now(),
			"updated_at":    time.Now(),
		},
		{
			"template_type": "sms",
			"name":          "Reconnection Notice",
			"body":          "Dear {customer_name},\nYour water supply for meter {meter_number} has been reconnected.\nPlease ensure future payments are made on time to avoid disconnection.",
			"variables":     []string{"{customer_name}", "{meter_number}"},
			"language":      "en",
			"is_active":     true,
			"created_at":    time.Now(),
			"updated_at":    time.Now(),
		},
	}

	for _, template := range templates {
		// Check if template exists
		count, _ := collection.CountDocuments(ctx, bson.M{"name": template["name"]})
		if count == 0 {
			_, err := collection.InsertOne(ctx, template)
			if err != nil {
				log.Printf("Error creating template %s: %v", template["name"], err)
			} else {
				fmt.Printf("✓ Created notification template: %s\n", template["name"])
			}
		}
	}
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
