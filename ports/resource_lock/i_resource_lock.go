package _i_resource_lock

type IResourceLock interface {
	Lock(id string)
	Unlock(id string)
}
