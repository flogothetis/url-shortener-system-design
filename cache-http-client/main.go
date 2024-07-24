package main

import (
	"github.com/bradfitz/gomemcache/memcache"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

var cache *memcache.Client

func init() {
	cache = memcache.New("memcached:11211")
}

func main() {
	router := gin.Default()

	router.GET("/", getValue)
	router.POST("/", setValue)

	router.Run(":8080")
}

func getValue(c *gin.Context) {
	key := c.Query("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Please provide a 'key' parameter in the query"})
		return
	}

	log.Printf("Getting value for key: %s\n", key)

	item, err := cache.Get(key)
	if err != nil {
		log.Printf("Error getting value for key %s: %v\n", key, err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Key not found"})
		return
	}

	var value string
	if err != nil {
		log.Printf("Error decoding value for key %s: %v : %v\n", key, err, string(item.Value))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	log.Printf("Value found for key %s: %s\n", key, value)
	c.JSON(http.StatusOK, gin.H{key: string(item.Value)})
}

func setValue(c *gin.Context) {
	var data map[string]string
	if err := c.ShouldBindJSON(&data); err != nil {
		log.Printf("Error binding JSON: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Please provide both 'key' and 'value' in the request body"})
		return
	}

	key, ok1 := data["key"]
	value, ok2 := data["value"]
	if !ok1 || !ok2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Please provide both 'key' and 'value' in the request body"})
		return
	}

	log.Printf("Setting value for key %s: %s\n", key, value)

	item := &memcache.Item{
		Key:   key,
		Value: []byte(value),
	}
	if err := cache.Set(item); err != nil {
		log.Printf("Error setting value for key %s: %v\n", key, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	log.Printf("Value successfully set for key %s\n", key)
	c.JSON(http.StatusOK, gin.H{"success": true})
}
