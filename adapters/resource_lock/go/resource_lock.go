package _rl_go

import (
	"fmt"
	_i_resource_lock "github.com/dacalin/resource_lock/ports/resource_lock"
	"sync"
	"time"
)

var once sync.Once
var instance *GoResourceLock

type GoResourceLock struct {
	_i_resource_lock.IResourceLock
	resource      sync.Map
	cleanMemMilis int64
}

func Instance() *GoResourceLock {
	once.Do(func() {
		// Clean memory every 1 seconds by default
		instance = &GoResourceLock{
			cleanMemMilis: 30000,
		}
		go instance.cleanMemLoop()
	})

	return instance
}
func (self *GoResourceLock) SetMaxLockTime(ms int64) {
	Instance().cleanMemMilis = ms
}
func (self *GoResourceLock) Lock(id string) {
	value, ok := self.resource.Load(id)
	if !ok {
		value = self.create(id)
	}

	value.(*resource).mutex.Lock()
	rsc := &resource{
		mutex:    value.(*resource).mutex,
		deadline: time.Now().Add(time.Duration(self.cleanMemMilis) * time.Millisecond),
	}
	self.resource.Store(id, rsc)
}

func (self *GoResourceLock) Unlock(id string) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("Unlock::Recovered from panic: ", err)
		}
	}()

	value, ok := self.resource.Load(id)

	if ok {
		_, okrsc := value.(*resource)
		if !okrsc {
			fmt.Println("Unexpected type for resource with ID: ", id)
			return
		}

		value.(*resource).mutex.Unlock()
	}
}

func (self *GoResourceLock) create(id string) any {
	rsc := &resource{
		mutex:    &sync.Mutex{},
		deadline: time.Now().Add(time.Duration(self.cleanMemMilis) * time.Millisecond),
	}
	value, _ := self.resource.LoadOrStore(id, rsc)
	return value
}

func (self *GoResourceLock) cleanMem() {
	var removeIds []string

	self.resource.Range(func(key, value interface{}) bool {
		if time.Now().After(value.(*resource).deadline) {
			removeIds = append(removeIds, key.(string))
		}
		return true
	})

	for _, id := range removeIds {
		self.resource.Delete(id)
	}
}

func (self *GoResourceLock) cleanMemLoop() {
	ticker := time.NewTicker(time.Millisecond * time.Duration(self.cleanMemMilis/20))

	for {
		select {
		case <-ticker.C:
			self.cleanMem()
		}
	}
}
