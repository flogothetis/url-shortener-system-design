package main

import (
	"context"
	"math/big"
	"net/http"
	"os"
	"time"

	"log/slog"

	"encoding/json"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// URL model
type URL struct {
	ID          string    `json:"id" bson:"_id"`
	OriginalURL string    `json:"originalUrl" bson:"originalUrl" binding:"required,url"`
	CreatedAt   time.Time `json:"createdAt" bson:"createdAt"`
}

//TODO:: Add cache (MEMCACHE - READ HEAVY)

var (
	base58Alphabet = []byte("123456789ABCDEFJKLMGHNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz")
)

func base58Encode(input int64) string {
	x := big.NewInt(input)
	base := big.NewInt(58)
	zero := big.NewInt(0)
	encoded := make([]byte, 0, 11) // log58(2^64) ~= 11

	for x.Cmp(zero) > 0 {
		mod := new(big.Int)
		x.DivMod(x, base, mod)
		encoded = append([]byte{base58Alphabet[mod.Int64()]}, encoded...)
	}

	return string(encoded)
}

func main() {

	logger := slog.Default()
	r := gin.Default()

	// Read MongoDB host from environment variable
	mongoHost := os.Getenv("MONGO_HOST")
	if mongoHost == "" {
		logger.Error("MONGO_HOST environment variable not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second) // Change the duration as needed

	mongoClient, err := mongo.Connect(
		ctx,
		options.Client().ApplyURI(mongoHost),
	)

	defer func() {
		cancel()
		if err := mongoClient.Disconnect(ctx); err != nil {
			logger.Error("mongodb disconnect error : %v", err)
		}
	}()

	if err != nil {
		logger.Error("connection error :%v", err)
		return
	}

	err = mongoClient.Ping(ctx, readpref.Primary())
	if err != nil {
		logger.Error("ping mongodb error :%v", err)
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

		// Call id-generator-load-balancer/getTime to fetch a unique time-based ID
		idResp, err := http.Get("http://load-balancer-id-generators/getTime")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer idResp.Body.Close()

		var idResponse map[string]int64
		if err := json.NewDecoder(idResp.Body).Decode(&idResponse); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode ID response"})
			return
		}

		timeID := idResponse["time"]

		// Convert ID to base58
		shortURL := base58Encode(timeID)

		// Create URL entry
		urlEntry := URL{
			ID:          shortURL,
			OriginalURL: input.OriginalURL,
			CreatedAt:   time.Now(),
		}

		ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
		_, err = urlCollection.InsertOne(ctx, urlEntry)
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
		ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

		err := urlCollection.FindOne(ctx, bson.M{"_id": shortID}).Decode(&urlEntry)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "URL not found"})
			return
		}

		c.Redirect(http.StatusTemporaryRedirect, urlEntry.OriginalURL)
	})

	// Run the server
	if err := r.Run(":3000"); err != nil {
		slog.Info(err.Error())
	}

}
