package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GeoLocation represents GPS coordinates for meter reading location
type GeoLocation struct {
	Type        string    `bson:"type" json:"type" default:"Point"`
	Coordinates []float64 `bson:"coordinates" json:"coordinates"` // [longitude, latitude]
}

// Customer represents a water company customer
type Customer struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	MeterNumber    string             `bson:"meter_number" json:"meter_number"`     // Primary identifier (UNIQUE)
	AccountNumber  string             `bson:"account_number" json:"account_number"` // Alternative ID (UNIQUE)
	FirstName      string             `bson:"first_name" json:"first_name"`
	LastName       string             `bson:"last_name" json:"last_name"`
	PhoneNumber    string             `bson:"phone_number" json:"phone_number"`
	Email          string             `bson:"email,omitempty" json:"email,omitempty"`
	IDNumber       string             `bson:"id_number,omitempty" json:"id_number,omitempty"` // National ID/Passport
	Address        Address            `bson:"address" json:"address"`
	CustomerType   string             `bson:"customer_type" json:"customer_type"`               // "residential", "commercial", "industrial", "institutional"
	ConnectionType string             `bson:"connection_type" json:"connection_type"`           // "metered", "unmetered"
	MeterType      string             `bson:"meter_type,omitempty" json:"meter_type,omitempty"` // "digital", "analog", "smart"
	Zone           string             `bson:"zone" json:"zone"`                                 // Administrative zone/ward
	Subzone        string             `bson:"subzone,omitempty" json:"subzone,omitempty"`       // Smaller area within zone
	TariffCode     string             `bson:"tariff_code" json:"tariff_code"`                   // Different rates for different customer types
	RatePerUnit    float64            `bson:"rate_per_unit" json:"rate_per_unit" default:"100.0"`
	FixedCharge    float64            `bson:"fixed_charge" json:"fixed_charge" default:"0"`

	// Meter Information
	MeterBrand            string    `bson:"meter_brand,omitempty" json:"meter_brand,omitempty"`
	MeterSize             string    `bson:"meter_size,omitempty" json:"meter_size,omitempty"` // e.g., "15mm", "20mm"
	MeterInstallationDate time.Time `bson:"meter_installation_date,omitempty" json:"meter_installation_date,omitempty"`
	MeterLocation         string    `bson:"meter_location,omitempty" json:"meter_location,omitempty"` // "indoors", "outdoors", "compound"

	// Reading Information
	InitialReading     float64    `bson:"initial_reading,omitempty" json:"initial_reading,omitempty"`
	ConnectionDate     time.Time  `bson:"connection_date" json:"connection_date"`
	LastReadingDate    *time.Time `bson:"last_reading_date,omitempty" json:"last_reading_date,omitempty"`
	LastReading        float64    `bson:"last_reading,omitempty" json:"last_reading,omitempty"`
	AverageConsumption float64    `bson:"average_consumption,omitempty" json:"average_consumption,omitempty"`

	// Financial Information
	Balance       float64 `bson:"balance" json:"balance" default:"0"` // Positive = credit, Negative = arrears
	TotalPaid     float64 `bson:"total_paid,omitempty" json:"total_paid,omitempty"`
	TotalConsumed float64 `bson:"total_consumed,omitempty" json:"total_consumed,omitempty"`

	// Status Information
	Status              string     `bson:"status" json:"status" default:"active"` // "active", "inactive", "disconnected", "pending", "suspended"
	DisconnectionReason string     `bson:"disconnection_reason,omitempty" json:"disconnection_reason,omitempty"`
	ReconnectionDate    *time.Time `bson:"reconnection_date,omitempty" json:"reconnection_date,omitempty"`

	// Additional Information
	EmergencyContact  string `bson:"emergency_contact,omitempty" json:"emergency_contact,omitempty"`
	EmergencyPhone    string `bson:"emergency_phone,omitempty" json:"emergency_phone,omitempty"`
	PropertyOwner     string `bson:"property_owner,omitempty" json:"property_owner,omitempty"`
	PropertyType      string `bson:"property_type,omitempty" json:"property_type,omitempty"` // "apartment", "house", "commercial"
	NumberOfOccupants int    `bson:"number_of_occupants,omitempty" json:"number_of_occupants,omitempty"`
	Notes             string `bson:"notes,omitempty" json:"notes,omitempty"`

	// Timestamps
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
}

// Address represents a complete address structure
type Address struct {
	StreetAddress string `bson:"street_address" json:"street_address"`
	City          string `bson:"city" json:"city"`
	State         string `bson:"state,omitempty" json:"state,omitempty"`
	PostalCode    string `bson:"postal_code,omitempty" json:"postal_code,omitempty"`
	Country       string `bson:"country" json:"country" default:"Kenya"`
	Landmark      string `bson:"landmark,omitempty" json:"landmark,omitempty"` // e.g., "near school", "opposite hospital"
}

