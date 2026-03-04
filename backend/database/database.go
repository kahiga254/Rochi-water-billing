package database

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net"
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
	log.Println("🔍 [DEBUG] Connect() called")
	once.Do(func() {
		log.Println("🔍 [DEBUG] Executing connection once.Do()")
		connectionError = connect()
	})

	if connectionError != nil {
		log.Printf("❌ [DEBUG] Connection error: %v", connectionError)
	} else {
		log.Println("✅ [DEBUG] Connection successful")
	}
	return connectionError
}

// connect performs the actual connection
func connect() error {
	log.Println("🔍 [DEBUG] connect() started")

	// Load environment variables
	log.Println("🔍 [DEBUG] Loading .env file...")
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️ [DEBUG] No .env file found, using environment variables")
	}

	// Get configuration from environment
	log.Println("🔍 [DEBUG] Getting config from environment...")
	config := getConfig()
	log.Printf("🔍 [DEBUG] Config loaded - Database: %s, Timeout: %v, PoolSize: %d",
		config.Database, config.Timeout, config.PoolSize)

	// Log URI (mask password)
	maskedURI := maskPassword(config.URI)
	log.Printf("🔍 [DEBUG] MongoDB URI: %s", maskedURI)
	log.Printf("🔍 [DEBUG] Username provided: %v", config.Username != "")
	log.Printf("🔍 [DEBUG] Password provided: %v", config.Password != "")

	// Set client options
	log.Println("🔍 [DEBUG] Setting client options...")
	clientOptions := options.Client().
		ApplyURI(config.URI).
		SetMaxPoolSize(config.PoolSize).
		SetMinPoolSize(10).
		SetMaxConnIdleTime(30 * time.Minute).
		SetConnectTimeout(config.Timeout).
		SetServerSelectionTimeout(10 * time.Second).
		SetRetryWrites(true).
		SetRetryReads(true)

	log.Println("🔍 [DEBUG] Setting TLS config with InsecureSkipVerify=true")
	clientOptions.SetTLSConfig(&tls.Config{
		InsecureSkipVerify: true,
	})

	// Add authentication if credentials are provided
	if config.Username != "" && config.Password != "" {
		log.Println("🔍 [DEBUG] Setting authentication credentials")
		clientOptions.SetAuth(options.Credential{
			Username: config.Username,
			Password: config.Password,
		})
	}

	// Try to resolve hostnames first (for debugging)
	log.Println("🔍 [DEBUG] Attempting DNS resolution of MongoDB hosts...")
	hosts := []string{
		"ac-kwr6zjv-shard-00-00.9wu5s9u.mongodb.net",
		"ac-kwr6zjv-shard-00-01.9wu5s9u.mongodb.net",
		"ac-kwr6zjv-shard-00-02.9wu5s9u.mongodb.net",
	}

	for _, host := range hosts {
		log.Printf("🔍 [DEBUG] Resolving %s...", host)
		ips, err := net.LookupIP(host)
		if err != nil {
			log.Printf("❌ [DEBUG] DNS resolution failed for %s: %v", host, err)
		} else {
			log.Printf("✅ [DEBUG] DNS resolution successful for %s: %v", host, ips)
		}
	}

	// Add retry logic for connection
	var client *mongo.Client
	var err error

	// Try to connect with retries
	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		log.Printf("🔍 [DEBUG] Connection attempt %d/%d starting...", i+1, maxRetries)

		ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)
		log.Printf("🔍 [DEBUG] Calling mongo.Connect() with timeout %v", config.Timeout)
		client, err = mongo.Connect(ctx, clientOptions)
		cancel()

		if err != nil {
			log.Printf("❌ [DEBUG] Connection attempt %d failed: %v", i+1, err)
			if i < maxRetries-1 {
				sleepTime := time.Duration(i+1) * time.Second
				log.Printf("🔍 [DEBUG] Retrying in %v...", sleepTime)
				time.Sleep(sleepTime)
				continue
			}
			return fmt.Errorf("failed to connect to MongoDB after %d attempts: %v", maxRetries, err)
		}

		log.Printf("✅ [DEBUG] Connection attempt %d succeeded, now pinging...", i+1)

		// Ping the database to verify connection
		pingCtx, pingCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer pingCancel()

		log.Println("🔍 [DEBUG] Pinging MongoDB primary...")
		if err = client.Ping(pingCtx, readpref.Primary()); err != nil {
			log.Printf("❌ [DEBUG] Ping attempt %d failed: %v", i+1, err)
			if i < maxRetries-1 {
				log.Println("🔍 [DEBUG] Disconnecting client before retry...")
				client.Disconnect(pingCtx)
				sleepTime := time.Duration(i+1) * time.Second
				log.Printf("🔍 [DEBUG] Retrying in %v...", sleepTime)
				time.Sleep(sleepTime)
				continue
			}
			return fmt.Errorf("failed to ping MongoDB after %d attempts: %v", maxRetries, err)
		}

		log.Printf("✅ [DEBUG] Ping attempt %d succeeded!", i+1)
		break
	}

	// Set global variables
	log.Println("🔍 [DEBUG] Setting global Client and DB variables")
	Client = client
	DB = client.Database(config.Database)
	connected = true

	log.Printf("✅ [DEBUG] Successfully connected to MongoDB database: %s", config.Database)
	log.Printf("📊 [DEBUG] Connection stats: MaxPoolSize=%d, MinPoolSize=10", config.PoolSize)

	return nil
}

