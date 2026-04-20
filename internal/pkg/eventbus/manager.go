package eventbus

import (
	"reflect"
	"sync"
)

var (
	mu   sync.RWMutex
	buss = make(map[reflect.Type]any)
)

func Get[T any]() *Bus[T] {
	key := eventTypeOf[T]()

	mu.RLock()
	bus, ok := buss[key]
	mu.RUnlock()
	if ok {
		return bus.(*Bus[T])
	}

	mu.Lock()
	defer mu.Unlock()

	if bus, ok = buss[key]; ok {
		return bus.(*Bus[T])
	}

	created := New[T]()
	buss[key] = created
	return created
}

func eventTypeOf[T any]() reflect.Type {
	typ := reflect.TypeFor[T]()
	if typ == nil {
		panic("eventbus: nil event type")
	}
	return typ
}
