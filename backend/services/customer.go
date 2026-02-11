package services

import (
	"context"
	"fmt"
	"time"

	"waterbilling/backend/models"
	"waterbilling/backend/utils"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type CustomerService struct {
	customersCollection *mongo.Collection
	tariffsCollection   *mongo.Collection
}

func NewCustomerService(customers, tariffs *mongo.Collection) *CustomerService {
	return &CustomerService{
		customersCollection: customers,
		tariffsCollection:   tariffs,
	}
}

// CreateCustomer creates a new customer
func (cs *CustomerService) CreateCustomer(customer *models.Customer) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Validate meter number
	if !utils.ValidateMeterNumber(customer.MeterNumber) {
		return fmt.Errorf("invalid meter number format")
	}

	// Format phone number
	customer.PhoneNumber = utils.FormatPhoneNumber(customer.PhoneNumber)

	// Set default values
	if customer.ConnectionDate.IsZero() {
		customer.ConnectionDate = time.Now()
	}

	if customer.TariffCode == "" {
		customer.TariffCode = "RES-BASIC" // Default tariff
	}

	if customer.CustomerType == "" {
		customer.CustomerType = "residential"
	}

	if customer.ConnectionType == "" {
		customer.ConnectionType = "metered"
	}

	if customer.Status == "" {
		customer.Status = "active"
	}

	customer.CreatedAt = time.Now()
	customer.UpdatedAt = time.Now()
	customer.ID = primitive.NewObjectID()

	// Check if meter number already exists
	existing, _ := cs.GetCustomerByMeterNumber(customer.MeterNumber)
	if existing != nil {
		return fmt.Errorf("customer with meter number %s already exists", customer.MeterNumber)
	}

	// Insert customer
	_, err := cs.customersCollection.InsertOne(ctx, customer)
	if err != nil {
		return fmt.Errorf("failed to create customer: %v", err)
	}

	return nil
}

// GetCustomerByMeterNumber retrieves customer by meter number
func (cs *CustomerService) GetCustomerByMeterNumber(meterNumber string) (*models.Customer, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var customer models.Customer
	err := cs.customersCollection.FindOne(ctx, bson.M{"meter_number": meterNumber}).Decode(&customer)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("error fetching customer: %v", err)
	}

	return &customer, nil
}

// UpdateCustomer updates customer information
func (cs *CustomerService) UpdateCustomer(meterNumber string, updates map[string]interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Remove fields that shouldn't be updated
	delete(updates, "_id")
	delete(updates, "meter_number")
	delete(updates, "created_at")

	// Format phone number if being updated
	if phone, ok := updates["phone_number"].(string); ok {
		updates["phone_number"] = utils.FormatPhoneNumber(phone)
	}

	updates["updated_at"] = time.Now()

	update := bson.M{"$set": updates}
	result, err := cs.customersCollection.UpdateOne(
		ctx,
		bson.M{"meter_number": meterNumber},
		update,
	)

	if err != nil {
		return fmt.Errorf("error updating customer: %v", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("customer with meter number %s not found", meterNumber)
	}

	return nil
}

// SearchCustomers searches customers by various criteria
func (cs *CustomerService) SearchCustomers(searchTerm string, zone string, status string,
	customerType string, limit int64) ([]models.Customer, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{}

	// Text search
	if searchTerm != "" {
		filter["$text"] = bson.M{"$search": searchTerm}
	}

	// Filter by zone
	if zone != "" {
		filter["zone"] = zone
	}

	// Filter by status
	if status != "" {
		filter["status"] = status
	}

	// Filter by customer type
	if customerType != "" {
		filter["customer_type"] = customerType
	}

	opts := options.Find()
	if limit > 0 {
		opts.SetLimit(limit)
	}
	opts.SetSort(bson.M{"first_name": 1})

	cursor, err := cs.customersCollection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("error searching customers: %v", err)
	}
	defer cursor.Close(ctx)

	var customers []models.Customer
	if err = cursor.All(ctx, &customers); err != nil {
		return nil, fmt.Errorf("error decoding customers: %v", err)
	}

	return customers, nil
}

