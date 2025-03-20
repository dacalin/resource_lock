package _rl_redis

import (
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
	lock := New(host, port, "", "", 0, poolSize, "test")

	id := "testResource"

	lock.Lock(id)
	lock.Unlock(id)
	lock.Lock(id)
	lock.Unlock(id)
}

func TestConcurrentRedisLocking(t *testing.T) {
	host, port, poolSize := getRedisConfig()
	lock := New(host, port, "", "", 0, poolSize, "test")

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

func TestRedisResourceLockTime(t *testing.T) {
	host, port, poolSize := getRedisConfig()

	lock := New(host, port, "", "", 0, poolSize, "test")
	lockTime := 1000
	Instance().SetMaxLockTime(int64(lockTime))

	// Test parameters
	lockId := "TestRedisResourceLockTime2"

	startTime := time.Now()
	t.Log(startTime)
	lock.Lock(lockId)
	lock.Lock(lockId)

	unlockTime := time.Now()
	t.Log(unlockTime)

	lock.Unlock(lockId)

	if unlockTime.Sub(startTime) > time.Millisecond*time.Duration(lockTime+100) {
		t.Errorf("Unlock TTL expected to be %d, but got %d", lockTime, unlockTime.Sub(startTime))
	}
}
