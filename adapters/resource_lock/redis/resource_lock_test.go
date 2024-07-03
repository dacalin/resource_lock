package _rl_redis

import (
	"fmt"
	"math/rand"
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

	id := "testResource"

	lock.Lock(id)
	lock.Unlock(id)
	lock.Lock(id)
	lock.Unlock(id)

}

func TestConcurrentRedisLocking(t *testing.T) {
	host, port, poolSize := getRedisConfig()
	lock := New(host, port, poolSize)

	id1 := "resource1"
	id2 := "resource2"

	var wg sync.WaitGroup

	wg.Add(3)

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
	go func() {
		defer wg.Done()
		lock.Lock(id2)
		time.Sleep(100 * time.Millisecond)
		lock.Unlock(id2)
	}()
	wg.Wait()

	t.Log("Concurrent locking and unlocking completed without panics or deadlocks.")
}

func TestRedisResourceLockBlock(t *testing.T) {
	host, port, poolSize := getRedisConfig()

	lock := New(host, port, poolSize)
	// Test parameters
	numGoroutines := 1000
	waitTime := 10 * time.Millisecond
	wg := sync.WaitGroup{}
	wg.Add(numGoroutines)

	// Channel to collect the order of lock acquisition
	lockOrder := make(chan int, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		time.Sleep(10 * time.Millisecond)
		go func(id int) {
			defer wg.Done()
			fmt.Printf("Goroutine %d trying to lock...\n", id)
			seed := time.Now().UnixNano()
			randomizer := rand.New(rand.NewSource(seed))
			resourceID := fmt.Sprintf("%d", randomizer.Intn(800))
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
