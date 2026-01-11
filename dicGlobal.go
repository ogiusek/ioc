package ioc

import (
	"errors"
	"fmt"
	"reflect"
)

func typeKey[T any]() serviceID {
	return (*T)(nil)
}

// Returns service instance of type T.
// Returns error when T is not registered
func TryGet[T any](c Dic) (T, error) {
	key := typeKey[T]()

	service, ok := c.c.services[key]
	if !ok {
		var t T
		return t, errors.Join(
			ErrServiceIsntRegistered,
			fmt.Errorf("Service of type '%s' is not registered", reflect.TypeFor[T]().String()),
		)
	}

	var res any

	switch service.lifetime {
	case singleton:
		if service.additional == nil {
			if ok := c.c.createLockset.TryLock(key); !ok {
				panic("detected circular dependency")
			}
			service.additional = SingletonAdditional{
				Service: service.creator(c),
			}
			c.c.services[key] = service
			c.c.createLockset.Unlock(key)
			service.wraps(c, service.additional.(SingletonAdditional).Service)
		}
		res = service.additional.(SingletonAdditional).Service
	case scoped:
		additional := service.additional.(ScopedAdditional)
		scope, ok := c.c.scopes[additional.Scope]
		if !ok {
			var t T
			return t, ErrScopeIsNotInitialized
		}
		res, ok = scope[key]
		if !ok {
			if ok := c.c.createLockset.TryLock(key); !ok {
				panic("detected circular dependency")
			}
			res, ok = c.c.scopes[key]
			if !ok {
				res = service.creator(c)
				scope[key] = res
				c.c.createLockset.Unlock(key)
				service.wraps(c, res)
			} else {
				c.c.createLockset.Unlock(key)
			}
		}
	case transient:
		res = service.creator(c)
		service.wraps(c, res)
	default:
		panic("requested service has invalid lifetime")
	}

	return res.(T), nil
}

// Returns service instance of type T.
// Panics when T is not registered
func Get[T any](c Dic) T {
	s, err := TryGet[T](c)
	if err != nil {
		panic(err.Error())
	}
	return s
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
func TryGetServices[T any](c Dic) (T, error) {
	var res T
	t := reflect.TypeFor[T]()
	if t.Kind() == reflect.Pointer {
		reflect.ValueOf(&res).Elem().Set(reflect.New(t.Elem()))
		err := c.InjectServices(res)
		return res, err
	}
	err := c.InjectServices(&res)
	return res, err
}

func GetServices[T any](c Dic) T {
	res, err := TryGetServices[T](c)
	if err != nil {
		panic(err.Error())
	}
	return res
}
