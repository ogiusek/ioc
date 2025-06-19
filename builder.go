package ioc

import (
	"log"
	"reflect"
	"sync"

	"github.com/ogiusek/lockset"
)

type builder struct {
	notAppliedWraps map[serviceID]ctorWraps
	services        map[serviceID]Service
}

type Builder struct {
	c builder
}

func NewBuilder() Builder {
	return Builder{
		builder{
			notAppliedWraps: map[serviceID]ctorWraps{},
			services:        map[serviceID]Service{},
		},
	}
}

func (b Builder) Build() Dic {
	singletons := map[serviceID]any{}
	c := Dic{
		&dic{
			serviceRegisterMutex:   &sync.Mutex{},
			serviceCreationLockSet: *lockset.New(),
			services:               &b.c.services,
			singletons:             &singletons,
			scopedCreateLockset:    lockset.New(),
			scoped:                 map[serviceID]any{},
		},
	}
	for key, service := range b.c.services {
		if service.lifetime != singleton {
			continue
		}
		if _, ok := singletons[key]; ok {
			continue
		}
		singletons[key] = service.creator(c)
	}
	return c
}

func (b Builder) Wrap(wrap func(Builder) Builder) Builder {
	return wrap(b)
}

func RegisterSingleton[T any](b Builder, creator func(c Dic) T) Builder {
	key := typeKey[T]()
	if _, ok := b.c.services[key]; ok {
		var t T
		log.Panicf("registered service already exists '%s'", reflect.TypeOf(t).String())
	}
	service := newSingleton(func(c Dic) any { return creator(c) })
	b.c.services[key] = service
	return ensureWrapped[T](b)
}

func RegisterScoped[T any](b Builder, creator func(c Dic) T) Builder {
	key := typeKey[T]()
	if _, ok := b.c.services[key]; ok {
		var t T
		log.Panicf("registered service already exists '%s'", reflect.TypeOf(t).String())
	}
	service := newScoped(func(c Dic) any { return creator(c) })
	b.c.services[key] = service
	return ensureWrapped[T](b)
}

func RegisterTransient[T any](b Builder, creator func(c Dic) T) Builder {
	key := typeKey[T]()
	if _, ok := b.c.services[key]; ok {
		var t T
		log.Panicf("registered service already exists '%s'", reflect.TypeOf(t).String())
	}
	service := newTransient(func(c Dic) any { return creator(c) })
	b.c.services[key] = service
	return ensureWrapped[T](b)
}

func ensureWrapped[T any](b Builder) Builder {
	key := typeKey[T]()
	notAppliedWraps, ok := b.c.notAppliedWraps[key]
	if !ok {
		return b
	}
	s, ok := b.c.services[key]
	if !ok {
		return b
	}
	s.wrap(notAppliedWraps)
	b.c.services[key] = s
	delete(b.c.notAppliedWraps, key)
	return b
}

func WrapService[T any](b Builder, wrap func(c Dic, s T) T) Builder {
	wraps := newCtorWrap(wrap)
	key := typeKey[T]()
	s, ok := b.c.services[key]
	if ok {
		s.wrap(wraps)
		b.c.services[key] = s
		return b
	}

	originalWraps, ok := b.c.notAppliedWraps[key]
	if !ok {
		b.c.notAppliedWraps[key] = wraps
		return b
	}

	originalWraps.wrap(wraps)
	b.c.notAppliedWraps[key] = originalWraps
	return b
}
