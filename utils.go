package main

import (
	"encoding/json"
	"fmt"
	"time"
)

func toJSON(v interface{}) (string, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return "", fmt.Errorf("error marshaling to JSON: %w", err)
	}
	return string(data), nil
}

func isExpired(expiration string) bool {
	expiryTime, err := time.Parse(time.RFC3339, expiration)
	if err != nil {
		logger.Printf("Error parsing time: %v", err)
		return true
	}
	return time.Now().After(expiryTime)
}
