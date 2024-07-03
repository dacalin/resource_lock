package _rl_go

import (
	"sync"
	"testing"
	"time"
)

func TestGoResourceLock(t *testing.T) {
	// Set a custom max lock time for testing
	// Create a lock instance
	lock := Instance()

	// Define a resource ID
	id := "testResource"

	// Lock the resource
	lock.Lock(id)
	t.Log("Resource locked.")

	// Unlock the resource and verify no panic occurs
	lock.Unlock(id)
	t.Log("Resource unlocked.")

	// Lock the resource again to ensure it can be re-locked
	lock.Lock(id)
	t.Log("Resource re-locked.")

	// Ensure Unlock does not panic when unlocking an unlocked resource
	lock.Unlock(id)
	t.Log("Resource unlocked again.")

	// Test cleanMem to ensure old locks are removed
	lock.Lock(id)
	time.Sleep(300 * time.Millisecond) // Wait for the lock to expire

	_, exists := lock.resource.Load(id)
	if exists {
		t.Error("Expected resource to be removed after cleanMem, but it still exists.")
	} else {
		t.Log("Resource successfully removed after cleanMem.")
	}
}

func TestConcurrentLocking(t *testing.T) {
	// Set a custom max lock time for testing
	// Create a lock instance
	lock := Instance()

	// Define resource IDs
	id1 := "resource1"
	id2 := "resource2"

	// WaitGroup to wait for all goroutines to finish
	var wg sync.WaitGroup

	// Test concurrent locking and unlocking
	wg.Add(2)

	go func() {
		defer wg.Done()
		lock.Lock(id1)
		time.Sleep(100 * time.Millisecond)
		lock.Unlock(id1)
	}()

	go func() {
		defer wg.Done()
		lock.Lock(id2)
		time.Sleep(100 * time.Millisecond)
		lock.Unlock(id2)
	}()

	wg.Wait()

	// Ensure no panics or deadlocks occurred
	t.Log("Concurrent locking and unlocking completed without panics or deadlocks.")
}

func TestCleanMemLoop(t *testing.T) {
	// Set a custom max lock time for testing
	// Create a lock instance
	lock := Instance()

	// Define a resource ID
	id := "testResourceLoop"

	// Lock the resource
	lock.Lock(id)

	// Ensure the resource is initially present
	_, exists := lock.resource.Load(id)
	if !exists {
		t.Error("Expected resource to be present after locking, but it does not exist.")
	}

	// Wait long enough for cleanMemLoop to clean the resource
	time.Sleep(300 * time.Millisecond)

	// Ensure the resource is removed by cleanMemLoop
	_, exists = lock.resource.Load(id)
	if exists {
		t.Error("Expected resource to be removed by cleanMemLoop, but it still exists.")
	} else {
		t.Log("Resource successfully removed by cleanMemLoop.")
	}
}
