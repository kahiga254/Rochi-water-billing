package services

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"waterbilling/backend/models"
	"waterbilling/backend/utils"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type BillingService struct {
	customersCollection *mongo.Collection
	readingsCollection  *mongo.Collection
	billsCollection     *mongo.Collection
	paymentsCollection  *mongo.Collection
	tariffsCollection   *mongo.Collection
	smsService          *SMSService // ADDED: SMS service for notifications
}

// UPDATED: Added smsService parameter
func NewBillingService(customers, readings, bills, payments, tariffs *mongo.Collection, smsService *SMSService) *BillingService {
	return &BillingService{
		customersCollection: customers,
		readingsCollection:  readings,
		billsCollection:     bills,
		paymentsCollection:  payments,
		tariffsCollection:   tariffs,
		smsService:          smsService, // ADDED: Store SMS service
	}
}

// GetCustomerByMeterNumber retrieves a customer by meter number
func (bs *BillingService) GetCustomerByMeterNumber(meterNumber string) (*models.Customer, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var customer models.Customer
	err := bs.customersCollection.FindOne(ctx, bson.M{"meter_number": meterNumber}).Decode(&customer)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("customer with meter number %s not found", meterNumber)
		}
		return nil, fmt.Errorf("error fetching customer: %v", err)
	}

	return &customer, nil
}

// GetCustomerPreviousReading gets the last reading for a customer
func (bs *BillingService) GetCustomerPreviousReading(meterNumber string) (*models.MeterReading, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var reading models.MeterReading
	opts := options.FindOne().SetSort(bson.M{"reading_date": -1})
	err := bs.readingsCollection.FindOne(
		ctx,
		bson.M{"meter_number": meterNumber},
		opts,
	).Decode(&reading)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil // No previous reading found (new customer)
		}
		return nil, fmt.Errorf("error fetching previous reading: %v", err)
	}

	return &reading, nil
}

