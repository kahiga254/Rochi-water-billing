package services

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"waterbilling/backend/models"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type SMSService struct {
	apiKey    string
	username  string
	senderID  string
	db        *mongo.Database
	isEnabled bool
	provider  string
}

func NewSMSService(db *mongo.Database) (*SMSService, error) {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Get Africa's Talking credentials
	apiKey := os.Getenv("AFRICASTALKING_API_KEY")
	username := os.Getenv("AFRICASTALKING_USERNAME")
	senderID := os.Getenv("AFRICASTALKING_SENDER_ID")

	// Check if credentials are available
	if apiKey == "" || username == "" {
		log.Println("⚠️ Africa's Talking credentials not found. Using mock SMS service.")
		return &SMSService{
			db:        db,
			isEnabled: false,
			provider:  "mock",
		}, nil
	}

	log.Println("✅ SMS Service initialized with Africa's Talking (HTTP client)")
	return &SMSService{
		apiKey:    apiKey,
		username:  username,
		senderID:  senderID,
		db:        db,
		isEnabled: true,
		provider:  "africastalking",
	}, nil
}

// SendSMS sends an SMS message
func (s *SMSService) SendSMS(to, message string) error {
	if !s.isEnabled {
		log.Printf("[MOCK SMS] To: %s, Message: %s", to, message)
		return nil
	}
	return s.sendAfricasTalkingSMS(to, message)
}

// sendAfricasTalkingSMS sends SMS via Africa's Talking HTTP API
func (s *SMSService) sendAfricasTalkingSMS(to, message string) error {
	// Format phone number
	phone := s.formatPhoneNumberForAT(to)

	// Determine API environment
	apiURL := "https://api.africastalking.com/version1/messaging"
	if os.Getenv("APP_ENV") == "development" {
		apiURL = "https://api.sandbox.africastalking.com/version1/messaging"
	}

	// Prepare form data (x-www-form-urlencoded)
	formData := url.Values{}
	formData.Set("username", s.username)
	formData.Set("to", phone)
	formData.Set("message", message)

	// Add sender ID if available
	if s.senderID != "" {
		formData.Set("from", s.senderID)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", apiURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	// Set correct headers for Africa's Talking
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("apiKey", s.apiKey)

	// Send request
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("❌ Africa's Talking SMS failed: %v", err)
		return fmt.Errorf("failed to send SMS: %v", err)
	}
	defer resp.Body.Close()

	// Read response body for debugging
	body, _ := io.ReadAll(resp.Body)

	// Check response
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		log.Printf("❌ Africa's Talking error (%d): %s", resp.StatusCode, string(body))
		return fmt.Errorf("SMS API returned status: %d", resp.StatusCode)
	}

	log.Printf("✅ Africa's Talking SMS sent to %s", phone)
	log.Printf("📥 Response: %s", string(body))
	return nil
}

// formatPhoneNumberForAT formats phone number for Africa's Talking (Kenya)
func (s *SMSService) formatPhoneNumberForAT(phone string) string {
	// Remove any non-digit characters
	phone = strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, phone)

	// Kenyan number formats: 0712345678 -> 254712345678
	if strings.HasPrefix(phone, "0") && len(phone) == 10 {
		phone = "254" + phone[1:]
	}

	// If starts with +254, remove the +
	if strings.HasPrefix(phone, "+254") {
		phone = phone[1:]
	}

	// Ensure we have the correct format (254XXXXXXXXX)
	if !strings.HasPrefix(phone, "254") {
		phone = "254" + phone
	}

	return phone
}

// SendBillNotification sends a bill notification SMS to customer
func (s *SMSService) SendBillNotification(bill *models.Bill, customer *models.Customer) error {
	message := s.generateBillMessage(bill, customer)
	err := s.SendSMS(customer.PhoneNumber, message)
	s.logSMS(customer.ID, bill.ID, customer.PhoneNumber, message, err == nil, "bill_notification")
	return err
}

