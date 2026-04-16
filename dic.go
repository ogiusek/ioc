package ioc

import (
	"errors"
	"fmt"
	"reflect"
	"sync"
)

type dic struct {
	serviceRegisterMutex *sync.Mutex
	services             map[serviceID]Service

	creationMapMutex sync.Mutex
	creationMap      map[serviceID]struct{}
}

type Dic struct {
	c *dic
}

func serviceKey(serviceType reflect.Type) serviceID {
	return reflect.Zero(reflect.PointerTo(serviceType)).Interface()
}

func (c Dic) tryLock(id serviceID) bool {
	c.c.creationMapMutex.Lock()
	defer c.c.creationMapMutex.Unlock()

	_, ok := c.c.creationMap[id]
	c.c.creationMap[id] = struct{}{}
	return !ok
}
func (c Dic) unlock(id serviceID) {
	c.c.creationMapMutex.Lock()
	defer c.c.creationMapMutex.Unlock()
	delete(c.c.creationMap, id)
}

// Inject replaces servicePointer value with a service from container.
// Can return ErrServiceIsntRegistered or ErrIsntPointer
func (c Dic) Inject(servicePointer any) error {
	if servicePointer == nil {
		return ErrIsntPointer
	}
	serviceValue := reflect.ValueOf(servicePointer)
	if serviceValue.Kind() != reflect.Pointer {
		return ErrIsntPointer
	}
	serviceElement := serviceValue.Elem()

	key := serviceKey(serviceElement.Type())

	service, ok := c.c.services[key]
	if !ok {
		return errors.Join(
			ErrServiceIsntRegistered,
			fmt.Errorf("Service of type '%s' is not registered", serviceElement.Type().String()),
		)
	}

	instance := *service.instance
	if instance == nil {
		if ok := c.tryLock(key); !ok {
			panic(errors.Join(
				ErrCircularDependency,
				fmt.Errorf("Service of type '%s' is requested before being registered", serviceElement.Type().String()),
			))
		}
		instance = service.creator(c)
		*service.instance = instance
		c.c.services[key] = service
		c.unlock(key)
		service.wraps(c, instance)
	}

	var newServiceValue reflect.Value
	switch serviceElement.Type().Kind() {
	case reflect.Interface:
		if instance == nil {
			newServiceValue = reflect.ValueOf(&instance).Elem()
		} else {
			newServiceValue = reflect.ValueOf(instance)
		}
	default:
		newServiceValue = reflect.ValueOf(instance)
	}

	serviceElement.Set(newServiceValue)
	return nil
}

// InjectServices injects dependencies into the provided struct.
//
// The parameter `services` must be a pointer to a struct. All fields of this struct
// that have the tag `inject:"1"` will be automatically injected with corresponding
// instances from the DI container.
//
// Example:
//
//	type MyServices struct {
//	    Logger Logger `inject:"1"`
//	    Repo   Repo   `inject:"1"`
//	}
//	var svc MyServices
//	dic.InjectServices(&svc)
//
// can return ErrIsntPointerToStruct error or any error returned by c.Inject() method
func (c Dic) InjectServices(services any) error {
	servicePointer := reflect.ValueOf(services)
	if servicePointer.Kind() != reflect.Pointer {
		return errors.Join(
			ErrIsntPointerToStruct,
			fmt.Errorf("not a pointer: %T", services),
		)
	}

	serviceElem := servicePointer.Elem()
	if serviceElem.Kind() != reflect.Struct {
		return errors.Join(
			ErrIsntPointerToStruct,
			fmt.Errorf("expected pointer to struct, got pointer to %s", serviceElem.Kind()),
		)
	}

	serviceType := serviceElem.Type()
	fields := serviceType.NumField()

	injected := false

	for i := range fields {
		field := serviceType.Field(i)
		if field.Tag.Get("inject") != "1" {
			continue
		}

		fieldPointer := serviceElem.Field(i).Addr().Interface()
		if err := c.Inject(fieldPointer); err == nil {
			injected = true
			continue
		}
		if err := c.InjectServices(fieldPointer); err != nil {
			return errors.Join(
				ErrServiceIsntRegistered,
				fmt.Errorf("service %v isn't registered", field.Type.String()),
			)
		}
	}

	if !injected {
		return errors.Join(
			ErrMissingDependency,
		)
	}

	return nil
}
