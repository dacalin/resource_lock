package rl

import (
	_rl_go "github.com/dacalin/resource_lock/adapters/resource_lock/go"
	_rl_redis "github.com/dacalin/resource_lock/adapters/resource_lock/redis"
	_i_resource_lock "github.com/dacalin/resource_lock/ports/resource_lock"
)

type LockType int

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

func (b *ResourceLockBuilder) WithRedisConfig(host, port string, poolSize int) *ResourceLockBuilder {
	b.redisHost = host
	b.redisPort = port
	b.redisPoolSize = poolSize
	return b
}

func (b *ResourceLockBuilder) WithCleanMemInterval(milis int64) *ResourceLockBuilder {
	b.cleanMemMilis = milis
	return b
}

func (b *ResourceLockBuilder) Build() _i_resource_lock.IResourceLock {
	switch b.lockType {
	case Local:
		instance := _rl_go.Instance()
		instance.SetMaxLockTime(b.cleanMemMilis)
		return _rl_go.Instance()
	case Redis:
		instance := _rl_redis.New(b.redisHost, b.redisPort, b.redisPoolSize)
		instance.SetMaxLockTime(b.cleanMemMilis)
		return instance
	default:
		panic("Invalid lock type")
	}
}
