package ioc

import (
	"log"
	"reflect"
	"sync"
)

type dic struct {
	serviceResiterMutex *sync.Mutex
	services            *map[any]Service
	// single mutex is used instead of the map because this is unnecesary optimization which does not optimize for this use case
	singletonCreateMutex *sync.Mutex
	singletons           *map[any]any
	// single mutex is used instead of the map because this is unnecesary optimization which does not optimize for this use case
	scopedCreateMutex *sync.Mutex
	scoped            map[any]any
}

type Dic *dic

func typeKey[T any]() any {
	return (*T)(nil)
}

func Get[T any](c Dic) T {
	key := typeKey[T]()
	service, ok := (*c.services)[key]
	if !ok {
		var t T
		log.Panicf("Service of type '%s' is not registered", reflect.TypeOf(t).String())
	}
	switch service.lifetime {
	case singleton:
		existing, ok := (*c.singletons)[key]
		if !ok {
			c.singletonCreateMutex.Lock()
			existing, ok = (*c.singletons)[key]
			if !ok {
				existing = service.creator(c)
				(*c.singletons)[key] = existing
			}
			c.singletonCreateMutex.Unlock()
		}
		return existing.(T)
	case scoped:
		existing, ok := c.scoped[key]
		if !ok {
			c.scopedCreateMutex.Lock()
			existing, ok = c.scoped[key]
			if !ok {
				existing = service.creator(c)
				c.scoped[key] = existing
			}
			c.scopedCreateMutex.Unlock()
		}
		return existing.(T)
	case transient:
		return service.creator(c).(T)
	default:
		panic("requested service has invalid lifetime")
	}
}

func Scope(c Dic) Dic {
	return &dic{
		serviceResiterMutex:  c.serviceResiterMutex,
		services:             c.services,
		singletonCreateMutex: c.singletonCreateMutex,
		singletons:           c.singletons,
		scopedCreateMutex:    &sync.Mutex{},
		scoped:               map[any]any{},
	}
}

func RegisterSingleton[T any](c Dic, creator func(c Dic) T) {
	c.serviceResiterMutex.Lock()
	defer c.serviceResiterMutex.Unlock()
	key := typeKey[T]()
	if _, ok := (*c.services)[key]; ok {
		var t T
		log.Panicf("registered service already exists '%s'", reflect.TypeOf(t).String())
	}
	service := newSingleton(func(c Dic) any { return creator(c) })
	(*c.services)[key] = service
}

func RegisterScoped[T any](c Dic, creator func(c Dic) T) {
	c.serviceResiterMutex.Lock()
	defer c.serviceResiterMutex.Unlock()
	key := typeKey[T]()
	if _, ok := (*c.services)[key]; ok {
		var t T
		log.Panicf("registered service already exists '%s'", reflect.TypeOf(t).String())
	}
	service := newScoped(func(c Dic) any { return creator(c) })
	(*c.services)[key] = service
}

func RegisterTransient[T any](c Dic, creator func(c Dic) T) {
	c.serviceResiterMutex.Lock()
	defer c.serviceResiterMutex.Unlock()
	key := typeKey[T]()
	if _, ok := (*c.services)[key]; ok {
		var t T
		log.Panicf("registered service already exists '%s'", reflect.TypeOf(t).String())
	}
	service := newTransient(func(c Dic) any { return creator(c) })
	(*c.services)[key] = service
}

func NewContainer() Dic {
	return &dic{
		serviceResiterMutex:  &sync.Mutex{},
		services:             &map[any]Service{},
		singletonCreateMutex: &sync.Mutex{},
		singletons:           &map[any]any{},
		scopedCreateMutex:    &sync.Mutex{},
		scoped:               map[any]any{},
	}
}
