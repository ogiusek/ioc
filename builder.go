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
	b *builder
}

func NewBuilder() Builder {
	return Builder{
		b: &builder{
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
			services:               &b.b.services,
			singletons:             &singletons,
			scopedCreateLockset:    lockset.New(),
			scoped:                 map[serviceID]any{},
		},
	}
	for key, service := range b.b.services {
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
	if _, ok := b.b.services[key]; ok {
		var t T
		log.Panicf("registered service already exists '%s'", reflect.TypeOf(t).String())
	}
	service := newSingleton(func(c Dic) any { return creator(c) })
	b.b.services[key] = service
	return ensureWrapped[T](b)
}

func RegisterScoped[T any](b Builder, creator func(c Dic) T) Builder {
	key := typeKey[T]()
	if _, ok := b.b.services[key]; ok {
		var t T
		log.Panicf("registered service already exists '%s'", reflect.TypeOf(t).String())
	}
	service := newScoped(func(c Dic) any { return creator(c) })
	b.b.services[key] = service
	return ensureWrapped[T](b)
}

func RegisterTransient[T any](b Builder, creator func(c Dic) T) Builder {
	key := typeKey[T]()
	if _, ok := b.b.services[key]; ok {
		var t T
		log.Panicf("registered service already exists '%s'", reflect.TypeOf(t).String())
	}
	service := newTransient(func(c Dic) any { return creator(c) })
	b.b.services[key] = service
	return ensureWrapped[T](b)
}

func ensureWrapped[T any](b Builder) Builder {
	key := typeKey[T]()
	notAppliedWraps, ok := b.b.notAppliedWraps[key]
	if !ok {
		return b
	}
	s, ok := b.b.services[key]
	if !ok {
		return b
	}
	s.wrap(notAppliedWraps)
	b.b.services[key] = s
	delete(b.b.notAppliedWraps, key)
	return b
}

func WrapService[T any](b Builder, wrap func(c Dic, s T) T) Builder {
	wraps := newCtorWrap(wrap)
	key := typeKey[T]()
	s, ok := b.b.services[key]
	if ok {
		s.wrap(wraps)
		b.b.services[key] = s
		return b
	}

	originalWraps, ok := b.b.notAppliedWraps[key]
	if !ok {
		b.b.notAppliedWraps[key] = wraps
		return b
	}

	originalWraps.wrap(wraps)
	b.b.notAppliedWraps[key] = originalWraps
	return b
}
