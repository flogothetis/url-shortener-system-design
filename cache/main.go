package main

import (
	"github.com/bradfitz/gomemcache/memcache"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"os"
)

func main() {
	r := gin.Default()

	
	memcachedClient := memcache.New(os.Getenv("ME_CONFIG_MEMCACHE_URL"))
	// r.GET("/getAll", func(c *gin.Context) {
	// 	keys, err := memcachedClient.FlushAll() // Passing empty string fetches all items
	// 	if err != nil {
	// 		log.Printf("Error getting all key-value pairs from Memcached: %v\n", err)
	// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	// 		return
	// 	}

	// 	var keyValues []gin.H
	// 	for _, key := range keys {
	// 		item, err := memcachedClient.Get(key)
	// 		if err != nil {
	// 			log.Printf("Error getting item with key %s from Memcached: %v\n", key, err)
	// 			continue
	// 		}
	// 		keyValues = append(keyValues, gin.H{"key": key, "value": string(item.Value)})
	// 	}
	// 	log.Printf("Retrieved all key-value pairs from Memcached: %v\n", keyValues)
	// 	c.JSON(http.StatusOK, keyValues)
	// })

	r.GET("/get/:key", func(c *gin.Context) {
		key := c.Param("key")
		item, err := memcachedClient.Get(key)
		if err != nil {
			log.Printf("Error getting value for key %s from Memcached: %v\n", key, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		log.Printf("Retrieved value for key %s from Memcached: %s\n", key, string(item.Value))
		c.JSON(http.StatusOK, gin.H{"value": string(item.Value)})
	})

	r.POST("/set", func(c *gin.Context) {
		var data struct {
			Key   string `json:"key"`
			Value string `json:"value"`
		}
		if err := c.ShouldBindJSON(&data); err != nil {

			log.Printf("Error binding JSON data: %v\n", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		item := &memcache.Item{
			Key:   data.Key,
			Value: []byte(data.Value),
		}

		if err := memcachedClient.Set(item); err != nil {
			log.Printf("Error setting value for key %s in Memcached: %v\n", data.Key, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		log.Printf("Set value for key %s in Memcached successfully\n", data.Key)
		c.JSON(http.StatusOK, gin.H{"message": "Value set successfully"})
	})

	// r.DELETE("/delete/:key", func(c *gin.Context) {
	//     key := c.Param("key")
	//     if err := memcachedClient.Delete(key); err != nil {
	//         c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	//         return
	//     }
	//     c.JSON(http.StatusOK, gin.H{"message": "Key deleted successfully"})
	// })

	port := os.Getenv("CACHE_API_PORT")
	if port == "" {
		port = "8080"
	}
	if err := r.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}

