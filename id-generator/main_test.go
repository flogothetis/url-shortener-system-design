// main_test.go

package main

import (
	"testing"
	"time"
)

func TestGenerateID(t *testing.T) {
	sf := NewSnowflake(1, 0)

	// Generate multiple IDs and check if they are unique
	idSet := make(map[int64]bool)
	for i := 0; i < 1000; i++ {
		id, err := sf.GenerateID()
		if err != nil {
			t.Errorf("Error generating ID: %v", err)
		}
		if idSet[id] {
			t.Errorf("Duplicate ID generated: %d", id)
		}
		idSet[id] = true
	}

}

func TestTilNextMillis(t *testing.T) {
	sf := NewSnowflake(1, 0)
	currentTimestamp := time.Now().UnixNano() / 1e6

	nextTimestamp := sf.tilNextMillis(currentTimestamp)
	if nextTimestamp <= currentTimestamp {
		t.Errorf("Expected next timestamp to be greater than current timestamp, got: %d", nextTimestamp)
	}
}
