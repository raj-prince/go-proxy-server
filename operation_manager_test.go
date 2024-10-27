package main

import (
	"testing"
)

func TestOperationManager(t *testing.T) {
	config := Config{
		RetryConfig: []RetryConfig{
			{Method: "XmlRead", RetryCount: 2, RetryInstruction: "retry_GET"},
			{Method: "JsonStat", SkipCount: 1, RetryCount: 1, RetryInstruction: "retry_STAT"},
		},
	}
	om := NewOperationManager(config)

	// Test GET operation
	if op := om.retrieveOperation("XmlRead"); op != "retry_GET" {
		t.Errorf("Expected 'retry_GET', got '%s'", op)
	}
	if op := om.retrieveOperation("XmlRead"); op != "retry_GET" {
		t.Errorf("Expected 'retry_GET', got '%s'", op)
	}
	if op := om.retrieveOperation("XmlRead"); op != "" {
		t.Errorf("Expected '', got '%s'", op)
	}

	// Test JsonStat operation
	if op := om.retrieveOperation("JsonStat"); op != "" {
		t.Errorf("Expected '', got '%s'", op)
	}
	if op := om.retrieveOperation("JsonStat"); op != "retry_STAT" {
		t.Errorf("Expected 'retry_POST', got '%s'", op)
	}
	if op := om.retrieveOperation("JsonStat"); op != "" {
		t.Errorf("Expected '', got '%s'", op)
	}

	// Test non-existent operation
	if op := om.retrieveOperation("JsonPut"); op != "" {
		t.Errorf("Expected '', got '%s'", op)
	}
}
