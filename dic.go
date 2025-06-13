package ioc

import (
	"errors"
	"fmt"
	"reflect"
	"sync"

	"github.com/optimus-hft/lockset"
)

type dic struct {
	serviceRegisterMutex   *sync.Mutex
	serviceCreationLockSet lockset.Set
	services               *map[any]Service

	singletonCreateLockset *lockset.Set
	singletons             *map[any]any

	scopedCreateLockset *lockset.Set
	scoped              map[any]any
}

type Dic struct {
	c *dic
}

func serviceKey(serviceType reflect.Type) any {
	return serviceType
}

func (c Dic) Scope() Dic {
	return Dic{
		c: &dic{
			serviceRegisterMutex:   c.c.serviceRegisterMutex,
			services:               c.c.services,
			singletonCreateLockset: c.c.singletonCreateLockset,
			singletons:             c.c.singletons,
			scopedCreateLockset:    lockset.New(),
			scoped:                 map[any]any{},
		},
	}
}

var (
	ErrIsntPointer           error = errors.New("isn't a pointer")
	ErrIsntPointerToStruct   error = errors.New("isn't a pointer to a struct")
	ErrServiceIsntRegistered error = errors.New("service isn't registered")
)

// Inject replaces servicePointer value with a service from container.
// Can return ErrServiceIsntRegistered or ErrIsntPointer
func (c Dic) Inject(servicePointer any) error {
	serviceValue := reflect.ValueOf(servicePointer)
	if serviceValue.Kind() != reflect.Ptr || serviceValue.IsNil() {
		return ErrIsntPointer
	}
	serviceElement := serviceValue.Elem()

	key := serviceKey(serviceElement.Type())
	stringKey := fmt.Sprintf("%s", key)

	service, ok := (*c.c.services)[key]
	if !ok {
		return errors.Join(
			ErrServiceIsntRegistered,
			errors.New(fmt.Sprintf("Service of type '%s' is not registered", serviceElement.Type().String())),
		)
	}

	var existing any

	switch service.lifetime {
	case singleton:
		existing, ok = (*c.c.singletons)[key]
		if !ok {
			c.c.singletonCreateLockset.Lock(stringKey)
			existing, ok = (*c.c.singletons)[key]
			if !ok {
				existing = service.creator(c)
				(*c.c.singletons)[key] = existing
			}
			c.c.singletonCreateLockset.Unlock(stringKey)
		}
		break
	case scoped:
		existing, ok = c.c.scoped[key]
		if !ok {
			c.c.scopedCreateLockset.Lock(stringKey)
			existing, ok = c.c.scoped[key]
			if !ok {
				existing = service.creator(c)
				c.c.scoped[key] = existing
			}
			c.c.scopedCreateLockset.Unlock(stringKey)
		}
		break
	case transient:
		existing = service.creator(c)
		break
	default:
		panic("requested service has invalid lifetime")
	}

	var newServiceValue reflect.Value
	switch serviceElement.Type().Kind() {
	case reflect.Interface:
		if existing == nil {
			newServiceValue = reflect.ValueOf(&existing).Elem()
		} else {
			newServiceValue = reflect.ValueOf(existing)
		}
		break
	default:
		newServiceValue = reflect.ValueOf(existing)
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
	if servicePointer.Kind() != reflect.Ptr {
		return errors.Join(
			ErrIsntPointerToStruct,
			errors.New(fmt.Sprintf("not a pointer: %T", services)),
		)
	}

	serviceElem := servicePointer.Elem()
	if serviceElem.Kind() != reflect.Struct {
		return errors.Join(
			ErrIsntPointerToStruct,
			errors.New(fmt.Sprintf("expected pointer to struct, got pointer to %s", serviceElem.Kind())),
		)
	}

	serviceType := serviceElem.Type()
	fields := serviceType.NumField()

	for i := 0; i < fields; i++ {
		field := serviceType.Field(i)
		if field.Tag.Get("inject") != "1" {
			continue
		}

		fieldPointer := serviceElem.Field(i).Addr().Interface()
		err := c.Inject(fieldPointer)
		if err != nil {
			return err
		}
	}

	return nil
}
