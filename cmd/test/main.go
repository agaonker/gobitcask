package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/ashish/gobitcask/bitcask"
)

func main() {
	// Create a new Bitcask instance with debug mode enabled
	bc, err := bitcask.New("testdata", nil)
	if err != nil {
		log.Fatalf("Failed to create Bitcask: %v", err)
	}
	defer bc.Close()

	// Test 1: Store and retrieve simple string
	fmt.Println("\nTest 1: Simple string operations")
	err = bc.Put("greeting", "Hello, World!")
	if err != nil {
		log.Printf("Failed to put greeting: %v", err)
	}

	value, err := bc.Get("greeting")
	if err != nil {
		log.Printf("Failed to get greeting: %v", err)
	} else {
		fmt.Printf("Retrieved greeting: %v\n", value)
	}

	// Test 2: Store and retrieve JSON object
	fmt.Println("\nTest 2: JSON object operations")
	user := map[string]interface{}{
		"name":    "John Doe",
		"age":     30,
		"email":   "john@example.com",
		"active":  true,
		"tags":    []string{"user", "admin"},
		"created": time.Now().Unix(),
	}

	err = bc.Put("user:123", user)
	if err != nil {
		log.Printf("Failed to put user: %v", err)
	}

	value, err = bc.Get("user:123")
	if err != nil {
		log.Printf("Failed to get user: %v", err)
	} else {
		// Pretty print the JSON
		jsonData, _ := json.MarshalIndent(value, "", "  ")
		fmt.Printf("Retrieved user:\n%s\n", jsonData)
	}

	// Test 3: List all keys
	fmt.Println("\nTest 3: List all keys")
	keys := bc.ListKeys()
	fmt.Printf("Found %d keys:\n", len(keys))
	for _, key := range keys {
		fmt.Printf("  - %s\n", key)
	}

	// Test 4: Delete operation
	fmt.Println("\nTest 4: Delete operation")
	err = bc.Delete("greeting")
	if err != nil {
		log.Printf("Failed to delete greeting: %v", err)
	}

	// Try to get deleted key
	value, err = bc.Get("greeting")
	if err != nil {
		fmt.Printf("Expected error for deleted key: %v\n", err)
	}

	// Test 5: Store and retrieve multiple values
	fmt.Println("\nTest 5: Multiple values")
	for i := 1; i <= 5; i++ {
		key := fmt.Sprintf("counter:%d", i)
		value := map[string]interface{}{
			"number": i,
			"time":   time.Now().UnixNano(),
		}
		err = bc.Put(key, value)
		if err != nil {
			log.Printf("Failed to put %s: %v", key, err)
		}
	}

	// List keys again
	keys = bc.ListKeys()
	fmt.Printf("\nFinal key count: %d\n", len(keys))
	for _, key := range keys {
		fmt.Printf("  - %s\n", key)
	}

	// Test 6: Clear all data
	fmt.Println("\nTest 6: Clear all data")
	err = bc.Clear()
	if err != nil {
		log.Printf("Failed to clear data: %v", err)
	}

	// Verify clear
	keys = bc.ListKeys()
	fmt.Printf("Keys after clear: %d\n", len(keys))
}
