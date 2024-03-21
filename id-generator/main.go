// main.go

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
	"errors"
)

// Snowflake struct
type Snowflake struct {
	workerID          int
	sequence          int
	epoch             int64
	workerIDBits      uint
	sequenceBits      uint
	maxWorkerID       int
	sequenceMask      int
	workerIDShift     uint
	timestampLeftShift uint
	mu                sync.Mutex
	lastTimestamp     int64
}

// NewSnowflake initializes a new Snowflake instance
func NewSnowflake(workerID int, epoch int64) *Snowflake {
	return &Snowflake{
		workerID:          workerID,
		sequence:          0,
		epoch:             epoch,
		workerIDBits:      10,
		sequenceBits:      12,
		maxWorkerID:       int(-1 ^ (-1 << 10)),
		sequenceMask:      int(-1 ^ (-1 << 12)),
		workerIDShift:     12,
		timestampLeftShift: 22,
		lastTimestamp:     -1,
	}
}

// GenerateID generates a unique ID
func (sf *Snowflake) GenerateID() (int64, error) {
	sf.mu.Lock()
	defer sf.mu.Unlock()

	timestamp := time.Now().UnixNano() / 1e6
	if timestamp < sf.lastTimestamp {
		return 0, errors.New("Clock moved backwards")
	}

	if timestamp == sf.lastTimestamp {
		sf.sequence = (sf.sequence + 1) & sf.sequenceMask
		if sf.sequence == 0 {
			timestamp = sf.tilNextMillis(sf.lastTimestamp)
		}
	} else {
		sf.sequence = 0
	}

	sf.lastTimestamp = timestamp
	id := int64(0) | (int64((timestamp - sf.epoch)) << sf.timestampLeftShift) | (int64(sf.workerID) << sf.workerIDShift) | int64(sf.sequence)

	return id, nil
}

// tilNextMillis waits until the next millisecond
func (sf *Snowflake) tilNextMillis(lastTimestamp int64) int64 {
	timestamp := time.Now().UnixNano() / 1e6
	for timestamp <= lastTimestamp {
		timestamp = time.Now().UnixNano() / 1e6
	}
	return timestamp
}

// getTimeHandler handles requests to get the current time
func getTimeHandler(w http.ResponseWriter, r *http.Request) {
	id, err := snowflake.GenerateID()
	if err != nil {
		http.Error(w, fmt.Sprintf("Internal Server Error: %v", err), http.StatusInternalServerError)
		return
	}
	response := map[string]int64{"time": id}
	http.Header.Add(w.Header(), "content-type", "application/json")
	json.NewEncoder(w).Encode(response)
}

var snowflake *Snowflake

func main() {
	workerID, err := strconv.Atoi(os.Getenv("WORKER_ID"))
	if err != nil {
		panic("Invalid WORKER_ID")
	}

	epoch, err := strconv.ParseInt(os.Getenv("EPOCH"), 10, 64)
	if err != nil {
		panic("Invalid EPOCH")
	}

	// Initialize Snowflake with worker ID and epoch time
	snowflake = NewSnowflake(workerID, epoch)

	// Setup routes
	r := http.NewServeMux()
	r.HandleFunc("/getTime", getTimeHandler)

	// Start HTTP server
	server := &http.Server{
		Addr:    ":5001",
		Handler: r,
	}

	fmt.Printf("Server with Worker ID %d and Epoch %d is running on :5001\n", workerID, epoch)
	server.ListenAndServe()
}