// SubmitMeterReading processes a new meter reading with FLAT RATE pricing
func (bs *BillingService) SubmitMeterReading(readingRequest *models.MeterReading) (*models.Bill, error) {
	// Start session for transaction
	session, err := bs.readingsCollection.Database().Client().StartSession()
	if err != nil {
		return nil, fmt.Errorf("failed to start session: %v", err)
	}
	defer session.EndSession(context.Background())

	var resultBill *models.Bill
	var customer *models.Customer // Moved outside for SMS access

	err = mongo.WithSession(context.Background(), session, func(sc mongo.SessionContext) error {
		// Start transaction
		if err = session.StartTransaction(); err != nil {
			return fmt.Errorf("failed to start transaction: %v", err)
		}

		// 1. Get customer details
		customer, err = bs.GetCustomerByMeterNumber(readingRequest.MeterNumber)
		if err != nil {
			session.AbortTransaction(sc)
			return err
		}

		// 2. Get previous reading
		previousReading, err := bs.GetCustomerPreviousReading(readingRequest.MeterNumber)

		// Set previous reading value
		var previousReadingValue float64
		if previousReading != nil {
			previousReadingValue = previousReading.CurrentReading
		} else {
			// First reading for this customer
			previousReadingValue = customer.InitialReading
		}

		// 3. Validate and calculate consumption
		if readingRequest.CurrentReading < previousReadingValue {
			session.AbortTransaction(sc)
			return fmt.Errorf("current reading (%.2f) cannot be less than previous reading (%.2f)",
				readingRequest.CurrentReading, previousReadingValue)
		}

		consumption := readingRequest.CurrentReading - previousReadingValue

		// 4. Calculate charges using SIMPLE FLAT RATE (KSh 100 per unit)
		ratePerUnit := 100.0 // KSh 100 per unit
		waterCharge := consumption * ratePerUnit
		fixedCharge := 0.0 // No fixed charges
		arrears := 0.0     // Start with zero arrears

		// If customer has negative balance, add to arrears
		if customer.Balance < 0 {
			arrears = -customer.Balance
		}

		// 5. Prepare meter reading record
		reading := &models.MeterReading{
			ID:              primitive.NewObjectID(),
			MeterNumber:     readingRequest.MeterNumber,
			CustomerID:      customer.ID,
			AccountNumber:   customer.AccountNumber,
			CustomerName:    customer.FullName(),
			ReadingDate:     readingRequest.ReadingDate,
			PreviousReading: previousReadingValue,
			CurrentReading:  readingRequest.CurrentReading,
			Consumption:     consumption,
			RatePerUnit:     ratePerUnit,
			WaterCharge:     waterCharge,
			FixedCharge:     fixedCharge,
			ReadingType:     readingRequest.ReadingType,
			ReadingMethod:   readingRequest.ReadingMethod,
			ReaderID:        readingRequest.ReaderID,
			ReaderName:      readingRequest.ReaderName,
			Month:           readingRequest.ReadingDate.Format("2006-01"),
			Year:            readingRequest.ReadingDate.Year(),
			BillingPeriod:   utils.GetBillingPeriod(readingRequest.ReadingDate),
			Status:          "recorded",
			CreatedAt:       time.Now(),
		}

		// 6. Insert meter reading
		_, err = bs.readingsCollection.InsertOne(sc, reading)
		if err != nil {
			session.AbortTransaction(sc)
			return fmt.Errorf("failed to save meter reading: %v", err)
		}

		// 7. Generate bill
		bill, err := bs.generateBill(sc, customer, reading, arrears)
		if err != nil {
			session.AbortTransaction(sc)
			return err
		}

		// 8. Update customer with latest reading and new balance
		err = bs.updateCustomerAfterBilling(sc, customer.ID, reading.CurrentReading, reading.ReadingDate, bill.TotalAmount)
		if err != nil {
			session.AbortTransaction(sc)
			return err
		}

		resultBill = bill

		// Commit transaction
		if err = session.CommitTransaction(sc); err != nil {
			return fmt.Errorf("failed to commit transaction: %v", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// ============ NEW: SMS NOTIFICATION ============
	// Send SMS notification to customer (non-blocking)
	if resultBill != nil && customer != nil && customer.PhoneNumber != "" {
		// Send SMS in a goroutine so it doesn't block the response
		go bs.sendBillSMSNotification(resultBill, customer)
	} else {
		if customer == nil {
			log.Println("⚠️ Cannot send SMS: customer is nil")
		} else if customer.PhoneNumber == "" {
			log.Printf("⚠️ Cannot send SMS: customer %s has no phone number", customer.MeterNumber)
		}
	}

	return resultBill, nil
}

// NEW: Send bill SMS notification
func (bs *BillingService) sendBillSMSNotification(bill *models.Bill, customer *models.Customer) {
	// Small delay to ensure bill is fully saved
	time.Sleep(200 * time.Millisecond)

	// Format billing period
	month := bill.BillingPeriod
	if month == "" {
		month = time.Now().Format("January 2006")
	}

	// Format due date
	dueDate := bill.DueDate.Format("02 Jan 2006")

	// Format the SMS message exactly as you requested
	message := fmt.Sprintf(`Dear %s,

Your %s water bill is KSh %.0f

Previous reading: %.1f units
Current reading: %.1f units
Consumption: %.1f units × KSh 100 = KSh %.0f

Please make your payment of KSh %.0f by %s for smooth running of our services.

Thank you,
Water Billing System`,
		customer.FullName(),
		month,
		bill.TotalAmount,
		bill.PreviousReading,
		bill.CurrentReading,
		bill.Consumption,
		bill.WaterCharge,
		bill.TotalAmount,
		dueDate)

	// Send the SMS
	log.Printf("📱 Sending SMS to %s (%s)", customer.FullName(), customer.PhoneNumber)
	err := bs.smsService.SendSMS(customer.PhoneNumber, message)

	if err != nil {
		log.Printf("❌ Failed to send SMS to %s: %v", customer.PhoneNumber, err)
	} else {
		log.Printf("✅ SMS sent successfully to %s (%s) for bill %s",
			customer.FullName(), customer.PhoneNumber, bill.BillNumber)

		// Update bill to mark SMS as sent
		bs.markSMSAsSent(bill.ID)
	}
}

// NEW: Mark SMS as sent in the bill record
func (bs *BillingService) markSMSAsSent(billID primitive.ObjectID) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	update := bson.M{
		"$set": bson.M{
			"sms_sent":    true,
			"sms_sent_at": time.Now(),
		},
	}

	_, err := bs.billsCollection.UpdateByID(ctx, billID, update)
	if err != nil {
		log.Printf("⚠️ Failed to update SMS sent status for bill %s: %v", billID.Hex(), err)
	}
}

// generateBill creates a bill from a meter reading using FLAT RATE pricing
func (bs *BillingService) generateBill(sc mongo.SessionContext, customer *models.Customer,
	reading *models.MeterReading, arrears float64) (*models.Bill, error) {

	// Calculate total amount: water charge + arrears (no fixed charges)
	totalAmount := reading.WaterCharge + arrears
	totalAmount = utils.RoundToTwoDecimal(totalAmount)

	// Generate bill number
	billNumber := "BILL-" + reading.MeterNumber + "-" + reading.ReadingDate.Format("200601")

	// Generate bill
	bill := &models.Bill{
		ID:              primitive.NewObjectID(),
		MeterNumber:     customer.MeterNumber,
		CustomerID:      customer.ID,
		ReadingID:       reading.ID,
		AccountNumber:   customer.AccountNumber,
		CustomerName:    customer.FullName(),
		BillNumber:      billNumber,
		BillDate:        time.Now(),
		DueDate:         time.Now().AddDate(0, 1, 0), // Due in 1 month
		BillingPeriod:   reading.BillingPeriod,
		PreviousReading: reading.PreviousReading,
		CurrentReading:  reading.CurrentReading,
		Consumption:     reading.Consumption,
		RatePerUnit:     reading.RatePerUnit,
		WaterCharge:     reading.WaterCharge,
		FixedCharge:     0.0, // No fixed charges
		Arrears:         arrears,
		TotalAmount:     totalAmount,
		Balance:         totalAmount, // Initially balance equals total amount
		Status:          "pending",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	// Insert bill
	_, err := bs.billsCollection.InsertOne(sc, bill)
	if err != nil {
		return nil, fmt.Errorf("failed to create bill: %v", err)
	}

	return bill, nil
}

// updateCustomerAfterBilling updates customer's last reading and adds the new bill amount to balance
func (bs *BillingService) updateCustomerAfterBilling(sc mongo.SessionContext,
	customerID primitive.ObjectID, currentReading float64, readingDate time.Time, billAmount float64) error {

	// Get current customer to get current balance
	var customer models.Customer
	err := bs.customersCollection.FindOne(sc, bson.M{"_id": customerID}).Decode(&customer)
	if err != nil {
		return fmt.Errorf("customer not found: %v", err)
	}

	// New balance = current balance - bill amount (since positive balance is credit)
	newBalance := customer.Balance - billAmount
	newBalance = utils.RoundToTwoDecimal(newBalance)

	// Calculate total consumed
	totalConsumed := customer.TotalConsumed
	if customer.LastReading > 0 {
		totalConsumed += (currentReading - customer.LastReading)
	} else {
		totalConsumed += currentReading
	}

	update := bson.M{
		"$set": bson.M{
			"last_reading":      currentReading,
			"last_reading_date": readingDate,
			"balance":           newBalance,
			"updated_at":        time.Now(),
			"total_consumed":    totalConsumed,
		},
	}

	_, err = bs.customersCollection.UpdateByID(sc, customerID, update)
	if err != nil {
		return fmt.Errorf("failed to update customer: %v", err)
	}

	return nil
}

// ProcessPayment processes a payment for a bill
func (bs *BillingService) ProcessPayment(payment *models.Payment) error {
	session, err := bs.paymentsCollection.Database().Client().StartSession()
	if err != nil {
		return fmt.Errorf("failed to start session: %v", err)
	}
	defer session.EndSession(context.Background())

	err = mongo.WithSession(context.Background(), session, func(sc mongo.SessionContext) error {
		if err = session.StartTransaction(); err != nil {
			return fmt.Errorf("failed to start transaction: %v", err)
		}

		// 1. Get the bill
		var bill models.Bill
		err := bs.billsCollection.FindOne(sc, bson.M{"_id": payment.BillID}).Decode(&bill)
		if err != nil {
			session.AbortTransaction(sc)
			return fmt.Errorf("bill not found: %v", err)
		}

		// 2. Validate payment amount
		if payment.Amount <= 0 {
			session.AbortTransaction(sc)
			return errors.New("payment amount must be greater than 0")
		}

		// 3. Create payment record
		payment.ID = primitive.NewObjectID()
		payment.PaymentDate = time.Now()
		payment.Status = "completed"
		payment.CreatedAt = time.Now()

		// Generate receipt number if not provided
		if payment.ReceiptNumber == "" {
			payment.ReceiptNumber = utils.GenerateReceiptNumber()
		}

		_, err = bs.paymentsCollection.InsertOne(sc, payment)
		if err != nil {
			session.AbortTransaction(sc)
			return fmt.Errorf("failed to save payment: %v", err)
		}

		// 4. Update bill payment status
		bill.UpdatePayment(payment.Amount, payment.PaymentMethod, payment.TransactionID)
		bill.UpdatedAt = time.Now()

		_, err = bs.billsCollection.ReplaceOne(sc, bson.M{"_id": bill.ID}, bill)
		if err != nil {
			session.AbortTransaction(sc)
			return fmt.Errorf("failed to update bill: %v", err)
		}

		// 5. Update customer balance (add payment to balance)
		err = bs.updateCustomerBalance(sc, bill.CustomerID, payment.Amount)
		if err != nil {
			session.AbortTransaction(sc)
			return err
		}

		if err = session.CommitTransaction(sc); err != nil {
			return fmt.Errorf("failed to commit transaction: %v", err)
		}

		return nil
	})

	return err
}

// updateCustomerBalance updates customer's balance after payment
func (bs *BillingService) updateCustomerBalance(sc mongo.SessionContext,
	customerID primitive.ObjectID, paymentAmount float64) error {

	// Get current customer
	var customer models.Customer
	err := bs.customersCollection.FindOne(sc, bson.M{"_id": customerID}).Decode(&customer)
	if err != nil {
		return fmt.Errorf("customer not found: %v", err)
	}

	// Update balance (add payment to balance - since positive balance = credit)
	newBalance := customer.Balance + paymentAmount
	newBalance = utils.RoundToTwoDecimal(newBalance)

	update := bson.M{
		"$set": bson.M{
			"balance":    newBalance,
			"updated_at": time.Now(),
		},
		"$inc": bson.M{
			"total_paid": paymentAmount,
		},
	}

	_, err = bs.customersCollection.UpdateByID(sc, customerID, update)
	if err != nil {
		return fmt.Errorf("failed to update customer balance: %v", err)
	}

	return nil
}

// GetCustomerBills retrieves all bills for a customer by meter number
func (bs *BillingService) GetCustomerBills(meterNumber string, status string, limit int64) ([]models.Bill, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"meter_number": meterNumber}
	if status != "" {
		filter["status"] = status
	}

	opts := options.Find().SetSort(bson.M{"bill_date": -1})
	if limit > 0 {
		opts.SetLimit(limit)
	}

	cursor, err := bs.billsCollection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("error fetching bills: %v", err)
	}
	defer cursor.Close(ctx)

	var bills []models.Bill
	if err = cursor.All(ctx, &bills); err != nil {
		return nil, fmt.Errorf("error decoding bills: %v", err)
	}

	return bills, nil
}

// GetCustomerReadingHistory gets reading history for a customer
func (bs *BillingService) GetCustomerReadingHistory(meterNumber string, limit int64) ([]models.MeterReading, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	opts := options.Find().SetSort(bson.M{"reading_date": -1})
	if limit > 0 {
		opts.SetLimit(limit)
	}

	cursor, err := bs.readingsCollection.Find(ctx, bson.M{"meter_number": meterNumber}, opts)
	if err != nil {
		return nil, fmt.Errorf("error fetching reading history: %v", err)
	}
	defer cursor.Close(ctx)

	var readings []models.MeterReading
	if err = cursor.All(ctx, &readings); err != nil {
		return nil, fmt.Errorf("error decoding readings: %v", err)
	}

	return readings, nil
}

// GetOverdueBills returns all overdue bills
func (bs *BillingService) GetOverdueBills() ([]models.Bill, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"status":   "pending",
		"due_date": bson.M{"$lt": time.Now()},
	}

	cursor, err := bs.billsCollection.Find(ctx, filter, options.Find().SetSort(bson.M{"due_date": 1}))
	if err != nil {
		return nil, fmt.Errorf("error fetching overdue bills: %v", err)
	}
	defer cursor.Close(ctx)

	var bills []models.Bill
	if err = cursor.All(ctx, &bills); err != nil {
		return nil, fmt.Errorf("error decoding overdue bills: %v", err)
	}

	return bills, nil
}

// GetUnpaidBills returns all unpaid bills (pending and overdue)
func (bs *BillingService) GetUnpaidBills() ([]models.Bill, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"status": bson.M{"$in": []string{"pending", "overdue"}},
	}

	cursor, err := bs.billsCollection.Find(ctx, filter, options.Find().SetSort(bson.M{"due_date": 1}))
	if err != nil {
		return nil, fmt.Errorf("error fetching unpaid bills: %v", err)
	}
	defer cursor.Close(ctx)

	var bills []models.Bill
	if err = cursor.All(ctx, &bills); err != nil {
		return nil, fmt.Errorf("error decoding unpaid bills: %v", err)
	}

	return bills, nil
}