// SendPaymentConfirmation sends payment confirmation SMS
func (s *SMSService) SendPaymentConfirmation(payment *models.Payment, customer *models.Customer) error {
	message := fmt.Sprintf(
		"Dear %s,\n\n"+
			"✅ Payment Received: KSh %.2f\n"+
			"Receipt: %s\n"+
			"Meter: %s\n"+
			"Date: %s\n\n"+
			"Thank you for your payment!\n"+
			"Rochi Pure Water",
		customer.FirstName,
		payment.Amount,
		payment.ReceiptNumber,
		payment.MeterNumber,
		payment.PaymentDate.Format("02 Jan 2006"),
	)

	err := s.SendSMS(customer.PhoneNumber, message)
	s.logSMS(customer.ID, payment.BillID, customer.PhoneNumber, message, err == nil, "payment_confirmation")
	return err
}

// SendDisconnectionWarning sends disconnection warning SMS
func (s *SMSService) SendDisconnectionWarning(bill *models.Bill, customer *models.Customer) error {
	dueDate := bill.DueDate.Format("02 Jan 2006")

	message := fmt.Sprintf(
		"⚠️ URGENT: Dear %s,\n\n"+
			"Your water account %s has overdue amount of KSh %.2f\n"+
			"Original Due Date: %s\n"+
			"Pay within 48 hours to avoid disconnection.\n\n"+
			"Contact: 0700 000 000\n"+
			"Rochi Pure Water",
		customer.FirstName,
		bill.MeterNumber,
		bill.Balance,
		dueDate,
	)

	err := s.SendSMS(customer.PhoneNumber, message)
	s.logSMS(customer.ID, bill.ID, customer.PhoneNumber, message, err == nil, "disconnection_warning")
	return err
}

// generateBillMessage creates the SMS message for a bill
func (s *SMSService) generateBillMessage(bill *models.Bill, customer *models.Customer) string {
	dueDate := bill.DueDate.Format("02 Jan 2006")

	message := fmt.Sprintf(
		"Dear %s,\n\n"+
			"Your water bill for %s is now ready.\n\n"+
			"Meter: %s\n"+
			"Previous Reading: %.1f units\n"+
			"Current Reading: %.1f units\n"+
			"Consumption: %.1f units\n"+
			"Amount Due: KSh %.0f\n"+
			"Due Date: %s\n\n"+
			"Please make payment to avoid service interruption.\n\n"+
			"Thank you,\n"+
			"Rochi Pure Water",
		customer.FirstName,
		bill.BillingPeriod,
		bill.MeterNumber,
		bill.PreviousReading,
		bill.CurrentReading,
		bill.Consumption,
		bill.TotalAmount,
		dueDate,
	)

	return message
}

// logSMS logs SMS sending to database
func (s *SMSService) logSMS(customerID, billID primitive.ObjectID, phone, message string, success bool, messageType string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := s.db.Collection("sms_logs")

	smsLog := models.SMSLog{
		ID:          primitive.NewObjectID(),
		CustomerID:  customerID,
		BillID:      billID,
		PhoneNumber: phone,
		Message:     message,
		MessageType: messageType,
		Status:      "sent",
		SentAt:      time.Now(),
		Provider:    s.provider,
	}

	if !success {
		smsLog.Status = "failed"
	}

	_, err := collection.InsertOne(ctx, smsLog)
	if err != nil {
		log.Printf("Failed to log SMS: %v", err)
	}
}

// GetSMSLogs retrieves SMS logs with optional filtering
func (s *SMSService) GetSMSLogs(filter bson.M, limit int64) ([]models.SMSLog, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := s.db.Collection("sms_logs")

	opts := options.Find().SetSort(bson.M{"sent_at": -1})
	if limit > 0 {
		opts.SetLimit(limit)
	}

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch SMS logs: %v", err)
	}
	defer cursor.Close(ctx)

	var logs []models.SMSLog
	if err = cursor.All(ctx, &logs); err != nil {
		return nil, fmt.Errorf("failed to decode SMS logs: %v", err)
	}

	return logs, nil
}

// IsEnabled returns true if SMS service is enabled
func (s *SMSService) IsEnabled() bool {
	return s.isEnabled
}
