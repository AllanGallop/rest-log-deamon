package main

import (
    "os"
    "context"
    "encoding/json"
    "fmt"
    "log"
    "net/http"

    "github.com/gin-gonic/gin"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

var collection *mongo.Collection

func main() {
    // Retrieve MongoDB URI from environment variable
    mongoURI := os.Getenv("MONGO_URI")
    if mongoURI == "" {
        log.Fatal("MONGO_URI environment variable is not set")
    }

    // MongoDB connection options
    clientOptions := options.Client().ApplyURI(mongoURI)

    // Connect to MongoDB
    client, err := mongo.Connect(context.TODO(), clientOptions)
    if err != nil {
        log.Fatal(err)
    }
    err = client.Ping(context.TODO(), nil)
    if err != nil {
        log.Fatal("MongoDB not reachable:", err)
    }

    collection = client.Database("logDB").Collection("logs")
    fmt.Println("Connected to MongoDB!")

    // Create TTL Index on timestamp field
    indexModel := mongo.IndexModel{
        Keys: bson.M{"timestamp": 1},
        Options: options.Index().SetExpireAfterSeconds(15552000), // 6 months
    }
    _, err = collection.Indexes().CreateOne(context.TODO(), indexModel)
    if err != nil {
        log.Fatal("Could not create TTL index:", err)
    }

    // HTTP server setup
    router := gin.Default()
    router.POST("/log", handleLogPost)

    router.Run(":8888") // Start the server
}


func handleLogPost(c *gin.Context) {
    var logEntry struct {
        Timestamp string          `json:"timestamp"`
        Issuer    string          `json:"issuer"`
        Level     string          `json:"level"`
        Type      string          `json:"type"`
        Data      json.RawMessage `json:"data"`
    }

    // Bind the JSON payload to the struct
    if err := c.ShouldBindJSON(&logEntry); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // Prepare the document to insert into MongoDB
    logDoc := bson.M{
        "timestamp": logEntry.Timestamp,
        "issuer":    logEntry.Issuer,
        "level":     logEntry.Level,
        "type":      logEntry.Type,
        "data":      logEntry.Data,
    }

    // Insert the log document into the MongoDB collection
    _, err := collection.InsertOne(context.TODO(), logDoc)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert log entry"})
        return
    }

    // Respond with a success message
    c.JSON(http.StatusOK, gin.H{"message": "Log entry saved successfully"})
}
