package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var (
	// Client is the MongoDB client
	Client *mongo.Client

	// DB is the database instance
	DB *mongo.Database

	// once ensures the connection is only established once
	once sync.Once

	// connected tracks connection status
	connected bool

	// connectionError holds any connection error
	connectionError error
)

// Config holds database configuration
type Config struct {
	URI      string
	Database string
	Username string
	Password string
	Timeout  time.Duration
	PoolSize uint64
}

// Connect establishes a connection to MongoDB
func Connect() error {
	once.Do(func() {
		connectionError = connect()
	})

	return connectionError
}

// connect performs the actual connection
func connect() error {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Get configuration from environment
	config := getConfig()

	// Set client options
	clientOptions := options.Client().
		ApplyURI(config.URI).
		SetMaxPoolSize(config.PoolSize).
		SetMinPoolSize(10).
		SetMaxConnIdleTime(30 * time.Minute).
		SetConnectTimeout(config.Timeout).
		SetServerSelectionTimeout(10 * time.Second)

	// Add authentication if credentials are provided
	if config.Username != "" && config.Password != "" {
		clientOptions.SetAuth(options.Credential{
			Username: config.Username,
			Password: config.Password,
		})
	}

	// Add retry logic for connection
	var client *mongo.Client
	var err error

	// Try to connect with retries
	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		log.Printf("Attempting to connect to MongoDB (attempt %d/%d)...", i+1, maxRetries)

		ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)
		client, err = mongo.Connect(ctx, clientOptions)
		cancel()

		if err != nil {
			log.Printf("Connection attempt %d failed: %v", i+1, err)
			if i < maxRetries-1 {
				time.Sleep(time.Duration(i+1) * time.Second) // Exponential backoff
				continue
			}
			return fmt.Errorf("failed to connect to MongoDB after %d attempts: %v", maxRetries, err)
		}

		// Ping the database to verify connection
		ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err = client.Ping(ctx, readpref.Primary()); err != nil {
			log.Printf("Ping attempt %d failed: %v", i+1, err)
			if i < maxRetries-1 {
				client.Disconnect(ctx)
				time.Sleep(time.Duration(i+1) * time.Second)
				continue
			}
			return fmt.Errorf("failed to ping MongoDB after %d attempts: %v", maxRetries, err)
		}

		break
	}

	// Set global variables
	Client = client
	DB = client.Database(config.Database)
	connected = true

	log.Printf("✅ Successfully connected to MongoDB database: %s", config.Database)
	log.Printf("📊 Connection stats: MaxPoolSize=%d, MinPoolSize=10", config.PoolSize)

	return nil
}

// getConfig loads configuration from environment
func getConfig() *Config {
	// Default configuration
	config := &Config{
		URI:      "mongodb://localhost:27017",
		Database: "water_billing",
		Timeout:  10 * time.Second,
		PoolSize: 100,
	}

	// Override with environment variables
	if uri := os.Getenv("MONGODB_URI"); uri != "" {
		config.URI = uri
	}

	if db := os.Getenv("DB_NAME"); db != "" {
		config.Database = db
	}

	if user := os.Getenv("MONGODB_USERNAME"); user != "" {
		config.Username = user
	}

	if pass := os.Getenv("MONGODB_PASSWORD"); pass != "" {
		config.Password = pass
	}

	// Parse timeout if provided
	if timeoutStr := os.Getenv("MONGODB_TIMEOUT"); timeoutStr != "" {
		if timeout, err := time.ParseDuration(timeoutStr); err == nil {
			config.Timeout = timeout
		}
	}

	// Parse pool size if provided
	if poolSizeStr := os.Getenv("MONGODB_POOL_SIZE"); poolSizeStr != "" {
		if poolSize, err := parseUint64(poolSizeStr); err == nil && poolSize > 0 {
			config.PoolSize = poolSize
		}
	}

	return config
}

// Disconnect closes the MongoDB connection
func Disconnect() error {
	if Client == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := Client.Disconnect(ctx)
	if err != nil {
		return fmt.Errorf("failed to disconnect from MongoDB: %v", err)
	}

	connected = false
	log.Println("✅ MongoDB connection closed")

	return nil
}

