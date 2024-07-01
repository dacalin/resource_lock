package _rl_redis

import (
	"context"
	"fmt"
	_i_resource_lock "github.com/dacalin/resource_lock/ports/resource_lock"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

var once sync.Once
var instance *RedisResourceLock

type RedisResourceLock struct {
	_i_resource_lock.IResourceLock
	client        *redis.Client
	cleanMemMilis int64
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
			cleanMemMilis: 1000}
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

func (self *RedisResourceLock) Lock(id string) {
	ctx := context.Background()
	result, err := self.client.SetNX(ctx, id, "locked", time.Duration(self.cleanMemMilis)*time.Millisecond).Result()
	if err != nil {
		fmt.Println("Error setting lock:", err)
	}
	if !result {
		fmt.Println("Failed to acquire lock for:", id)
	}
}

func (self *RedisResourceLock) Unlock(id string) {
	ctx := context.Background()
	_, err := self.client.Del(ctx, id).Result()
	if err != nil {
		fmt.Println("Error unlocking:", err)
	}
}
