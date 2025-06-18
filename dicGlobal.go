package ioc

import (
	"errors"
	"fmt"
	"log"
	"reflect"
	"sync"

	"github.com/ogiusek/lockset"
)

func typeKey[T any]() serviceID {
	return (*T)(nil)
}

// Returns service instance of type T.
// Panics when T is not registered
func Get[T any](c Dic) T {
	key := typeKey[T]()

	service, ok := (*c.c.services)[key]
	if !ok {
		panic(errors.Join(
			ErrServiceIsntRegistered,
			errors.New(fmt.Sprintf("Service of type '%s' is not registered", reflect.TypeFor[T]().String())),
		))
	}

	var res any

	switch service.lifetime {
	case singleton:
		res, ok = (*c.c.singletons)[key]
		if !ok {
			c.c.singletonCreateLockset.Lock(key)
			res, ok = (*c.c.singletons)[key]
			if !ok {
				res = service.creator(c)
				(*c.c.singletons)[key] = res
			}
			c.c.singletonCreateLockset.Unlock(key)
		}
		break
	case scoped:
		res, ok = c.c.scoped[key]
		if !ok {
			c.c.scopedCreateLockset.Lock(key)
			res, ok = c.c.scoped[key]
			if !ok {
				res = service.creator(c)
				c.c.scoped[key] = res
			}
			c.c.scopedCreateLockset.Unlock(key)
		}
		break
	case transient:
		res = service.creator(c)
		break
	default:
		panic("requested service has invalid lifetime")
	}

	return res.(T)
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
	ensureWrapped[T](c)
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
	ensureWrapped[T](c)
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
	ensureWrapped[T](c)
}

func ensureWrapped[T any](c Dic) {
	key := typeKey[T]()
	notAppliedWraps, ok := (*c.c.notAppliedWraps)[key]
	if !ok {
		return
	}
	s, ok := (*c.c.services)[key]
	if !ok {
		return
	}
	s.wrap(notAppliedWraps)
	(*c.c.services)[key] = s
	delete(*c.c.notAppliedWraps, key)

}

func WrapService[T any](c Dic, wrap func(c Dic, s T) T) {
	wraps := newCtorWrap(wrap)
	c.c.serviceRegisterMutex.Lock()
	defer c.c.serviceRegisterMutex.Unlock()
	key := typeKey[T]()
	s, ok := (*c.c.services)[key]
	if ok {
		s.wrap(wraps)
		(*c.c.services)[key] = s
		return
	}

	originalWraps, ok := (*c.c.notAppliedWraps)[key]
	if !ok {
		(*c.c.notAppliedWraps)[key] = wraps
		return
	}

	originalWraps.wrap(wraps)
	(*c.c.notAppliedWraps)[key] = originalWraps
}

func NewContainer() Dic {
	return Dic{
		c: &dic{
			serviceRegisterMutex:   &sync.Mutex{},
			services:               &map[serviceID]Service{},
			notAppliedWraps:        &map[serviceID]ctorWraps{},
			singletonCreateLockset: lockset.New(),
			singletons:             &map[serviceID]any{},
			scopedCreateLockset:    lockset.New(),
			scoped:                 map[serviceID]any{},
		},
	}
}
