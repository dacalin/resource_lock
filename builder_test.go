package rl

import (
	"fmt"
	"testing"
	"time"
)

func TestBuilder(t *testing.T) {
	rl := New(Local).
		WithMaxLockTime(30000). // Set clean memory interval to 5000 milliseconds
		Build()

	resourceID := "exampleResource"
	rl.Lock(resourceID)
	fmt.Println("Locked resource with Local Lock")

	// Simulate some work with the resource
	time.Sleep(500 * time.Millisecond)

	rl.Unlock(resourceID)
	fmt.Println("Unlocked resource with Local Lock")

	Instance().Lock(resourceID)
	Instance().Unlock(resourceID)
}
