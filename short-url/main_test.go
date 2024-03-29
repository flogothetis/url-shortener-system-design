package main

import (
	"context"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"net/http"
	"net/http/httptest"
	"testing"
)

// MockMongoDB is a mock implementation of MongoDB for testing
func MockMongoDB() *mongo.Database {
	client, _ := mongo.Connect(context.Background(), options.Client().ApplyURI("mongodb://localhost:27017"))
	return client.Database("test")
}

func TestRedirectHandler(t *testing.T) {
	// Create a test router
	r := gin.Default()

	// Mock the MongoDB collection
	urlCollectionMock := MockMongoDB().Collection("urls")

	// Mock a document in the collection
	_, err := urlCollectionMock.InsertOne(context.Background(), bson.M{"_id": "gomSgBExzM", "originalUrl": "http://example.com"})
	if err != nil {
		t.Fatalf("Failed to insert mock document into collection: %v", err)
	}

	// Perform a GET request to the redirection endpoint
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/gomSgBExzM", nil)

	t.Logf("Making request: %s %s", req.Method, req.URL)

	// Serve the request using the test router
	r.ServeHTTP(w, req)

	t.Logf("Received response: %d %s", w.Code, w.Body.String())

	// Check the response status code (it should be a redirect status code)
	if w.Code != http.StatusTemporaryRedirect {
		t.Errorf("Expected status code %d but got %d", http.StatusTemporaryRedirect, w.Code)
	}
}