// GetBillingSummary returns billing summary for a period
func (bs *BillingService) GetBillingSummary(startDate, endDate time.Time) (*BillingSummary, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Match bills within date range
	matchStage := bson.D{
		{Key: "$match", Value: bson.D{
			{Key: "bill_date", Value: bson.D{
				{Key: "$gte", Value: startDate},
				{Key: "$lte", Value: endDate},
			}},
		}},
	}

	// Group by status and calculate totals
	groupStage := bson.D{
		{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$status"},
			{Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}},
			{Key: "totalAmount", Value: bson.D{{Key: "$sum", Value: "$total_amount"}}},
			{Key: "totalPaid", Value: bson.D{{Key: "$sum", Value: "$amount_paid"}}},
		}},
	}

	cursor, err := bs.billsCollection.Aggregate(ctx, mongo.Pipeline{matchStage, groupStage})
	if err != nil {
		return nil, fmt.Errorf("error aggregating billing summary: %v", err)
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err = cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("error decoding summary results: %v", err)
	}

	summary := &BillingSummary{
		PeriodStart:     startDate,
		PeriodEnd:       endDate,
		StatusBreakdown: make(map[string]StatusSummary),
	}

	for _, result := range results {
		status := result["_id"].(string)

		// Handle MongoDB numeric types safely
		var count int32
		switch v := result["count"].(type) {
		case int32:
			count = v
		case int64:
			count = int32(v)
		case float64:
			count = int32(v)
		}

		totalAmount, _ := result["totalAmount"].(float64)
		totalPaid, _ := result["totalPaid"].(float64)

		summary.StatusBreakdown[status] = StatusSummary{
			Count:       count,
			TotalAmount: totalAmount,
			TotalPaid:   totalPaid,
		}
	}

	return summary, nil
}

// BillingSummary represents billing summary data
type BillingSummary struct {
	PeriodStart     time.Time                `json:"period_start"`
	PeriodEnd       time.Time                `json:"period_end"`
	StatusBreakdown map[string]StatusSummary `json:"status_breakdown"`
}

// StatusSummary represents summary for a specific bill status
type StatusSummary struct {
	Count       int32   `json:"count"`
	TotalAmount float64 `json:"total_amount"`
	TotalPaid   float64 `json:"total_paid"`
}
