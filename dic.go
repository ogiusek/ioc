package ioc

import (
	"fmt"
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

type Dic struct {
	c *dic
}

func serviceKey(serviceAddr reflect.Value) any {
	return serviceAddr.Type()
}

func (c Dic) Scope() Dic {
	return Dic{
		c: &dic{
			serviceResiterMutex:  c.c.serviceResiterMutex,
			services:             c.c.services,
			singletonCreateMutex: c.c.singletonCreateMutex,
			singletons:           c.c.singletons,
			scopedCreateMutex:    &sync.Mutex{},
			scoped:               map[any]any{},
		},
	}
}

func (c Dic) Inject(servicePointer any) {
	serviceValue := reflect.ValueOf(servicePointer)
	if serviceValue.Kind() != reflect.Ptr || serviceValue.IsNil() {
		panic("service must be a non-nil pointer")
	}
	serviceElement := serviceValue.Elem()

	key := serviceKey(serviceElement)

	service, ok := (*c.c.services)[key]
	if !ok {
		log.Panicf("Service of type '%s' is not registered", serviceElement.Type().String())
	}
	switch service.lifetime {
	case singleton:
		existing, ok := (*c.c.singletons)[key]
		if !ok {
			c.c.singletonCreateMutex.Lock()
			existing, ok = (*c.c.singletons)[key]
			if !ok {
				existing = service.creator(c)
				(*c.c.singletons)[key] = existing
			}
			c.c.singletonCreateMutex.Unlock()
		}
		serviceElement.Set(reflect.ValueOf(existing))
	case scoped:
		existing, ok := c.c.scoped[key]
		if !ok {
			c.c.scopedCreateMutex.Lock()
			existing, ok = c.c.scoped[key]
			if !ok {
				existing = service.creator(c)
				c.c.scoped[key] = existing
			}
			c.c.scopedCreateMutex.Unlock()
		}
		serviceElement.Set(reflect.ValueOf(existing))
	case transient:
		serviceElement.Set(reflect.ValueOf(service.creator(c)))
	default:
		panic("requested service has invalid lifetime")
	}

}

func (c Dic) InjectServices(services any) {
	servicePointer := reflect.ValueOf(services)
	if servicePointer.Kind() != reflect.Ptr {
		panic(fmt.Sprintf("not a pointer: %T", services))
	}

	serviceElem := servicePointer.Elem()
	if serviceElem.Kind() != reflect.Struct {
		panic(fmt.Sprintf("expected pointer to struct, got pointer to %s", serviceElem.Kind()))
	}

	serviceType := serviceElem.Type()
	fields := serviceType.NumField()

	for i := 0; i < fields; i++ {
		field := serviceType.Field(i)
		if field.Tag.Get("inject") != "1" {
			continue
		}

		fieldPointer := serviceElem.Field(i).Addr().Interface()
		c.Inject(fieldPointer)
	}
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
