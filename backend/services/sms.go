package services

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"waterbilling/backend/models"

	"github.com/joho/godotenv"
	"github.com/twilio/twilio-go"
	twilioApi "github.com/twilio/twilio-go/rest/api/v2010"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type SMSService struct {
	client    *twilio.RestClient
	from      string
	db        *mongo.Database
	isEnabled bool
}

func NewSMSService(db *mongo.Database) (*SMSService, error) {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	accountSid := os.Getenv("TWILIO_ACCOUNT_SID")
	authToken := os.Getenv("TWILIO_AUTH_TOKEN")
	fromNumber := os.Getenv("TWILIO_PHONE_NUMBER")

	// If no Twilio credentials, check for Africa's Talking as alternative
	africastalkingAPIKey := os.Getenv("AFRICASTALKING_API_KEY")
	africastalkingUsername := os.Getenv("AFRICASTALKING_USERNAME")

	var client *twilio.RestClient
	var from string
	var isEnabled bool

	// Priority 1: Use Twilio if credentials are available
	if accountSid != "" && authToken != "" && fromNumber != "" {
		client = twilio.NewRestClientWithParams(twilio.ClientParams{
			Username: accountSid,
			Password: authToken,
		})
		from = fromNumber
		isEnabled = true
		log.Println("✅ SMS Service initialized with Twilio")
	} else if africastalkingAPIKey != "" && africastalkingUsername != "" {
		// Priority 2: Use Africa's Talking if configured
		log.Println("⚠️ Africa's Talking detected but not implemented. Using mock SMS.")
		isEnabled = false
	} else {
		// No SMS provider configured
		log.Println("⚠️ No SMS provider configured. Using mock SMS service.")
		isEnabled = false
	}

	return &SMSService{
		client:    client,
		from:      from,
		db:        db,
		isEnabled: isEnabled,
	}, nil
}

// SendSMS sends an SMS message using Twilio
func (s *SMSService) SendSMS(to, message string) error {
	if !s.isEnabled || s.client == nil {
		// Mock SMS for development
		log.Printf("[MOCK SMS] To: %s, Message: %s", to, message)
		return nil
	}

	params := &twilioApi.CreateMessageParams{}
	params.SetTo(s.formatPhoneNumber(to))
	params.SetFrom(s.from)
	params.SetBody(message)

	resp, err := s.client.Api.CreateMessage(params)
	if err != nil {
		log.Printf("Failed to send SMS: %v", err)
		return fmt.Errorf("failed to send SMS: %v", err)
	}

	if resp.Sid != nil {
		log.Printf("SMS sent successfully. SID: %s", *resp.Sid)
	} else {
		log.Printf("SMS sent successfully.")
	}

	return nil
}

// SendBillNotification sends a bill notification SMS to customer
func (s *SMSService) SendBillNotification(bill *models.Bill, customer *models.Customer) error {
	message := s.generateBillMessage(bill, customer)

	err := s.SendSMS(customer.PhoneNumber, message)

	// Log the SMS attempt
	s.logSMS(customer.ID, bill.ID, customer.PhoneNumber, message, err == nil, "bill_notification")

	return err
}

// SendPaymentConfirmation sends payment confirmation SMS
func (s *SMSService) SendPaymentConfirmation(payment *models.Payment, customer *models.Customer) error {
	message := fmt.Sprintf(
		"Dear %s,\n\n"+
			"Payment Received: Ksh %.2f\n"+
			"Receipt: %s\n"+
			"Meter: %s\n"+
			"Thank you for your payment!\n"+
			"Water Billing System",
		customer.FirstName,
		payment.Amount,
		payment.ReceiptNumber,
		payment.MeterNumber,
	)

	err := s.SendSMS(customer.PhoneNumber, message)
	s.logSMS(customer.ID, payment.BillID, customer.PhoneNumber, message, err == nil, "payment_confirmation")

	return err
}

// SendDisconnectionWarning sends disconnection warning SMS
func (s *SMSService) SendDisconnectionWarning(bill *models.Bill, customer *models.Customer) error {
	dueDate := bill.DueDate.Format("02 Jan 2006")

	message := fmt.Sprintf(
		"URGENT: Dear %s,\n\n"+
			"Your water account %s has overdue amount of Ksh %.2f\n"+
			"Original Due Date: %s\n"+
			"Pay within 48 hours to avoid disconnection.\n"+
			"Contact: 0700 000 000\n"+
			"Water Billing System",
		customer.FirstName,
		bill.MeterNumber,
		bill.Balance,
		dueDate,
	)

	err := s.SendSMS(customer.PhoneNumber, message)
	s.logSMS(customer.ID, bill.ID, customer.PhoneNumber, message, err == nil, "disconnection_warning")

	return err
}

// BulkSendBillNotifications sends notifications for multiple bills
func (s *SMSService) BulkSendBillNotifications(bills []models.Bill, customerMap map[primitive.ObjectID]models.Customer) map[string]error {
	results := make(map[string]error)

	for _, bill := range bills {
		customer, exists := customerMap[bill.CustomerID]
		if !exists {
			results[bill.BillNumber] = fmt.Errorf("customer not found for bill %s", bill.BillNumber)
			continue
		}

		err := s.SendBillNotification(&bill, &customer)
		results[bill.BillNumber] = err
	}

	return results
}

// generateBillMessage creates the SMS message for a bill
func (s *SMSService) generateBillMessage(bill *models.Bill, customer *models.Customer) string {
	dueDate := bill.DueDate.Format("02 Jan 2006")

	message := fmt.Sprintf(
		"Dear %s,\n\n"+
			"Water Bill: %s\n"+
			"Account: %s\n"+
			"Meter: %s\n"+
			"Previous Reading: %.2f\n"+
			"Current Reading: %.2f\n"+
			"Consumption: %.2f m³\n"+
			"Amount Due: Ksh %.2f\n"+
			"Due Date: %s\n\n"+
			"Pay via M-Pesa: Paybill 123456 Account: %s\n"+
			"Thank you!\n"+
			"Water Billing System",
		customer.FirstName,
		bill.BillNumber,
		bill.AccountNumber,
		bill.MeterNumber,
		bill.PreviousReading,
		bill.CurrentReading,
		bill.Consumption,
		bill.TotalAmount,
		dueDate,
		bill.MeterNumber,
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
	}

	if !success {
		smsLog.Status = "failed"
	}

	_, err := collection.InsertOne(ctx, smsLog)
	if err != nil {
		log.Printf("Failed to log SMS: %v", err)
	}
}

// formatPhoneNumber formats phone number for SMS
func (s *SMSService) formatPhoneNumber(phone string) string {
	// Remove any non-digit characters
	phone = strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, phone)

	// If starts with 0, replace with country code (assuming Kenya)
	if strings.HasPrefix(phone, "0") {
		phone = "254" + phone[1:]
	}

	// Add + prefix for international format
	if !strings.HasPrefix(phone, "+") {
		phone = "+" + phone
	}

	return phone
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

// GetDeliveryStatus checks delivery status for a message
func (s *SMSService) GetDeliveryStatus(messageID string) (string, error) {
	if !s.isEnabled || s.client == nil {
		return "mock_delivered", nil
	}

	// Twilio implementation would go here
	// For now, return mock status
	return "delivered", nil
}
