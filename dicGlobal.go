package ioc

import (
	"log"
	"reflect"
	"sync"

	"github.com/optimus-hft/lockset"
)

func typeKey[T any]() any {
	var service T
	return reflect.TypeOf(&service).Elem()
}

// Returns service instance of type T.
// Panics when T is not registered
func Get[T any](c Dic) T {
	var res T
	if err := c.Inject(&res); err != nil {
		panic(err)
	}
	return res
}

// GetServices creates a new instance of type T, injects dependencies into it, and returns it.
//
// The type parameter T must be a struct type. All fields of the struct that have the tag
// `inject:"1"` will be automatically injected with corresponding instances from the DI container.
//
// Example:
//
//	type MyServices struct {
//	    Logger Logger `inject:"1"`
//	    Repo   Repo   `inject:"1"`
//	}
//	svc := GetServices[MyServices](dic)
//
// Note: If T is not a struct type, or if injection fails, this function may panic.
func GetServices[T any](c Dic) T {
	var res T
	if err := c.InjectServices(&res); err != nil {
		panic(err)
	}
	return res
}

// Returns new Scope
func Scope(c Dic) Dic {
	return c.Scope()
}

func RegisterSingleton[T any](c Dic, creator func(c Dic) T) {
	c.c.serviceRegisterMutex.Lock()
	defer c.c.serviceRegisterMutex.Unlock()
	key := typeKey[T]()
	if _, ok := (*c.c.services)[key]; ok {
		var t T
		log.Panicf("registered service already exists '%s'", reflect.TypeOf(t).String())
	}
	service := newSingleton(func(c Dic) any { return creator(c) })
	(*c.c.services)[key] = service
}

func RegisterScoped[T any](c Dic, creator func(c Dic) T) {
	c.c.serviceRegisterMutex.Lock()
	defer c.c.serviceRegisterMutex.Unlock()
	key := typeKey[T]()
	if _, ok := (*c.c.services)[key]; ok {
		var t T
		log.Panicf("registered service already exists '%s'", reflect.TypeOf(t).String())
	}
	service := newScoped(func(c Dic) any { return creator(c) })
	(*c.c.services)[key] = service
}

func RegisterTransient[T any](c Dic, creator func(c Dic) T) {
	c.c.serviceRegisterMutex.Lock()
	defer c.c.serviceRegisterMutex.Unlock()
	key := typeKey[T]()
	if _, ok := (*c.c.services)[key]; ok {
		var t T
		log.Panicf("registered service already exists '%s'", reflect.TypeOf(t).String())
	}
	service := newTransient(func(c Dic) any { return creator(c) })
	(*c.c.services)[key] = service
}

func NewContainer() Dic {
	return Dic{
		c: &dic{
			serviceRegisterMutex:   &sync.Mutex{},
			services:               &map[any]Service{},
			singletonCreateLockset: lockset.New(),
			singletons:             &map[any]any{},
			scopedCreateLockset:    lockset.New(),
			scoped:                 map[any]any{},
		},
	}
}