// maskPassword hides the password in the URI for logging
func maskPassword(uri string) string {
	// Simple masking - just show first 30 chars and last 10
	if len(uri) > 50 {
		return uri[:30] + "..." + uri[len(uri)-10:]
	}
	return "[uri too short to mask safely]"
}

// getConfig loads configuration from environment
func getConfig() *Config {
	log.Println("🔍 [DEBUG] getConfig() called")

	// Default configuration
	config := &Config{
		URI:      "mongodb://localhost:27017",
		Database: "water_billing",
		Timeout:  10 * time.Second,
		PoolSize: 100,
	}
	log.Println("🔍 [DEBUG] Default config set")

	// Override with environment variables
	if uri := os.Getenv("MONGODB_URI"); uri != "" {
		log.Println("🔍 [DEBUG] MONGODB_URI found in environment")
		config.URI = uri
	} else {
		log.Println("⚠️ [DEBUG] MONGODB_URI NOT found in environment!")
	}

	if db := os.Getenv("DB_NAME"); db != "" {
		log.Printf("🔍 [DEBUG] DB_NAME found: %s", db)
		config.Database = db
	}

	if user := os.Getenv("MONGODB_USERNAME"); user != "" {
		log.Printf("🔍 [DEBUG] MONGODB_USERNAME found (length: %d)", len(user))
		config.Username = user
	}

	if pass := os.Getenv("MONGODB_PASSWORD"); pass != "" {
		log.Printf("🔍 [DEBUG] MONGODB_PASSWORD found (length: %d)", len(pass))
		config.Password = pass
	}

	// Parse timeout if provided
	if timeoutStr := os.Getenv("MONGODB_TIMEOUT"); timeoutStr != "" {
		log.Printf("🔍 [DEBUG] MONGODB_TIMEOUT found: %s", timeoutStr)
		if timeout, err := time.ParseDuration(timeoutStr); err == nil {
			config.Timeout = timeout
			log.Printf("🔍 [DEBUG] Timeout set to: %v", timeout)
		} else {
			log.Printf("⚠️ [DEBUG] Failed to parse MONGODB_TIMEOUT: %v", err)
		}
	}

	// Parse pool size if provided
	if poolSizeStr := os.Getenv("MONGODB_POOL_SIZE"); poolSizeStr != "" {
		log.Printf("🔍 [DEBUG] MONGODB_POOL_SIZE found: %s", poolSizeStr)
		if poolSize, err := parseUint64(poolSizeStr); err == nil && poolSize > 0 {
			config.PoolSize = poolSize
			log.Printf("🔍 [DEBUG] PoolSize set to: %d", poolSize)
		}
	}

	log.Println("🔍 [DEBUG] getConfig() completed")
	return config
}

