package _rl_go

import (
	"sync"
	"time"
)

type resource struct {
	mutex    *sync.Mutex
	deadline time.Time
}
