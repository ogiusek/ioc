package ioc

import (
	"errors"
	"fmt"
	"reflect"
)

func typeKey[T any]() serviceID {
	return reflect.TypeFor[T]()
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
			errors.New(fmt.Sprintf("Service of type '%s' is not registered", reflect.TypeFor[T]().String())),
		)
	}

	var res any

	switch service.lifetime {
	case singleton:
		if service.additional == nil {
			c.c.scopedCreateLockset.Lock(key)
			service.additional = singletonAdditional{
				Service: service.creator(c),
			}
			c.c.services[key] = service
			c.c.scopedCreateLockset.Unlock(key)
		}
		res = service.additional.(singletonAdditional).Service
		break
	case scoped:
		additional := service.additional.(scopedAdditional)
		scope, ok := c.c.scopes[additional.Scope]
		if !ok {
			var t T
			return t, ErrScopeIsNotInitialized
		}
		res, ok = scope[key]
		if !ok {
			c.c.scopedCreateLockset.Lock(key)
			res, ok = c.c.scopes[key]
			if !ok {
				res = service.creator(c)
				scope[key] = res
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
