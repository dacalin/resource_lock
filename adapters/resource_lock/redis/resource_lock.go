package _rl_redis

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	_i_resource_lock "github.com/dacalin/resource_lock/ports/resource_lock"
	"github.com/redis/go-redis/v9"
	"sync"
	"time"
)

var once sync.Once
var instance *RedisResourceLock

type RedisResourceLock struct {
	_i_resource_lock.IResourceLock
	client        *redis.Client
	cleanMemMilis int64
	queuePrefix   string
	lockPrefix    string
}

func New(host string, port string, maxPoolSize int) *RedisResourceLock {
	addr := fmt.Sprintf("%s:%s", host, port)
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		PoolSize: maxPoolSize,
	})

	once.Do(func() {
		// Clean memory every 1 second by default
		instance = &RedisResourceLock{
			client:        client,
			cleanMemMilis: 1000,
			queuePrefix:   "dlq_",
			lockPrefix:    "dlk_",
		}
	})

	return instance
}

func Instance() _i_resource_lock.IResourceLock {
	if instance == nil {
		panic("Redis client is not set. Call New() first.")
	}

	return instance
}

func (self *RedisResourceLock) SetMaxLockTime(milis int64) {
	instance.cleanMemMilis = milis
}

func generateUniqueID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}

func (self *RedisResourceLock) Lock(id string) {
	ctx := context.Background()
	lockKey := self.lockPrefix + id
	queueKey := self.queuePrefix + id

	uniqueID := generateUniqueID()
	// Add the id to the queue
	self.client.RPush(ctx, queueKey, uniqueID)

	for {
		// Check if it's the turn of this id
		firstID, err := self.client.LIndex(ctx, queueKey, 0).Result()
		if err != nil {
			fmt.Println("Error checking queue:", err)
			time.Sleep(100 * time.Millisecond)
			continue
		}

		if firstID == uniqueID {
			// Try to acquire the lock
			result, errSet := self.client.SetNX(ctx, lockKey, uniqueID, time.Duration(self.cleanMemMilis)*time.Millisecond).Result()
			if errSet != nil {
				fmt.Println("Error setting lock:", err)
				time.Sleep(100 * time.Millisecond)
				continue
			}

			if result {
				// Lock acquired, break the loop
				return
			}
		}

		// Wait for a while before retrying
		time.Sleep(100 * time.Millisecond)
	}
}

func (self *RedisResourceLock) Unlock(id string) {
	ctx := context.Background()
	lockKey := self.lockPrefix + id
	queueKey := self.queuePrefix + id

	// Ensure only the current lock holder can unlock
	_, err := self.client.Get(ctx, lockKey).Result()
	if err != nil && err != redis.Nil {
		fmt.Println("Error getting current lock holder:", err)
		return
	}

	// Release the lock
	_, err = self.client.Del(ctx, lockKey).Result()
	if err != nil {
		fmt.Println("Error unlocking:", err)
		return
	}

	// Remove the first id from the queue
	_, err = self.client.LPop(ctx, queueKey).Result()
	if err != nil {
		fmt.Println("Error removing from queue:", err)
		return
	}
}
