package ioc

import (
	"log"
	"reflect"
	"sync"
)

func typeKey[T any]() any {
	var service T
	// serviceType := reflect.TypeOf(&service).Elem()
	// log.Printf("type key called for: %s", serviceType)
	return reflect.TypeOf(&service).Elem()
}

// Returns service instance of type T.
// Panics when T is not registered
func Get[T any](c Dic) T {
	var res T
	c.Inject(&res)
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
	c.InjectServices(&res)
	return res
}

// Returns new Scope
func Scope(c Dic) Dic {
	return c.Scope()
}

func RegisterSingleton[T any](c Dic, creator func(c Dic) T) {
	c.c.serviceResiterMutex.Lock()
	defer c.c.serviceResiterMutex.Unlock()
	key := typeKey[T]()
	if _, ok := (*c.c.services)[key]; ok {
		var t T
		log.Panicf("registered service already exists '%s'", reflect.TypeOf(t).String())
	}
	service := newSingleton(func(c Dic) any { return creator(c) })
	(*c.c.services)[key] = service
}

func RegisterScoped[T any](c Dic, creator func(c Dic) T) {
	c.c.serviceResiterMutex.Lock()
	defer c.c.serviceResiterMutex.Unlock()
	key := typeKey[T]()
	if _, ok := (*c.c.services)[key]; ok {
		var t T
		log.Panicf("registered service already exists '%s'", reflect.TypeOf(t).String())
	}
	service := newScoped(func(c Dic) any { return creator(c) })
	(*c.c.services)[key] = service
}

func RegisterTransient[T any](c Dic, creator func(c Dic) T) {
	c.c.serviceResiterMutex.Lock()
	defer c.c.serviceResiterMutex.Unlock()
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
			serviceResiterMutex:  &sync.Mutex{},
			services:             &map[any]Service{},
			singletonCreateMutex: &sync.Mutex{},
			singletons:           &map[any]any{},
			scopedCreateMutex:    &sync.Mutex{},
			scoped:               map[any]any{},
		},
	}
}
