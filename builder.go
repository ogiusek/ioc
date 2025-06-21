package ioc

import (
	"log"
	"reflect"
	"sync"

	"github.com/ogiusek/lockset"
)

type ScopeID any
type serviceID any

type builder struct {
	notAppliedWraps       map[serviceID]ctorWraps
	services              map[serviceID]Service
	eagerSingletonLoading bool
	scopes                map[ScopeID]struct{}
}

type Builder struct {
	b *builder
}

func NewBuilder() Builder {
	return Builder{
		b: &builder{
			notAppliedWraps:       map[serviceID]ctorWraps{},
			services:              map[serviceID]Service{},
			eagerSingletonLoading: true,
			scopes:                map[ScopeID]struct{}{},
		},
	}
}

func (b Builder) Clone() Builder {
	clonedB := Builder{
		b: &builder{
			notAppliedWraps:       make(map[serviceID]ctorWraps, len(b.b.notAppliedWraps)),
			services:              make(map[serviceID]Service, len(b.b.services)),
			eagerSingletonLoading: b.b.eagerSingletonLoading,
			scopes:                make(map[ScopeID]struct{}, len(b.b.scopes)),
		},
	}
	for key, val := range b.b.notAppliedWraps {
		clonedB.b.notAppliedWraps[key] = val
	}
	for key, val := range b.b.services {
		clonedB.b.services[key] = val
	}
	for key, val := range b.b.scopes {
		clonedB.b.scopes[key] = val
	}
	return clonedB
}

func (b Builder) Build() Dic {
	services := b.b.services
	c := Dic{
		c: &dic{
			serviceRegisterMutex: &sync.Mutex{},
			services:             services,
			scopedCreateLockset:  lockset.New(),
			scopes:               map[ScopeID]map[serviceID]any{},
		},
	}
	for scopeId := range b.b.scopes {
		c.c.scopes[scopeId] = map[serviceID]any{}
	}
	if b.b.eagerSingletonLoading {
		for key, service := range b.b.services {
			if service.lifetime != singleton {
				continue
			}
			if service.additional == nil {
				c.c.scopedCreateLockset.Lock(key)
				service.additional = SingletonAdditional{
					Service: service.creator(c),
				}
				b.b.services[key] = service
				c.c.scopedCreateLockset.Unlock(key)
			}
		}
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

func RegisterScoped[T any](b Builder, scope ScopeID, creator func(c Dic) T) Builder {
	key := typeKey[T]()
	if _, ok := b.b.services[key]; ok {
		var t T
		log.Panicf("registered service already exists '%s'", reflect.TypeOf(t).String())
	}
	service := newScoped(scope, func(c Dic) any { return creator(c) })
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

func (b Builder) RegisterScope(scope ScopeID) {
	b.b.scopes[scope] = struct{}{}
}

func (b Builder) LazySingletonLoading() {
	b.b.eagerSingletonLoading = false
}
