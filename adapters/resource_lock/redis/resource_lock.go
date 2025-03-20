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
	client      *redis.Client
	TTL         int64
	queuePrefix string
	uniqueId    map[string]string
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
			client:      client,
			TTL:         15000,
			queuePrefix: prefix + "_dlq_",
			uniqueId:    make(map[string]string),
		}
	})

	return instance
}

func Instance() *RedisResourceLock {
	if instance == nil {
		panic("Redis client is not set. Call New() first.")
	}

	return instance
}

func (l *RedisResourceLock) SetMaxLockTime(ms int64) {
	Instance().TTL = ms
}

func generateUniqueID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}

func (l *RedisResourceLock) LockWithTTL(id string, ms int64) {
	ctx := context.Background()
	key := l.queuePrefix + id
	uniqueId := generateUniqueID()

	var sleepTime time.Duration

	for {
		acquired := l.client.SetNX(ctx, key, uniqueId, time.Millisecond*time.Duration(ms))

		if acquired.Val() == true {
			l.uniqueId[id] = uniqueId
			return
		}

		sleepTime = time.Millisecond * time.Duration(1)
		time.Sleep(sleepTime)
	}
}

func (l *RedisResourceLock) TryLockWithTTL(id string, ms int64) bool {
	ctx := context.Background()
	key := l.queuePrefix + id
	uniqueId := generateUniqueID()

	acquired := l.client.SetNX(ctx, key, uniqueId, time.Millisecond*time.Duration(ms))

	if acquired.Val() == true {
		l.uniqueId[id] = uniqueId
		return true
	}

	return false
}

func (l *RedisResourceLock) TryLock(id string) bool {
	l.TryLockWithTTL(id, l.TTL)
}

func (l *RedisResourceLock) Lock(id string) {
	l.LockWithTTL(id, l.TTL)
}

func (l *RedisResourceLock) Unlock(id string) {
	ctx := context.Background()
	key := l.queuePrefix + id

	if l.client.Get(ctx, key).Val() == l.uniqueId[id] {
		l.client.Del(ctx, key)
	}
}
