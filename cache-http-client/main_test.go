package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/gin-gonic/gin"
)

func TestGetValueHandler_KeyNotFound(t *testing.T) {
	// Create a new HTTP request for the GET method with a query parameter "key"
	req, err := http.NewRequest("GET", "/?key=test", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a new gin context to simulate the request
	context, _ := gin.CreateTestContext(httptest.NewRecorder())
	context.Request = req

	// Call the handler function with the created context
	getValue(context)

	// Check the status code of the response
	if context.Writer.Status() != http.StatusNotFound {
		t.Errorf("expected status code %d; got %d", http.StatusNotFound, context.Writer.Status())
	}

	// Check the response body
	expectedBody := `{"error":"Key not found"}`
	if body := context.Writer.Body.String(); body != expectedBody {
		t.Errorf("expected response body %q; got %q", expectedBody, body)
	}
}

func TestSetValueHandler(t *testing.T) {
	// Create a JSON payload for the request body
	payload := map[string]string{"key": "test", "value": "example"}

	// Marshal the payload into JSON
	jsonPayload, err := json.Marshal(payload)
	assert.NoError(t, err)

	// Create a new HTTP request for the POST method with the JSON payload as the request body
	req, err := http.NewRequest("POST", "/", bytes.NewBuffer(jsonPayload))
	assert.NoError(t, err)

	// Set the content type header to application/json
	req.Header.Set("Content-Type", "application/json")

	// Create a response recorder to record the response
	rr := httptest.NewRecorder()

	// Call the handler function with the created request and response recorder
	setValue(rr, req)

	// Check the status code of the response
	assert.Equal(t, http.StatusOK, rr.Code)

	// Check the response body
	expectedBody := `{"success":true}`
	assert.Equal(t, expectedBody, rr.Body.String())
}
