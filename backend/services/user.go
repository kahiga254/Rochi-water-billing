package services

import (
	"context"
	"fmt"
	"time"

	"waterbilling/backend/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	collection *mongo.Collection // ✅ THIS MUST BE HERE
}

func NewUserService(collection *mongo.Collection) *UserService {
	return &UserService{
		collection: collection, // ✅ Initialize the collection
	}
}

// GetUserByUsername retrieves a user by username
func (s *UserService) GetUserByUsername(username string) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var user models.User
	err := s.collection.FindOne(ctx, bson.M{"username": username}).Decode(&user)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("error fetching user: %v", err)
	}

	return &user, nil
}

// GetUserByEmail retrieves a user by email
func (s *UserService) GetUserByEmail(email string) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var user models.User
	err := s.collection.FindOne(ctx, bson.M{"email": email}).Decode(&user)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("error fetching user: %v", err)
	}

	return &user, nil
}

// GetUserByID retrieves a user by ID
func (s *UserService) GetUserByID(id string) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID format: %v", err)
	}

	var user models.User
	err = s.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&user)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("error fetching user: %v", err)
	}

	return &user, nil
}

// CreateUser creates a new user
// CreateUser creates a new user
func (s *UserService) CreateUser(user *models.User, password string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check if username already exists
	existingUser, _ := s.GetUserByUsername(user.Username)
	if existingUser != nil {
		return fmt.Errorf("user with username %s already exists", user.Username)
	}

	// Check if email already exists
	existingEmail, _ := s.GetUserByEmail(user.Email)
	if existingEmail != nil {
		return fmt.Errorf("user with email %s already exists", user.Email)
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("error hashing password: %v", err)
	}

	user.Password = string(hashedPassword)
	user.ID = primitive.NewObjectID()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	// Set default values for new fields if needed
	if user.MeterNumber == "" {
		user.MeterNumber = "" // This is fine, it can be empty for non-customers
	}

	_, err = s.collection.InsertOne(ctx, user)
	if err != nil {
		return fmt.Errorf("error creating user: %v", err)
	}

	return nil
}

// UpdateUser updates a user
func (s *UserService) UpdateUser(id string, updates map[string]interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid user ID format: %v", err)
	}

	update := bson.M{
		"$set": updates,
	}

	result, err := s.collection.UpdateByID(ctx, objectID, update)
	if err != nil {
		return fmt.Errorf("error updating user: %v", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// UpdateLastLogin updates user's last login timestamp
func (s *UserService) UpdateLastLogin(userID string, lastLogin time.Time) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID format: %v", err)
	}

	update := bson.M{
		"$set": bson.M{
			"last_login": lastLogin,
			"updated_at": time.Now(),
		},
	}

	_, err = s.collection.UpdateByID(ctx, objectID, update)
	return err
}

// Authenticate authenticates a user
func (s *UserService) Authenticate(username, password string) (*models.User, error) {
	user, err := s.GetUserByUsername(username)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	if !user.IsActive {
		return nil, fmt.Errorf("account is deactivated")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	return user, nil
}

// VerifyPassword verifies user's password
func (s *UserService) VerifyPassword(userID string, password string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID format: %v", err)
	}

	var user models.User
	err = s.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&user)
	if err != nil {
		return fmt.Errorf("user not found: %v", err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return fmt.Errorf("invalid password")
	}

	return nil
}

// ChangePassword changes user's password
func (s *UserService) ChangePassword(userID string, newPassword string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID format: %v", err)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("error hashing password: %v", err)
	}

	update := bson.M{
		"$set": bson.M{
			"password":   string(hashedPassword),
			"updated_at": time.Now(),
		},
	}

	result, err := s.collection.UpdateByID(ctx, objectID, update)
	if err != nil {
		return fmt.Errorf("error changing password: %v", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// ListUsers retrieves all users with pagination
func (s *UserService) ListUsers(filter bson.M, page, limit int64) ([]models.User, int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	skip := (page - 1) * limit
	opts := options.Find().
		SetSkip(skip).
		SetLimit(limit).
		SetSort(bson.M{"created_at": -1})

	cursor, err := s.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("error listing users: %v", err)
	}
	defer cursor.Close(ctx)

	var users []models.User
	if err = cursor.All(ctx, &users); err != nil {
		return nil, 0, fmt.Errorf("error decoding users: %v", err)
	}

	total, err := s.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("error counting users: %v", err)
	}

	return users, total, nil
}

// DeleteUser deletes a user
func (s *UserService) DeleteUser(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid user ID format: %v", err)
	}

	result, err := s.collection.DeleteOne(ctx, bson.M{"_id": objectID})
	if err != nil {
		return fmt.Errorf("error deleting user: %v", err)
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}
