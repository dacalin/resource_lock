package _rl_redis

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	_i_resource_lock "github.com/dacalin/resource_lock/ports/resource_lock"
	"github.com/redis/go-redis/v9"
	"strings"
	"sync"
	"time"
)

var once sync.Once
var instance *RedisResourceLock
var sleepTime = 10 * time.Millisecond

type RedisResourceLock struct {
	_i_resource_lock.IResourceLock
	client        *redis.Client
	cleanMemMilis int64
	queuePrefix   string
}

func New(host string, port string, user string, password string, DB int, maxPoolSize int, prefix string) *RedisResourceLock {
	addr := fmt.Sprintf("%s:%s", host, port)
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		PoolSize: maxPoolSize,
		DB:       DB,
		Username: user,
		Password: password,
	})

	once.Do(func() {
		// Clean memory every 1 second by default
		instance = &RedisResourceLock{
			client:        client,
			cleanMemMilis: 15000,
			queuePrefix:   prefix + "_dlq_",
		}
		go instance.cleanMemLoop()
	})

	return instance
}

func Instance() *RedisResourceLock {
	if instance == nil {
		panic("Redis client is not set. Call New() first.")
	}

	return instance
}

func (self *RedisResourceLock) SetMaxLockTime(ms int64) {
	Instance().cleanMemMilis = ms
}
func generateUniqueID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}

func (self *RedisResourceLock) cleanMem() {
	ctx := context.Background()

	pattern := self.queuePrefix + "*"
	iter := self.client.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		queueKey := iter.Val()

		_, expTime, err := self.getNextQueueItem(ctx, queueKey)

		for err != nil && expTime.Before(time.Now()) {

			self.client.LPop(ctx, queueKey).Result()

			_, expTime, err = self.getNextQueueItem(ctx, queueKey)
		}

	}
	if err := iter.Err(); err != nil {
		fmt.Println("Error iterating keys:", err)
	}

}
func (self *RedisResourceLock) cleanMemLoop() {

	for {
		self.cleanMem()
		time.Sleep(time.Duration(self.cleanMemMilis) * time.Millisecond)
	}

}

func (self *RedisResourceLock) Lock(id string) {
	ctx := context.Background()
	queueKey := self.queuePrefix + id
	uniqueId := generateUniqueID()
	currentTime := time.Now()
	duration := time.Duration(self.cleanMemMilis) * time.Millisecond
	expirationTime := currentTime.Add(duration)
	compositeId := uniqueId + "|" + expirationTime.Format(time.RFC3339Nano)

	// Add the id to the queue
	self.client.RPush(ctx, queueKey, compositeId)

	for {
		// Check if it's the turn of this id
		nextCompositeId, _, err := self.getNextQueueItem(ctx, queueKey)
		if err == redis.Nil {
			fmt.Println("Error getting next queue item:", queueKey, "e: ", err)
			break
		}
		if err != nil {
			fmt.Println("Error getting next queue item:", queueKey, "e: ", err)
			time.Sleep(sleepTime)
			continue
		}

		if nextCompositeId == compositeId {
			return
		}

		// Wait for a while before retrying
		time.Sleep(sleepTime)
	}
}

func (self *RedisResourceLock) getNextQueueItem(ctx context.Context, queueKey string) (string, time.Time, error) {
	firstCompositeID, err := self.client.LIndex(ctx, queueKey, 0).Result()
	if err == nil {
		// Split string into uniqueID and expiration time
		split := strings.Split(firstCompositeID, "|")
		firstExpirationTime, err := time.Parse(time.RFC3339Nano, split[1])
		return firstCompositeID, firstExpirationTime, err
	}
	return "", time.Time{}, err
}

func (self *RedisResourceLock) Unlock(id string) {
	ctx := context.Background()
	queueKey := self.queuePrefix + id

	// Remove the first id from the queue
	_, err := self.client.LPop(ctx, queueKey).Result()

	if err != nil {
		fmt.Println("Error removing from queue:", err)
		return
	}

}
