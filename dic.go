package ioc

import (
	"errors"
	"fmt"
	"reflect"
	"sync"

	"github.com/ogiusek/lockset"
)

type dic struct {
	serviceRegisterMutex *sync.Mutex
	services             map[serviceID]Service

	scopedCreateLockset *lockset.Set
	scopes              map[ScopeID]map[serviceID]any
}

type Dic struct {
	c *dic
}

func serviceKey(serviceType reflect.Type) serviceID {
	return reflect.Zero(reflect.PointerTo(serviceType)).Interface()
}

// can return ErrScopeDoesNotExist
func (c Dic) TryScope(scope ScopeID) (Dic, error) {
	if _, ok := c.c.scopes[scope]; !ok {
		return Dic{}, errors.Join(
			ErrScopeDoesNotExist,
			fmt.Errorf("scope %s", scope),
		)
	}
	s := Dic{
		c: &dic{
			serviceRegisterMutex: c.c.serviceRegisterMutex,
			services:             c.c.services,
			scopedCreateLockset:  lockset.New(),
			scopes:               map[ScopeID]map[serviceID]any{},
		},
	}
	for copiedScope, scopeServices := range c.c.scopes {
		s.c.scopes[copiedScope] = scopeServices
	}
	s.c.scopes[scope] = map[serviceID]any{}
	return s, nil
}

func (c Dic) Scope(scope ScopeID) Dic {
	s, err := c.TryScope(scope)
	if err != nil {
		panic(fmt.Sprintf("%s\n", err.Error()))
	}
	return s
}

// Inject replaces servicePointer value with a service from container.
// Can return ErrServiceIsntRegistered or ErrIsntPointer
func (c Dic) Inject(servicePointer any) error {
	if servicePointer == nil {
		return ErrIsntPointer
	}
	serviceValue := reflect.ValueOf(servicePointer)
	if serviceValue.Kind() != reflect.Ptr {
		return ErrIsntPointer
	}
	serviceElement := serviceValue.Elem()

	key := serviceKey(serviceElement.Type())

	service, ok := c.c.services[key]
	if !ok {
		return errors.Join(
			ErrServiceIsntRegistered,
			errors.New(fmt.Sprintf("Service of type '%s' is not registered", serviceElement.Type().String())),
		)
	}

	var existing any

	switch service.lifetime {
	case singleton:
		if service.additional == nil {
			c.c.scopedCreateLockset.Lock(key)
			service.additional = SingletonAdditional{
				Service: service.creator(c),
			}
			c.c.services[key] = service
			c.c.scopedCreateLockset.Unlock(key)
		}
		existing = service.additional.(SingletonAdditional).Service
		break
	case scoped:
		additional := service.additional.(ScopedAdditional)
		scope, ok := c.c.scopes[additional.Scope]
		if !ok {
			return ErrScopeIsNotInitialized
		}
		existing, ok = scope[key]
		if !ok {
			c.c.scopedCreateLockset.Lock(key)
			existing, ok = c.c.scopes[key]
			if !ok {
				existing = service.creator(c)
				scope[key] = existing
			}
			c.c.scopedCreateLockset.Unlock(key)
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
