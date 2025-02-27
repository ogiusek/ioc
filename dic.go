package ioc

import (
	"log"
	"reflect"
)

type Dic struct {
	services   *map[any]Service
	singletons *map[any]any
	scoped     map[any]any
}

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
			existing = service.creator(c)
			(*c.singletons)[key] = existing
		}
		return existing.(T)
	case scoped:
		existing, ok := c.scoped[key]
		if !ok {
			existing = service.creator(c)
			c.scoped[key] = existing
		}
		return existing.(T)
	case transient:
		return service.creator(c).(T)
	default:
		panic("requested service has invalid lifetime")
	}
}

func Scope(c Dic) Dic {
	return Dic{
		services:   c.services,
		singletons: c.singletons,
		scoped:     map[any]any{},
	}
}

func RegisterSingleton[T any](c Dic, creator func(Dic) T) {
	key := typeKey[T]()
	if _, ok := (*c.services)[key]; ok {
		var t T
		log.Panicf("registered service already exists '%s'", reflect.TypeOf(t).String())
	}
	service := newSingleton(func(c Dic) any { return creator(c) })
	(*c.services)[key] = service
}

func RegisterScoped[T any](c Dic, creator func(Dic) T) {
	key := typeKey[T]()
	if _, ok := (*c.services)[key]; ok {
		var t T
		log.Panicf("registered service already exists '%s'", reflect.TypeOf(t).String())
	}
	service := newScoped(func(c Dic) any { return creator(c) })
	(*c.services)[key] = service
}

func RegisterTransient[T any](c Dic, creator func(Dic) T) {
	key := typeKey[T]()
	if _, ok := (*c.services)[key]; ok {
		var t T
		log.Panicf("registered service already exists '%s'", reflect.TypeOf(t).String())
	}
	service := newTransient(func(c Dic) any { return creator(c) })
	(*c.services)[key] = service
}

func NewContainer() Dic {
	return Dic{
		services:   &map[any]Service{},
		singletons: &map[any]any{},
		scoped:     map[any]any{},
	}
}
