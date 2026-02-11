package services

import (
	"context"
	"errors"
	"fmt"
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
}

func NewBillingService(customers, readings, bills, payments, tariffs *mongo.Collection) *BillingService {
	return &BillingService{
		customersCollection: customers,
		readingsCollection:  readings,
		billsCollection:     bills,
		paymentsCollection:  payments,
		tariffsCollection:   tariffs,
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

// GetCustomerTariff retrieves the tariff for a customer
func (bs *BillingService) GetCustomerTariff(tariffCode string) (*models.Tariff, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var tariff models.Tariff
	err := bs.tariffsCollection.FindOne(
		ctx,
		bson.M{"code": tariffCode, "is_active": true},
		options.FindOne().SetSort(bson.M{"effective_date": -1}),
	).Decode(&tariff)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("active tariff with code %s not found", tariffCode)
		}
		return nil, fmt.Errorf("error fetching tariff: %v", err)
	}

	return &tariff, nil
}

// SubmitMeterReading processes a new meter reading
func (bs *BillingService) SubmitMeterReading(readingRequest *models.MeterReading) (*models.Bill, error) {
	// Start session for transaction
	session, err := bs.readingsCollection.Database().Client().StartSession()
	if err != nil {
		return nil, fmt.Errorf("failed to start session: %v", err)
	}
	defer session.EndSession(context.Background())

	var resultBill *models.Bill
	err = mongo.WithSession(context.Background(), session, func(sc mongo.SessionContext) error {
		// Start transaction
		if err = session.StartTransaction(); err != nil {
			return fmt.Errorf("failed to start transaction: %v", err)
		}

		// 1. Get customer details
		customer, err := bs.GetCustomerByMeterNumber(readingRequest.MeterNumber)
		if err != nil {
			session.AbortTransaction(sc)
			return err
		}

		// 2. Get tariff information
		tariff, err := bs.GetCustomerTariff(customer.TariffCode)
		if err != nil {
			session.AbortTransaction(sc)
			return err
		}

		// 3. Get previous reading
		previousReading, err := bs.GetCustomerPreviousReading(readingRequest.MeterNumber)

		// Set previous reading value
		var previousReadingValue float64
		if previousReading != nil {
			previousReadingValue = previousReading.CurrentReading
		} else {
			// First reading for this customer
			previousReadingValue = customer.InitialReading
		}

		// 4. Validate and calculate consumption
		if readingRequest.CurrentReading < previousReadingValue {
			session.AbortTransaction(sc)
			return fmt.Errorf("current reading (%.2f) cannot be less than previous reading (%.2f)",
				readingRequest.CurrentReading, previousReadingValue)
		}

		consumption := readingRequest.CurrentReading - previousReadingValue

		// 5. Calculate charges
		waterCharge := consumption * tariff.BaseRate

		// Apply tiered pricing if available
		if len(tariff.Tiers) > 0 {
			waterCharge = bs.calculateTieredCharge(consumption, tariff.Tiers)
		}

		// 6. Prepare meter reading record
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
			RatePerUnit:     tariff.BaseRate,
			WaterCharge:     waterCharge,
			FixedCharge:     tariff.FixedCharge,
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

		// 7. Insert meter reading
		_, err = bs.readingsCollection.InsertOne(sc, reading)
		if err != nil {
			session.AbortTransaction(sc)
			return fmt.Errorf("failed to save meter reading: %v", err)
		}

		// 8. Generate bill
		bill, err := bs.generateBill(sc, customer, tariff, reading)
		if err != nil {
			session.AbortTransaction(sc)
			return err
		}

		// 9. Update customer with latest reading
		err = bs.updateCustomerLastReading(sc, customer.ID, reading.CurrentReading, reading.ReadingDate)
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

	return resultBill, nil
}

// calculateTieredCharge calculates water charge based on tiered pricing
func (bs *BillingService) calculateTieredCharge(consumption float64, tiers []models.TariffTier) float64 {
	var totalCharge float64
	remainingConsumption := consumption

	for _, tier := range tiers {
		if remainingConsumption <= 0 {
			break
		}

		tierRange := tier.MaxConsumption - tier.MinConsumption
		if tierRange <= 0 {
			continue
		}

		if remainingConsumption <= tierRange {
			totalCharge += remainingConsumption * tier.Rate
			break
		} else {
			totalCharge += tierRange * tier.Rate
			remainingConsumption -= tierRange
		}
	}

	return totalCharge
}

// generateBill creates a bill from a meter reading
func (bs *BillingService) generateBill(sc mongo.SessionContext, customer *models.Customer,
	tariff *models.Tariff, reading *models.MeterReading) (*models.Bill, error) {

	// Calculate arrears (previous balance)
	arrears := customer.Balance
	if arrears < 0 {
		arrears = -arrears // Convert negative balance to positive arrears
	} else {
		arrears = 0
	}

	// Calculate total amount
	totalAmount := reading.WaterCharge + reading.FixedCharge + arrears
	totalAmount = utils.RoundToTwoDecimal(totalAmount)

	// Generate bill
	bill := &models.Bill{
		ID:              primitive.NewObjectID(),
		MeterNumber:     customer.MeterNumber,
		CustomerID:      customer.ID,
		ReadingID:       reading.ID,
		AccountNumber:   customer.AccountNumber,
		CustomerName:    customer.FullName(),
		BillNumber:      utils.GenerateBillNumber(),
		BillDate:        time.Now(),
		DueDate:         time.Now().AddDate(0, 1, 0), // Due in 1 month
		BillingPeriod:   reading.BillingPeriod,
		PreviousReading: reading.PreviousReading,
		CurrentReading:  reading.CurrentReading,
		Consumption:     reading.Consumption,
		RatePerUnit:     reading.RatePerUnit,
		WaterCharge:     reading.WaterCharge,
		FixedCharge:     reading.FixedCharge,
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

// updateCustomerLastReading updates customer's last reading information
func (bs *BillingService) updateCustomerLastReading(sc mongo.SessionContext,
	customerID primitive.ObjectID, currentReading float64, readingDate time.Time) error {

	update := bson.M{
		"$set": bson.M{
			"last_reading":      currentReading,
			"last_reading_date": readingDate,
			"updated_at":        time.Now(),
		},
	}

	_, err := bs.customersCollection.UpdateByID(sc, customerID, update)
	if err != nil {
		return fmt.Errorf("failed to update customer reading: %v", err)
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

		// 5. Update customer balance
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

	// Update balance (subtract payment from negative balance)
	newBalance := customer.Balance + paymentAmount
	newBalance = utils.RoundToTwoDecimal(newBalance)

	update := bson.M{
		"$set": bson.M{
			"balance":    newBalance,
			"updated_at": time.Now(),
			"total_paid": customer.TotalPaid + paymentAmount,
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
		summary.StatusBreakdown[status] = StatusSummary{
			Count:       result["count"].(int32),
			TotalAmount: result["totalAmount"].(float64),
			TotalPaid:   result["totalPaid"].(float64),
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
