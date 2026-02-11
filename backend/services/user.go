package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"waterbilling/backend/models"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	usersCollection *mongo.Collection
}

func NewUserService(usersCollection *mongo.Collection) *UserService {
	return &UserService{
		usersCollection: usersCollection,
	}
}

// CreateUser creates a new user
func (us *UserService) CreateUser(user *models.User, password string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check if username already exists
	existingUser, _ := us.GetUserByUsername(user.Username)
	if existingUser != nil {
		return fmt.Errorf("user with username %s already exists", user.Username)
	}

	// Check if email already exists
	existingEmail, _ := us.GetUserByEmail(user.Email)
	if existingEmail != nil {
		return fmt.Errorf("user with email %s already exists", user.Email)
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %v", err)
	}

	// Set user properties
	user.ID = primitive.NewObjectID()
	user.Password = string(hashedPassword)

	if user.EmployeeID == "" {
		uuidStr := uuid.New().String()
		user.EmployeeID = "EMP-" + strings.ToUpper(uuidStr[:8])
	}

	if user.CreatedAt.IsZero() {
		user.CreatedAt = time.Now()
	}

	user.UpdatedAt = time.Now()

	// Insert user
	_, err = us.usersCollection.InsertOne(ctx, user)
	if err != nil {
		return fmt.Errorf("failed to create user: %v", err)
	}

	return nil
}

// GetUserByUsername gets user by username
func (us *UserService) GetUserByUsername(username string) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var user models.User
	err := us.usersCollection.FindOne(ctx, bson.M{"username": username}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("error fetching user: %v", err)
	}

	return &user, nil
}

// GetUserByEmail gets user by email
func (us *UserService) GetUserByEmail(email string) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var user models.User
	err := us.usersCollection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("error fetching user: %v", err)
	}

	return &user, nil
}

// GetUserByID gets user by ID
func (us *UserService) GetUserByID(id string) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %v", err)
	}

	var user models.User
	err = us.usersCollection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("error fetching user: %v", err)
	}

	return &user, nil
}

// UpdateUser updates user information
func (us *UserService) UpdateUser(id string, updates map[string]interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid user ID: %v", err)
	}

	// Remove fields that shouldn't be updated
	delete(updates, "_id")
	delete(updates, "password")
	delete(updates, "created_at")

	updates["updated_at"] = time.Now()

	// Check for email uniqueness if email is being updated
	if email, ok := updates["email"].(string); ok {
		existingUser, _ := us.GetUserByEmail(email)
		if existingUser != nil && existingUser.ID != objectID {
			return fmt.Errorf("user with email %s already exists", email)
		}
	}

	update := bson.M{"$set": updates}
	result, err := us.usersCollection.UpdateOne(
		ctx,
		bson.M{"_id": objectID},
		update,
	)

	if err != nil {
		return fmt.Errorf("error updating user: %v", err)
	}

	if result.MatchedCount == 0 {
		return errors.New("user not found")
	}

	return nil
}

// ChangePassword changes user password
func (us *UserService) ChangePassword(id string, newPassword string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid user ID: %v", err)
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %v", err)
	}

	update := bson.M{
		"$set": bson.M{
			"password":   string(hashedPassword),
			"updated_at": time.Now(),
		},
	}

	result, err := us.usersCollection.UpdateOne(
		ctx,
		bson.M{"_id": objectID},
		update,
	)

	if err != nil {
		return fmt.Errorf("error changing password: %v", err)
	}

	if result.MatchedCount == 0 {
		return errors.New("user not found")
	}

	return nil
}

// GetUsers gets all users with pagination
func (us *UserService) GetUsers(role string, isActive *bool, skip, limit int64) ([]models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{}

	if role != "" {
		filter["role"] = role
	}

	if isActive != nil {
		filter["is_active"] = *isActive
	}

	opts := options.Find().
		SetSort(bson.M{"first_name": 1}).
		SetSkip(skip).
		SetLimit(limit)

	cursor, err := us.usersCollection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("error fetching users: %v", err)
	}
	defer cursor.Close(ctx)

	var users []models.User
	if err = cursor.All(ctx, &users); err != nil {
		return nil, fmt.Errorf("error decoding users: %v", err)
	}

	return users, nil
}

// UpdateUserStatus updates user active status
func (us *UserService) UpdateUserStatus(id string, isActive bool) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid user ID: %v", err)
	}

	update := bson.M{
		"$set": bson.M{
			"is_active":  isActive,
			"updated_at": time.Now(),
		},
	}

	result, err := us.usersCollection.UpdateOne(
		ctx,
		bson.M{"_id": objectID},
		update,
	)

	if err != nil {
		return fmt.Errorf("error updating user status: %v", err)
	}

	if result.MatchedCount == 0 {
		return errors.New("user not found")
	}

	return nil
}