// IsConnected returns true if database is connected
func IsConnected() bool {
	if !connected || Client == nil {
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := Client.Ping(ctx, readpref.Primary())
	return err == nil
}

// GetCollection returns a collection from the database
func GetCollection(collectionName string) *mongo.Collection {
	if DB == nil {
		log.Printf("Warning: Database not initialized, attempting to connect...")
		if err := Connect(); err != nil {
			log.Printf("Failed to connect to database: %v", err)
			return nil
		}
	}

	return DB.Collection(collectionName)
}

// GetCollectionWithOptions returns a collection with custom options
func GetCollectionWithOptions(collectionName string, opts *options.CollectionOptions) *mongo.Collection {
	if DB == nil {
		log.Printf("Warning: Database not initialized, attempting to connect...")
		if err := Connect(); err != nil {
			log.Printf("Failed to connect to database: %v", err)
			return nil
		}
	}

	return DB.Collection(collectionName, opts)
}

// WithTransaction executes a function within a transaction
func WithTransaction(ctx context.Context, fn func(sessCtx mongo.SessionContext) error) error {
	if Client == nil {
		return fmt.Errorf("database client not initialized")
	}

	session, err := Client.StartSession()
	if err != nil {
		return fmt.Errorf("failed to start session: %v", err)
	}
	defer session.EndSession(ctx)

	// Execute the function within a transaction
	err = session.StartTransaction()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %v", err)
	}

	// Use WithTransaction helper
	err = mongo.WithSession(ctx, session, func(sessCtx mongo.SessionContext) error {
		if err := fn(sessCtx); err != nil {
			return err
		}

		return session.CommitTransaction(sessCtx)
	})

	if err != nil {
		// Try to abort transaction on error
		abortErr := session.AbortTransaction(ctx)
		if abortErr != nil {
			log.Printf("Failed to abort transaction: %v", abortErr)
		}
		return err
	}

	return nil
}

// HealthCheck performs a health check on the database
func HealthCheck(ctx context.Context) error {
	if Client == nil {
		return fmt.Errorf("database client not initialized")
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err := Client.Ping(ctx, readpref.Primary())
	if err != nil {
		return fmt.Errorf("database health check failed: %v", err)
	}

	return nil
}

// GetDatabaseStats returns database statistics
func GetDatabaseStats(ctx context.Context) (map[string]interface{}, error) {
	if DB == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Run dbStats command
	cmd := bson.D{{Key: "dbStats", Value: 1}}
	var result bson.M

	err := DB.RunCommand(ctx, cmd).Decode(&result)
	if err != nil {
		return nil, fmt.Errorf("failed to get database stats: %v", err)
	}

	return result, nil
}

// GetCollectionStats returns statistics for a specific collection
func GetCollectionStats(ctx context.Context, collectionName string) (map[string]interface{}, error) {
	collection := GetCollection(collectionName)
	if collection == nil {
		return nil, fmt.Errorf("collection %s not found", collectionName)
	}

	// Get count of documents
	count, err := collection.EstimatedDocumentCount(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get document count: %v", err)
	}

	// Get storage size (this would require running collStats command)
	// For simplicity, we'll return basic stats

	stats := map[string]interface{}{
		"collection":     collectionName,
		"document_count": count,
		"database":       DB.Name(),
	}

	return stats, nil
}

// CreateIndex creates an index on a collection
func CreateIndex(ctx context.Context, collectionName string, keys interface{}, opts *options.IndexOptions) error {
	collection := GetCollection(collectionName)
	if collection == nil {
		return fmt.Errorf("collection %s not found", collectionName)
	}

	indexModel := mongo.IndexModel{
		Keys:    keys,
		Options: opts,
	}

	_, err := collection.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		return fmt.Errorf("failed to create index on %s: %v", collectionName, err)
	}

	return nil
}

// DropCollection drops a collection from the database
func DropCollection(ctx context.Context, collectionName string) error {
	collection := GetCollection(collectionName)
	if collection == nil {
		return fmt.Errorf("collection %s not found", collectionName)
	}

	err := collection.Drop(ctx)
	if err != nil {
		return fmt.Errorf("failed to drop collection %s: %v", collectionName, err)
	}

	log.Printf("Dropped collection: %s", collectionName)
	return nil
}

// ListCollections returns all collection names in the database
func ListCollections(ctx context.Context) ([]string, error) {
	if DB == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	collections, err := DB.ListCollectionNames(ctx, bson.D{})
	if err != nil {
		return nil, fmt.Errorf("failed to list collections: %v", err)
	}

	return collections, nil
}

// BulkInsert performs bulk insertion of documents
func BulkInsert(ctx context.Context, collectionName string, documents []interface{}) (*mongo.InsertManyResult, error) {
	collection := GetCollection(collectionName)
	if collection == nil {
		return nil, fmt.Errorf("collection %s not found", collectionName)
	}

	result, err := collection.InsertMany(ctx, documents)
	if err != nil {
		return nil, fmt.Errorf("failed to bulk insert into %s: %v", collectionName, err)
	}

	return result, nil
}

// helper function to parse uint64
func parseUint64(s string) (uint64, error) {
	var n uint64
	_, err := fmt.Sscanf(s, "%d", &n)
	return n, err
}
