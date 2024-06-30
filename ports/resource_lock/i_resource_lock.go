package _i_resource_lock

type IResourceLock interface {
	Instance() IResourceLock
	Lock(id string)
	Unlock(id string)
}
