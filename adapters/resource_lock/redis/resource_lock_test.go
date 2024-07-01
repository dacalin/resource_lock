package _rl_redis

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"sync"
	"testing"
	"time"
)

func getRedisConfig() (string, string, int) {
	host := os.Getenv("REDIS_HOST")
	if host == "" {
		host = "localhost"
	}

	port := os.Getenv("REDIS_PORT")
	if port == "" {
		port = "6379"
	}

	poolSize := 10
	if val := os.Getenv("REDIS_POOL_SIZE"); val != "" {
		poolSize = atoi(val)
	}

	return host, port, poolSize
}

func atoi(str string) int {
	value, err := strconv.Atoi(str)
	if err != nil {
		return 10 // default value if conversion fails
	}
	return value
}

func TestRedisResourceLock(t *testing.T) {
	host, port, poolSize := getRedisConfig()
	lock := New(host, port, poolSize)
	lock.SetMaxLockTime(200) // 200 milliseconds

	id := "testResource"

	// Lock the resource
	lock.Lock(id)

	ctx := context.Background()
	exists, err := lock.client.Exists(ctx, id).Result()
	if err != nil {
		t.Fatal("Error checking existence of key:", err)
	}

	if exists == 0 {
		t.Error("Expected resource to be locked, but it does not exist.")
	}

	// Unlock the resource
	lock.Unlock(id)
	lock.Lock(id)
	lock.Unlock(id)

	exists, err = lock.client.Exists(ctx, id).Result()
	if err != nil {
		t.Fatal("Error checking existence of key:", err)
	}

	if exists != 0 {
		t.Error("Expected resource to be unlocked, but it still exists.")
	}
}

func TestConcurrentRedisLocking(t *testing.T) {
	host, port, poolSize := getRedisConfig()
	lock := New(host, port, poolSize)
	lock.SetMaxLockTime(200) // 200 milliseconds

	id1 := "resource1"
	id2 := "resource2"

	var wg sync.WaitGroup

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

	t.Log("Concurrent locking and unlocking completed without panics or deadlocks.")
}

func TestRedisMaxTime(t *testing.T) {
	host, port, poolSize := getRedisConfig()
	lock := New(host, port, poolSize)
	lock.SetMaxLockTime(100) // 100 milliseconds

	id := "testResourceLoop"

	lock.Lock(id)

	ctx := context.Background()
	exists, err := lock.client.Exists(ctx, id).Result()
	if err != nil {
		t.Fatal("Error checking existence of key:", err)
	}

	if exists == 0 {
		t.Error("Expected resource to be present after locking, but it does not exist.")
	}

	time.Sleep(300 * time.Millisecond)

	exists, err = lock.client.Exists(ctx, id).Result()
	if err != nil {
		t.Fatal("Error checking existence of key:", err)
	}

	if exists != 0 {
		t.Error("Expected resource to be removed by cleanMemLoop, but it still exists.")
	} else {
		t.Log("Resource successfully removed by cleanMemLoop.")
	}

	lock.Unlock(id)

}

func TestSetMaxLockTime(t *testing.T) {
	host, port, poolSize := getRedisConfig()
	lock := New(host, port, poolSize)

	lock.SetMaxLockTime(500) // 500 milliseconds

	if lock.cleanMemMilis != 500 {
		t.Errorf("Expected cleanMemMilis to be 500, got %d", lock.cleanMemMilis)
	}
}

func TestRedisResourceLockBlock(t *testing.T) {
	host := "localhost"
	port := "6379"
	maxPoolSize := 10

	lock := New(host, port, maxPoolSize)

	// Test parameters
	resourceID := "test_resource"
	numGoroutines := 5
	waitTime := 200 * time.Millisecond
	wg := sync.WaitGroup{}
	wg.Add(numGoroutines)

	// Channel to collect the order of lock acquisition
	lockOrder := make(chan int, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			fmt.Printf("Goroutine %d trying to lock...\n", id)
			lock.Lock(resourceID)
			fmt.Printf("Goroutine %d acquired the lock!\n", id)
			lockOrder <- id
			time.Sleep(waitTime) // Simulate some work with the resource
			lock.Unlock(resourceID)
			fmt.Printf("Goroutine %d released the lock!\n", id)
		}(i)
	}

	wg.Wait()
	close(lockOrder)

	// Check the order of lock acquisition
	order := []int{}
	for id := range lockOrder {
		order = append(order, id)
	}

	// Ensure the order is as expected
	for i := 0; i < numGoroutines; i++ {
		if order[i] != i {
			t.Errorf("Expected goroutine %d to acquire the lock at position %d, but got position %d", i, i, order[i])
		}
	}
}
