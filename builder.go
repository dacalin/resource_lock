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

func (b *ResourceLockBuilder) WithRedisConfig(host, port string, poolSize int) *ResourceLockBuilder {
	b.redisHost = host
	b.redisPort = port
	b.redisPoolSize = poolSize
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
		itc := _rl_redis.New(b.redisHost, b.redisPort, b.redisPoolSize)
		itc.SetMaxLockTime(b.cleanMemMilis)
		instance = itc
		return instance

	default:
		panic("Invalid lock type")
	}
}