// MeterReading represents a single meter reading entry
type MeterReading struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	MeterNumber   string             `bson:"meter_number" json:"meter_number"` // Foreign key to Customer
	CustomerID    primitive.ObjectID `bson:"customer_id" json:"customer_id"`
	AccountNumber string             `bson:"account_number" json:"account_number"`
	CustomerName  string             `bson:"customer_name" json:"customer_name"` // For quick reference: FirstName + LastName

	// Reading Details
	ReadingDate     time.Time `bson:"reading_date" json:"reading_date"`
	PreviousReading float64   `bson:"previous_reading" json:"previous_reading"`
	CurrentReading  float64   `bson:"current_reading" json:"current_reading"`
	Consumption     float64   `bson:"consumption" json:"consumption"` // Calculated: current - previous

	// Charges
	RatePerUnit float64 `bson:"rate_per_unit" json:"rate_per_unit"`
	WaterCharge float64 `bson:"water_charge" json:"water_charge"` // consumption * rate
	FixedCharge float64 `bson:"fixed_charge" json:"fixed_charge"`
	Arrears     float64 `bson:"arrears,omitempty" json:"arrears,omitempty"` // Previous balance brought forward
	Penalty     float64 `bson:"penalty,omitempty" json:"penalty,omitempty"`
	Discount    float64 `bson:"discount,omitempty" json:"discount,omitempty"`
	TotalAmount float64 `bson:"total_amount" json:"total_amount"` // Sum of all charges

	// Reading Metadata
	ReadingType   string             `bson:"reading_type" json:"reading_type"`     // "manual", "estimated", "actual", "self-read"
	ReadingMethod string             `bson:"reading_method" json:"reading_method"` // "mobile_app", "field_agent", "customer_portal", "sms"
	ReaderID      primitive.ObjectID `bson:"reader_id,omitempty" json:"reader_id,omitempty"`
	ReaderName    string             `bson:"reader_name" json:"reader_name"`

	// Location & Verification
	Location         GeoLocation `bson:"location,omitempty" json:"location,omitempty"`
	MeterPhotoURL    string      `bson:"meter_photo_url,omitempty" json:"meter_photo_url,omitempty"`
	IsVerified       bool        `bson:"is_verified" json:"is_verified" default:"false"`
	VerifiedBy       string      `bson:"verified_by,omitempty" json:"verified_by,omitempty"`
	VerificationDate *time.Time  `bson:"verification_date,omitempty" json:"verification_date,omitempty"`

	// Additional Info
	MeterCondition string `bson:"meter_condition,omitempty" json:"meter_condition,omitempty"` // "good", "damaged", "tampered"
	Notes          string `bson:"notes,omitempty" json:"notes,omitempty"`

	// Time Period
	Month         string `bson:"month" json:"month"` // Format: "YYYY-MM"
	Year          int    `bson:"year" json:"year"`
	BillingPeriod string `bson:"billing_period" json:"billing_period"`     // e.g., "January 2024"
	Season        string `bson:"season,omitempty" json:"season,omitempty"` // "dry", "wet", "normal"

	// Status
	Status        string `bson:"status" json:"status"` // "recorded", "billed", "verified", "disputed"
	DisputeReason string `bson:"dispute_reason,omitempty" json:"dispute_reason,omitempty"`
	Resolution    string `bson:"resolution,omitempty" json:"resolution,omitempty"`
	// Timestamps
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
}