// Disconnect closes the MongoDB connection
func Disconnect() error {
	log.Println("🔍 [DEBUG] Disconnect() called")
	if Client == nil {
		log.Println("⚠️ [DEBUG] Client is nil, nothing to disconnect")
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	log.Println("🔍 [DEBUG] Disconnecting MongoDB client...")
	err := Client.Disconnect(ctx)
	if err != nil {
		log.Printf("❌ [DEBUG] Failed to disconnect: %v", err)
		return fmt.Errorf("failed to disconnect from MongoDB: %v", err)
	}

	connected = false
	log.Println("✅ [DEBUG] MongoDB connection closed")
	return nil
}

// IsConnected returns true if database is connected
func IsConnected() bool {
	log.Println("🔍 [DEBUG] IsConnected() called")
	if !connected || Client == nil {
		log.Println("⚠️ [DEBUG] Not connected (connected flag false or client nil)")
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	log.Println("🔍 [DEBUG] Pinging to verify connection...")
	err := Client.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Printf("⚠️ [DEBUG] Ping failed: %v", err)
		return false
	}
	log.Println("✅ [DEBUG] Connection verified")
	return true
}

// GetCollection returns a collection from the database
func GetCollection(collectionName string) *mongo.Collection {
	log.Printf("🔍 [DEBUG] GetCollection(%s) called", collectionName)
	if DB == nil {
		log.Printf("⚠️ [DEBUG] Database not initialized, attempting to connect...")
		if err := Connect(); err != nil {
			log.Printf("❌ [DEBUG] Failed to connect to database: %v", err)
			return nil
		}
	}

	collection := DB.Collection(collectionName)
	log.Printf("✅ [DEBUG] Collection %s retrieved", collectionName)
	return collection
}

// GetCollectionWithOptions returns a collection with custom options
func GetCollectionWithOptions(collectionName string, opts *options.CollectionOptions) *mongo.Collection {
	log.Printf("🔍 [DEBUG] GetCollectionWithOptions(%s) called", collectionName)
	if DB == nil {
		log.Printf("⚠️ [DEBUG] Database not initialized, attempting to connect...")
		if err := Connect(); err != nil {
			log.Printf("❌ [DEBUG] Failed to connect to database: %v", err)
			return nil
		}
	}

	collection := DB.Collection(collectionName, opts)
	log.Printf("✅ [DEBUG] Collection %s retrieved with options", collectionName)
	return collection
}

// WithTransaction executes a function within a transaction
func WithTransaction(ctx context.Context, fn func(sessCtx mongo.SessionContext) error) error {
	log.Println("🔍 [DEBUG] WithTransaction() called")
	if Client == nil {
		return fmt.Errorf("database client not initialized")
	}

	session, err := Client.StartSession()
	if err != nil {
		log.Printf("❌ [DEBUG] Failed to start session: %v", err)
		return fmt.Errorf("failed to start session: %v", err)
	}
	defer session.EndSession(ctx)

	err = session.StartTransaction()
	if err != nil {
		log.Printf("❌ [DEBUG] Failed to start transaction: %v", err)
		return fmt.Errorf("failed to start transaction: %v", err)
	}

	err = mongo.WithSession(ctx, session, func(sessCtx mongo.SessionContext) error {
		if err := fn(sessCtx); err != nil {
			return err
		}
		return session.CommitTransaction(sessCtx)
	})

	if err != nil {
		log.Printf("❌ [DEBUG] Transaction failed: %v", err)
		abortErr := session.AbortTransaction(ctx)
		if abortErr != nil {
			log.Printf("⚠️ [DEBUG] Failed to abort transaction: %v", abortErr)
		}
		return err
	}

	log.Println("✅ [DEBUG] Transaction completed successfully")
	return nil
}

// HealthCheck performs a health check on the database
func HealthCheck(ctx context.Context) error {
	log.Println("🔍 [DEBUG] HealthCheck() called")
	if Client == nil {
		return fmt.Errorf("database client not initialized")
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err := Client.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Printf("❌ [DEBUG] Health check failed: %v", err)
		return fmt.Errorf("database health check failed: %v", err)
	}

	log.Println("✅ [DEBUG] Health check passed")
	return nil
}

// GetDatabaseStats returns database statistics
func GetDatabaseStats(ctx context.Context) (map[string]interface{}, error) {
	log.Println("🔍 [DEBUG] GetDatabaseStats() called")
	if DB == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	cmd := bson.D{{Key: "dbStats", Value: 1}}
	var result bson.M

	err := DB.RunCommand(ctx, cmd).Decode(&result)
	if err != nil {
		log.Printf("❌ [DEBUG] Failed to get database stats: %v", err)
		return nil, fmt.Errorf("failed to get database stats: %v", err)
	}

	log.Println("✅ [DEBUG] Database stats retrieved")
	return result, nil
}

// GetCollectionStats returns statistics for a specific collection
func GetCollectionStats(ctx context.Context, collectionName string) (map[string]interface{}, error) {
	log.Printf("🔍 [DEBUG] GetCollectionStats(%s) called", collectionName)
	collection := GetCollection(collectionName)
	if collection == nil {
		return nil, fmt.Errorf("collection %s not found", collectionName)
	}

	count, err := collection.EstimatedDocumentCount(ctx)
	if err != nil {
		log.Printf("❌ [DEBUG] Failed to get document count: %v", err)
		return nil, fmt.Errorf("failed to get document count: %v", err)
	}

	stats := map[string]interface{}{
		"collection":     collectionName,
		"document_count": count,
		"database":       DB.Name(),
	}

	log.Printf("✅ [DEBUG] Collection stats retrieved: %d documents", count)
	return stats, nil
}

// CreateIndex creates an index on a collection
func CreateIndex(ctx context.Context, collectionName string, keys interface{}, opts *options.IndexOptions) error {
	log.Printf("🔍 [DEBUG] CreateIndex(%s) called", collectionName)
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
		log.Printf("❌ [DEBUG] Failed to create index: %v", err)
		return fmt.Errorf("failed to create index on %s: %v", collectionName, err)
	}

	log.Printf("✅ [DEBUG] Index created successfully on %s", collectionName)
	return nil
}

