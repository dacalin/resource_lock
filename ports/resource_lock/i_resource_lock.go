package _i_resource_lock

type IResourceLock interface {
	LockWithTTL(id string, ms int64)
	TryLockWithTTL(id string, ms int64)

	Lock(id string)
	TryLock(id string)

	Unlock(id string)
}