// Bill represents a generated bill for a customer
type Bill struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	MeterNumber   string             `bson:"meter_number" json:"meter_number"`
	CustomerID    primitive.ObjectID `bson:"customer_id" json:"customer_id"`
	ReadingID     primitive.ObjectID `bson:"reading_id" json:"reading_id"`
	AccountNumber string             `bson:"account_number" json:"account_number"`
	CustomerName  string             `bson:"customer_name" json:"customer_name"`

	// Bill Identification
	BillNumber    string    `bson:"bill_number" json:"bill_number"` // Auto-generated: BILL-YYYYMM-XXXX
	BillDate      time.Time `bson:"bill_date" json:"bill_date"`
	DueDate       time.Time `bson:"due_date" json:"due_date"`
	BillingPeriod string    `bson:"billing_period" json:"billing_period"` // Format: "January 2024"

	// Reading Information
	PreviousReading float64 `bson:"previous_reading" json:"previous_reading"`
	CurrentReading  float64 `bson:"current_reading" json:"current_reading"`
	Consumption     float64 `bson:"consumption" json:"consumption"`

	// Charges Breakdown
	RatePerUnit  float64 `bson:"rate_per_unit" json:"rate_per_unit"`
	WaterCharge  float64 `bson:"water_charge" json:"water_charge"` // consumption * rate
	FixedCharge  float64 `bson:"fixed_charge" json:"fixed_charge"`
	Arrears      float64 `bson:"arrears" json:"arrears"`                     // Previous balance
	Penalty      float64 `bson:"penalty,omitempty" json:"penalty,omitempty"` // Late payment penalty
	Discount     float64 `bson:"discount,omitempty" json:"discount,omitempty"`
	Tax          float64 `bson:"tax,omitempty" json:"tax,omitempty"` // VAT or other taxes
	OtherCharges float64 `bson:"other_charges,omitempty" json:"other_charges,omitempty"`
	TotalAmount  float64 `bson:"total_amount" json:"total_amount"`

	// Payment Information
	AmountPaid    float64    `bson:"amount_paid" json:"amount_paid" default:"0"`
	Balance       float64    `bson:"balance" json:"balance"` // total_amount - amount_paid
	Status        string     `bson:"status" json:"status"`   // "pending", "paid", "overdue", "partially_paid", "cancelled"
	PaymentDate   *time.Time `bson:"payment_date,omitempty" json:"payment_date,omitempty"`
	PaymentMethod string     `bson:"payment_method,omitempty" json:"payment_method,omitempty"` // "cash", "mpesa", "bank", "cheque", "credit_card"
	TransactionID string     `bson:"transaction_id,omitempty" json:"transaction_id,omitempty"`
	ReceiptNumber string     `bson:"receipt_number,omitempty" json:"receipt_number,omitempty"`
	PaymentNotes  string     `bson:"payment_notes,omitempty" json:"payment_notes,omitempty"`

	// Notification Status
	SMSsent     bool       `bson:"sms_sent" json:"sms_sent" default:"false"`
	SMSsentAt   *time.Time `bson:"sms_sent_at,omitempty" json:"sms_sent_at,omitempty"`
	EmailSent   bool       `bson:"email_sent" json:"email_sent" default:"false"`
	EmailSentAt *time.Time `bson:"email_sent_at,omitempty" json:"email_sent_at,omitempty"`
	Printed     bool       `bson:"printed" json:"printed" default:"false"`
	PrintedAt   *time.Time `bson:"printed_at,omitempty" json:"printed_at,omitempty"`

	// Timestamps
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
}

// User represents system users (admin, meter readers, cashiers, etc.)
type User struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	FirstName    string             `bson:"first_name" json:"first_name"`
	LastName     string             `bson:"last_name" json:"last_name"`
	Email        string             `bson:"email" json:"email"`
	PhoneNumber  string             `bson:"phone_number" json:"phone_number"`
	Username     string             `bson:"username" json:"username"`
	Password     string             `bson:"password" json:"-"` // Hashed password
	Role         string             `bson:"role" json:"role"`  // "admin", "reader", "cashier", "manager", "customer_service"
	MeterNumber  string             `bson:"meter_number,omitempty" json:"meter_number,omitempty"`
	Department   string             `bson:"department,omitempty" json:"department,omitempty"`
	EmployeeID   string             `bson:"employee_id,omitempty" json:"employee_id,omitempty"`
	AssignedZone string             `bson:"assigned_zone,omitempty" json:"assigned_zone,omitempty"` // For meter readers
	Permissions  []string           `bson:"permissions,omitempty" json:"permissions,omitempty"`     // Fine-grained permissions
	IsActive     bool               `bson:"is_active" json:"is_active" default:"true"`
	LastLogin    *time.Time         `bson:"last_login,omitempty" json:"last_login,omitempty"`
	CreatedAt    time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt    time.Time          `bson:"updated_at" json:"updated_at"`
}

// Payment represents a payment transaction
type Payment struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	BillID        primitive.ObjectID `bson:"bill_id" json:"bill_id"`
	MeterNumber   string             `bson:"meter_number" json:"meter_number"`
	CustomerID    primitive.ObjectID `bson:"customer_id" json:"customer_id"`
	CustomerName  string             `bson:"customer_name" json:"customer_name"`
	PaymentDate   time.Time          `bson:"payment_date" json:"payment_date"`
	Amount        float64            `bson:"amount" json:"amount"`
	PaymentMethod string             `bson:"payment_method" json:"payment_method"` // "cash", "mpesa", "bank", "cheque"
	TransactionID string             `bson:"transaction_id" json:"transaction_id"` // MPesa code, bank ref, etc.
	ReceiptNumber string             `bson:"receipt_number" json:"receipt_number"`
	PayerName     string             `bson:"payer_name,omitempty" json:"payer_name,omitempty"`
	PayerPhone    string             `bson:"payer_phone,omitempty" json:"payer_phone,omitempty"`
	CollectedBy   string             `bson:"collected_by" json:"collected_by"` // User who collected payment
	Status        string             `bson:"status" json:"status"`             // "completed", "pending", "failed", "refunded"
	Notes         string             `bson:"notes,omitempty" json:"notes,omitempty"`
	CreatedAt     time.Time          `bson:"created_at" json:"created_at"`
}

