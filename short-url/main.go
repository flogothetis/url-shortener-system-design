package main

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"math/big"
	"net/http"
	"os"
	"time"
	"github.com/gin-contrib/cors"

	"encoding/json"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var logger = slog.Default()

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
func getUrlFromCache(shortURL string) (bool, string) {
	// Make a GET request to the endpoint responsible for checking short URL existence
	resp, err := http.Get("http://memcached_http_server-load-balancer/?key=" + shortURL)
	if err != nil {
		return false, "Get request to cache failed"
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, "Status not OK"
	}

	// Read the response body
	var response map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return false, "Failed to decode cache response"
	}
	value, ok := response[shortURL]
	if !ok {
		return false, "Cache response is invalid"
	}
	return true, value
}

func setValueInCache(key, value string) error {
	// Prepare JSON payload
	payload := map[string]string{
		"key":   key,
		"value": value,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	// Make a POST request to the cache server endpoint
	resp, err := http.Post("http://memcached_http_server-load-balancer/", "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("failed to set value in cache: " + resp.Status)
	}

	return nil
}

func main() {

	r := gin.Default()

	r.Use(cors.Default())

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
		if err = setValueInCache(shortURL, input.OriginalURL); err != nil {
			logger.Info("Failed to store key in cache  %s\n", err.Error())
		}

		c.JSON(http.StatusOK, gin.H{"shortUrl": shortURL})
	})

	r.GET("/:shortID", func(c *gin.Context) {
		shortID := c.Param("shortID")

		var urlEntry URL
		var originalURL string
		var isOk bool
		if isOk, originalURL = getUrlFromCache(shortID); !isOk {
			logger.Warn("Failed to get key " + shortID + " from cache. Request will be forwared to database\n")
			// Find URL in MongoDB
			ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

			err := urlCollection.FindOne(ctx, bson.M{"_id": shortID}).Decode(&urlEntry)

			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "URL not found in database"})
				return
			}
			originalURL = urlEntry.OriginalURL
		}

		c.Redirect(http.StatusTemporaryRedirect, originalURL)
	})

	// Run the server
	if err := r.Run(":3001"); err != nil {
		slog.Info(err.Error())
	}

}
