package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"log/slog"
	"github.com/google/uuid"
)

// URL model
type URL struct {
	ID          string    `json:"id" bson:"_id"`
	OriginalURL string    `json:"originalUrl" bson:"originalUrl" binding:"required,url"`
	ShortURL    string    `json:"shortUrl" bson:"shortUrl"`
	CreatedAt   time.Time `json:"createdAt" bson:"createdAt"`
}


func main() {
	opts := &slog.HandlerOptions{
        Level: slog.LevelDebug,
    }

    handler := slog.NewJSONHandler(os.Stdout, opts)

    logger := slog.New(handler)
	r := gin.Default()

	// Read MongoDB host from environment variable
	mongoHost := os.Getenv("MONGO_HOST")
	if mongoHost == "" {
		logger.Info("MONGO_HOST environment variable not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second) // Change the duration as needed

	mongoClient, err := mongo.Connect(
	 ctx,
	 options.Client().ApplyURI("mongodb://mongodb:27017/"),
	)
   
	defer func() {
	 cancel()
	 if err := mongoClient.Disconnect(ctx); err != nil {
		logger.Info("mongodb disconnect error : %v", err)
	 }
	}()
   
	if err != nil {
		logger.Info("connection error :%v", err)
	 return
	}
   
	err = mongoClient.Ping(ctx, readpref.Primary())
	if err != nil {
		logger.Info("ping mongodb error :%v", err)
	 return
	}
	logger.Warn("ping success")

	db := mongoClient.Database("urlshortener")
	urlCollection := db.Collection("urls")

	// Define routes
	r.POST("/shorten", func(c *gin.Context) {
		var input struct {
			OriginalURL string `json:"originalUrl" binding:"required,url"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Generate short ID using base64 encoding of UUID
		shortID := base64.URLEncoding.EncodeToString([]byte(uuid.New().String())[:12])

		shortURL := fmt.Sprintf("%s/%s", c.Request.Host, shortID)

		// Create URL entry
		urlEntry := URL{
			ID:          shortID,
			OriginalURL: input.OriginalURL,
			ShortURL:    shortURL,
			CreatedAt:   time.Now(),
		}
		ctx, _ := context.WithTimeout(context.Background(), 30*time.Second) // Change the duration as needed

		_, err := urlCollection.InsertOne(ctx, urlEntry)
		if err != nil {
			logger.Info("Failed to insert into MongoDB: %v\n", err)
		   c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to shorten URL"})
		   return
		}

		c.JSON(http.StatusOK, gin.H{"shortUrl": shortURL})
	})

	r.GET("/:shortID", func(c *gin.Context) {
		shortID := c.Param("shortID")

		var urlEntry URL

		// Find URL in MongoDB
		err := urlCollection.FindOne(ctx, bson.M{"_id": shortID}).Decode(&urlEntry)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "URL not found"})
			return
		}

		c.Redirect(http.StatusTemporaryRedirect, urlEntry.OriginalURL)
	})

	// Run the server
	if err := r.Run(":3000"); err != nil {
		//slog.Info(err)
	}
}