// GetCustomersByZone gets all customers in a specific zone
func (cs *CustomerService) GetCustomersByZone(zone string) ([]models.Customer, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := cs.customersCollection.Find(
		ctx,
		bson.M{"zone": zone, "status": "active"},
		options.Find().SetSort(bson.M{"meter_number": 1}),
	)
	if err != nil {
		return nil, fmt.Errorf("error fetching customers by zone: %v", err)
	}
	defer cursor.Close(ctx)

	var customers []models.Customer
	if err = cursor.All(ctx, &customers); err != nil {
		return nil, fmt.Errorf("error decoding customers: %v", err)
	}

	return customers, nil
}

// UpdateCustomerStatus updates customer status
func (cs *CustomerService) UpdateCustomerStatus(meterNumber string, status string, reason string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	update := bson.M{
		"$set": bson.M{
			"status":               status,
			"disconnection_reason": reason,
			"updated_at":           time.Now(),
		},
	}

	if status == "disconnected" {
		update["$set"].(bson.M)["disconnection_date"] = time.Now()
	} else if status == "active" {
		now := time.Now()
		update["$set"].(bson.M)["reconnection_date"] = &now
	}

	result, err := cs.customersCollection.UpdateOne(
		ctx,
		bson.M{"meter_number": meterNumber},
		update,
	)

	if err != nil {
		return fmt.Errorf("error updating customer status: %v", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("customer with meter number %s not found", meterNumber)
	}

	return nil
}

// GetCustomerStatistics returns customer statistics
func (cs *CustomerService) GetCustomerStatistics() (*CustomerStatistics, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Total customers
	total, err := cs.customersCollection.CountDocuments(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("error counting total customers: %v", err)
	}

	// Active customers
	active, err := cs.customersCollection.CountDocuments(ctx, bson.M{"status": "active"})
	if err != nil {
		return nil, fmt.Errorf("error counting active customers: %v", err)
	}

	// Inactive customers
	inactive, err := cs.customersCollection.CountDocuments(ctx, bson.M{"status": "inactive"})
	if err != nil {
		return nil, fmt.Errorf("error counting inactive customers: %v", err)
	}

	// Disconnected customers
	disconnected, err := cs.customersCollection.CountDocuments(ctx, bson.M{"status": "disconnected"})
	if err != nil {
		return nil, fmt.Errorf("error counting disconnected customers: %v", err)
	}

	// Customers by type
	pipeline := mongo.Pipeline{
		bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$customer_type"},
			{Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}},
		}}},
	}

	cursor, err := cs.customersCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("error aggregating customer types: %v", err)
	}
	defer cursor.Close(ctx)

	var typeResults []bson.M
	if err = cursor.All(ctx, &typeResults); err != nil {
		return nil, fmt.Errorf("error decoding customer types: %v", err)
	}

	customerTypes := make(map[string]int64)
	for _, result := range typeResults {
		customerTypes[result["_id"].(string)] = result["count"].(int64)
	}

	// Customers by zone
	zonePipeline := mongo.Pipeline{
		bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$zone"},
			{Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}},
		}}},
		bson.D{{Key: "$sort", Value: bson.D{{Key: "count", Value: -1}}}},
		bson.D{{Key: "$limit", Value: 10}},
	}

	zoneCursor, err := cs.customersCollection.Aggregate(ctx, zonePipeline)
	if err != nil {
		return nil, fmt.Errorf("error aggregating zones: %v", err)
	}
	defer zoneCursor.Close(ctx)

	var zoneResults []bson.M
	if err = zoneCursor.All(ctx, &zoneResults); err != nil {
		return nil, fmt.Errorf("error decoding zones: %v", err)
	}

	topZones := make(map[string]int64)
	for _, result := range zoneResults {
		topZones[result["_id"].(string)] = result["count"].(int64)
	}

	return &CustomerStatistics{
		Total:         total,
		Active:        active,
		Inactive:      inactive,
		Disconnected:  disconnected,
		CustomerTypes: customerTypes,
		TopZones:      topZones,
	}, nil
}

// CustomerStatistics represents customer statistics
type CustomerStatistics struct {
	Total         int64            `json:"total"`
	Active        int64            `json:"active"`
	Inactive      int64            `json:"inactive"`
	Disconnected  int64            `json:"disconnected"`
	CustomerTypes map[string]int64 `json:"customer_types"`
	TopZones      map[string]int64 `json:"top_zones"`
}
