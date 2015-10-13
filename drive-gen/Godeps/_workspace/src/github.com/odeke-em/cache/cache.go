package expirable

import (
	"sync"
	"time"
)

type Operation int

const (
	Noop Operation = 1 << iota
	Create
	Read
	Update
	Delete
)

type Keyable interface{}

type Expirable interface {
	Expired(baseTime time.Time) bool
	Value() interface{}
}

type KeyValue struct {
	Key   Keyable
	Value Expirable
}

type OperationCache struct {
	mu   *sync.Mutex
	body map[Keyable]Expirable
}

func New() *OperationCache {
	return &OperationCache{
		mu:   &sync.Mutex{},
		body: make(map[Keyable]Expirable),
	}
}

func (oc *OperationCache) Put(k Keyable, v Expirable) bool {
	oc.mu.Lock()
	defer oc.mu.Unlock()

	oc.body[k] = v
	return true
}

func (oc *OperationCache) Get(k Keyable) (Expirable, bool) {
	oc.mu.Lock()
	v, ok := oc.body[k]
	oc.mu.Unlock()

	if ok && v != nil && v.Expired(time.Now()) {
		prev, _ := oc.Remove(k)
		return prev, false
	}

	return v, ok
}

func (oc *OperationCache) Remove(k Keyable) (prev Expirable, exist bool) {
	oc.mu.Lock()
	defer oc.mu.Unlock()

	prev, exist = oc.body[k]
	if exist {
		delete(oc.body, k)
	}

	return prev, exist
}

func (oc *OperationCache) Items() chan *KeyValue {
	items := make(chan *KeyValue)

	go func() {
		defer close(items)

		for k, v := range oc.body {
			items <- &KeyValue{Key: k, Value: v}
		}
	}()

	return items
}
