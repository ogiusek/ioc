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
			fmt.Errorf("service of type '%s' is not registered", reflect.TypeFor[T]().String()),
		)
	}

	if instance := *service.instance; instance != nil {
		return instance.(T), nil
	}
	if ok := c.tryLock(key); !ok {
		panic(errors.Join(
			ErrCircularDependency,
			fmt.Errorf("service of type '%s' is requested before being registered", reflect.TypeFor[T]().String()),
		))
	}
	if instance := *service.instance; instance != nil {
		c.unlock(key)
		return instance.(T), nil
	}
	instance := service.creator(c)
	*service.instance = instance
	c.c.services[key] = service
	c.unlock(key)
	service.wraps(c, instance)

	return instance.(T), nil
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
// `inject:""` will be automatically injected with corresponding instances from the DI container.
//
// Example:
//
//	type MyServices struct {
//	    Logger Logger `inject:""`
//	    Repo   Repo   `inject:""`
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

// GetServices creates a new instance of type T where T is struct or pointer to a struct and
// injects services into every public property with inject struct tag.
// This function panics when injection fails.
// This is an intentional choice, its go idiomatic because its on startup.
// It ensures application consistency and there is no proper way to handle invalid application wiring.
func GetServices[T any](c Dic) T {
	res, err := TryGetServices[T](c)
	if err != nil {
		panic(err.Error())
	}
	return res
}
