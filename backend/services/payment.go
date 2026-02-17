package services

import (
	"context"
	"fmt"
	"time"

	"waterbilling/backend/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type PaymentService struct {
	collection *mongo.Collection
}

func NewPaymentService(collection *mongo.Collection) *PaymentService {
	return &PaymentService{
		collection: collection,
	}
}

// CreatePayment inserts a new payment record
func (s *PaymentService) CreatePayment(payment *models.Payment) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := s.collection.InsertOne(ctx, payment)
	if err != nil {
		return fmt.Errorf("failed to create payment: %v", err)
	}

	return nil
}

// GetPaymentsByMeter retrieves payments for a specific meter
func (s *PaymentService) GetPaymentsByMeter(meterNumber string, limit int) ([]models.Payment, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"meter_number": meterNumber}
	opts := options.Find().SetSort(bson.M{"payment_date": -1}).SetLimit(int64(limit))

	cursor, err := s.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("error fetching payments: %v", err)
	}
	defer cursor.Close(ctx)

	var payments []models.Payment
	if err = cursor.All(ctx, &payments); err != nil {
		return nil, fmt.Errorf("error decoding payments: %v", err)
	}

	return payments, nil
}