// SMSLog tracks sent messages
type SMSLog struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	CustomerID   primitive.ObjectID `bson:"customer_id" json:"customer_id"`
	BillID       primitive.ObjectID `bson:"bill_id,omitempty" json:"bill_id,omitempty"`
	MeterNumber  string             `bson:"meter_number" json:"meter_number"`
	PhoneNumber  string             `bson:"phone_number" json:"phone_number"`
	CustomerName string             `bson:"customer_name,omitempty" json:"customer_name,omitempty"`
	MessageType  string             `bson:"message_type" json:"message_type"` // "bill_notification", "payment_confirmation", "reminder", "disconnection_warning"
	Message      string             `bson:"message" json:"message"`
	Status       string             `bson:"status" json:"status"`                             // "sent", "failed", "delivered", "pending"
	Provider     string             `bson:"provider,omitempty" json:"provider,omitempty"`     // "twilio", "africas_talking", "nexmo"
	MessageID    string             `bson:"message_id,omitempty" json:"message_id,omitempty"` // Provider's message ID
	Cost         float64            `bson:"cost,omitempty" json:"cost,omitempty"`
	Error        string             `bson:"error,omitempty" json:"error,omitempty"`
	SentAt       time.Time          `bson:"sent_at" json:"sent_at"`
}

// NotificationTemplate for SMS/Email messages
type NotificationTemplate struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	TemplateType string             `bson:"template_type" json:"template_type"` // "sms", "email"
	Name         string             `bson:"name" json:"name"`
	Subject      string             `bson:"subject,omitempty" json:"subject,omitempty"` // For emails
	Body         string             `bson:"body" json:"body"`
	Variables    []string           `bson:"variables" json:"variables"` // e.g., ["{customer_name}", "{amount}", "{due_date}"]
	Language     string             `bson:"language" json:"language" default:"en"`
	IsActive     bool               `bson:"is_active" json:"is_active" default:"true"`
	CreatedAt    time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt    time.Time          `bson:"updated_at" json:"updated_at"`
}

// Tariff defines water pricing structure
type Tariff struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Code         string             `bson:"code" json:"code"` // e.g., "RES-A", "COM-B"
	Name         string             `bson:"name" json:"name"` // e.g., "Residential Basic"
	CustomerType string             `bson:"customer_type" json:"customer_type"`
	Description  string             `bson:"description,omitempty" json:"description,omitempty"`

	// Rate Structure (could be tiered)
	BaseRate    float64 `bson:"base_rate" json:"base_rate"`       // Rate per cubic meter
	FixedCharge float64 `bson:"fixed_charge" json:"fixed_charge"` // Monthly fixed charge

	// Tiered rates (optional)
	Tiers []TariffTier `bson:"tiers,omitempty" json:"tiers,omitempty"`

	// Validity
	EffectiveDate time.Time  `bson:"effective_date" json:"effective_date"`
	ExpiryDate    *time.Time `bson:"expiry_date,omitempty" json:"expiry_date,omitempty"`
	IsActive      bool       `bson:"is_active" json:"is_active" default:"true"`
	CreatedAt     time.Time  `bson:"created_at" json:"created_at"`
	UpdatedAt     time.Time  `bson:"updated_at" json:"updated_at"`
}

// TariffTier for tiered pricing (e.g., 0-10m³ @ 50, 11-20m³ @ 75, etc.)
type TariffTier struct {
	MinConsumption float64 `bson:"min_consumption" json:"min_consumption"`
	MaxConsumption float64 `bson:"max_consumption" json:"max_consumption"`
	Rate           float64 `bson:"rate" json:"rate"`
}

// Helper Methods for Customer
func (c *Customer) FullName() string {
	return c.FirstName + " " + c.LastName
}

func (c *Customer) UpdateLastReading(reading float64, date time.Time) {
	c.LastReading = reading
	c.LastReadingDate = &date
}

// Helper Methods for Bill
func (b *Bill) IsOverdue() bool {
	return b.Status == "pending" && time.Now().After(b.DueDate)
}

func (b *Bill) UpdatePayment(amount float64, method string, transactionID string) {
	b.AmountPaid += amount
	b.Balance = b.TotalAmount - b.AmountPaid

	if b.Balance <= 0 {
		b.Status = "paid"
		now := time.Now()
		b.PaymentDate = &now
		b.PaymentMethod = method
		b.TransactionID = transactionID
	} else {
		b.Status = "partially_paid"
	}
	b.UpdatedAt = time.Now()
}
