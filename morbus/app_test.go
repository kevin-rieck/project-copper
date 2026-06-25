package main

import (
	"context"
	"testing"
)

func TestAppEngineDelegation(t *testing.T) {
	app := NewApp()
	app.startup(context.Background())

	err := app.AddConnection("conn1", "tcp://192.168.1.50:502")
	if err != nil {
		t.Fatalf("AddConnection failed: %v", err)
	}

	// Verify delegation occurred by ensuring no crash and error is nil
	// Since engine encapsulates its state, we just check if it propagates correctly
	err = app.AddDevice("dev1", "conn1", 1)
	if err != nil {
		t.Fatalf("AddDevice failed: %v", err)
	}
}