// DropCollection drops a collection from the database
func DropCollection(ctx context.Context, collectionName string) error {
	log.Printf("🔍 [DEBUG] DropCollection(%s) called", collectionName)
	collection := GetCollection(collectionName)
	if collection == nil {
		return fmt.Errorf("collection %s not found", collectionName)
	}

	err := collection.Drop(ctx)
	if err != nil {
		log.Printf("❌ [DEBUG] Failed to drop collection: %v", err)
		return fmt.Errorf("failed to drop collection %s: %v", collectionName, err)
	}

	log.Printf("✅ [DEBUG] Collection %s dropped", collectionName)
	return nil
}

// ListCollections returns all collection names in the database
func ListCollections(ctx context.Context) ([]string, error) {
	log.Println("🔍 [DEBUG] ListCollections() called")
	if DB == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	collections, err := DB.ListCollectionNames(ctx, bson.D{})
	if err != nil {
		log.Printf("❌ [DEBUG] Failed to list collections: %v", err)
		return nil, fmt.Errorf("failed to list collections: %v", err)
	}

	log.Printf("✅ [DEBUG] Found %d collections", len(collections))
	return collections, nil
}

// BulkInsert performs bulk insertion of documents
func BulkInsert(ctx context.Context, collectionName string, documents []interface{}) (*mongo.InsertManyResult, error) {
	log.Printf("🔍 [DEBUG] BulkInsert(%s, %d documents) called", collectionName, len(documents))
	collection := GetCollection(collectionName)
	if collection == nil {
		return nil, fmt.Errorf("collection %s not found", collectionName)
	}

	result, err := collection.InsertMany(ctx, documents)
	if err != nil {
		log.Printf("❌ [DEBUG] Bulk insert failed: %v", err)
		return nil, fmt.Errorf("failed to bulk insert into %s: %v", collectionName, err)
	}

	log.Printf("✅ [DEBUG] Bulk insert successful, inserted %d documents", len(result.InsertedIDs))
	return result, nil
}

// helper function to parse uint64
func parseUint64(s string) (uint64, error) {
	var n uint64
	_, err := fmt.Sscanf(s, "%d", &n)
	return n, err
}
