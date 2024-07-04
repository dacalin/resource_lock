package rl

import (
	_rl_go "github.com/dacalin/resource_lock/adapters/resource_lock/go"
	_rl_redis "github.com/dacalin/resource_lock/adapters/resource_lock/redis"
	_i_resource_lock "github.com/dacalin/resource_lock/ports/resource_lock"
)

type LockType int

var instance _i_resource_lock.IResourceLock

const (
	Local LockType = iota
	Redis
)

type ResourceLockBuilder struct {
	lockType      LockType
	redisHost     string
	redisPort     string
	redisPoolSize int
	redisUser     string
	redisPassword string
	redisDB       int
	redisPrefix   string
	cleanMemMilis int64
}

func New(lockType LockType) *ResourceLockBuilder {
	return &ResourceLockBuilder{
		lockType:      lockType,
		cleanMemMilis: 1000, // Default clean memory interval
	}
}

func Instance() _i_resource_lock.IResourceLock {
	if instance == nil {
		panic("ResourceLock is not set. Call Build() first.")
	}

	return instance
}

func (b *ResourceLockBuilder) WithRedisConfig(host string, port string, user string, password string, DB int, maxPoolSize int, prefix string) *ResourceLockBuilder {
	b.redisHost = host
	b.redisPort = port
	b.redisPoolSize = maxPoolSize
	b.redisUser = user
	b.redisPassword = password
	b.redisDB = DB
	b.redisPrefix = prefix
	return b
}

func (b *ResourceLockBuilder) WithMaxLockTime(milis int64) *ResourceLockBuilder {
	b.cleanMemMilis = milis
	return b
}

func (b *ResourceLockBuilder) Build() _i_resource_lock.IResourceLock {
	switch b.lockType {
	case Local:
		itc := _rl_go.Instance()
		itc.SetMaxLockTime(b.cleanMemMilis)
		instance = itc
		return instance

	case Redis:
		itc := _rl_redis.New(b.redisHost, b.redisPort, b.redisUser, b.redisPassword, b.redisDB, b.redisPoolSize, b.redisPrefix)
		itc.SetMaxLockTime(b.cleanMemMilis)
		instance = itc
		return instance

	default:
		panic("Invalid lock type")
	}
}
