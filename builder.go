package ioc

import (
	"log"
	"reflect"
	"sort"
	"sync"

	"github.com/optimus-hft/lockset/v2"
)

type ScopeID any

var (
	ScopeSingleton ScopeID = struct{}{}
	ScopeTransient ScopeID = struct{}{}
)

type serviceID any

type builder struct {
	wraps                 map[serviceID][]ctorWrap
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
			wraps:                 map[serviceID][]ctorWrap{},
			services:              map[serviceID]Service{},
			eagerSingletonLoading: true,
			scopes:                map[ScopeID]struct{}{},
		},
	}
}

func (b Builder) Clone() Builder {
	clonedB := Builder{
		b: &builder{
			wraps:                 make(map[serviceID][]ctorWrap, len(b.b.wraps)),
			services:              make(map[serviceID]Service, len(b.b.services)),
			eagerSingletonLoading: b.b.eagerSingletonLoading,
			scopes:                make(map[ScopeID]struct{}, len(b.b.scopes)),
		},
	}
	for key, val := range b.b.wraps {
		wraps := make([]ctorWrap, len(val))
		copy(wraps, val)
		clonedB.b.wraps[key] = wraps
	}
	for key, val := range b.b.services {
		clonedB.b.services[key] = val
	}
	for key, val := range b.b.scopes {
		clonedB.b.scopes[key] = val
	}
	return clonedB
}

type ctorWraps []ctorWrap

func (a ctorWraps) Len() int      { return len(a) }
func (a ctorWraps) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ctorWraps) Less(i, j int) bool {
	if a[i].order != a[j].order {
		return a[i].order < a[j].order
	}
	return false
}

func (b Builder) Build() Dic {
	services := b.b.services
	for key, service := range services {
		wraps, ok := b.b.wraps[key]
		if !ok || len(wraps) == 0 {
			continue
		}
		sort.Sort(ctorWraps(wraps))
		ctor := service.creator
		w := []ctorWrap(wraps)
		service.creator = func(d Dic) any {
			s := ctor(d)
			return s
		}
		service.wraps = func(d Dic, s any) {
			for _, wrap := range w {
				wrap.wraps(d, s)
			}
		}
		services[key] = service
	}
	c := Dic{
		c: &dic{
			serviceRegisterMutex: &sync.Mutex{},
			services:             services,
			createLockset:        lockset.New(),
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
				c.c.createLockset.Lock(key)
				serviceValue := service.creator(c)
				service.additional = SingletonAdditional{
					Service: serviceValue,
				}
				b.b.services[key] = service
				c.c.createLockset.Unlock(key)
				service.wraps(c, serviceValue)
			}
		}
	}
	return c
}

func (b Builder) Wrap(wrap func(Builder) Builder) Builder {
	return wrap(b)
}

func RegisterSingleton[Service any](b Builder, creator func(c Dic) Service) Builder {
	key := typeKey[Service]()
	if _, ok := b.b.services[key]; ok {
		var t Service
		log.Panicf("registered service already exists '%s'", reflect.TypeOf(t).String())
	}
	service := newSingleton(func(c Dic) any { return creator(c) })
	b.b.services[key] = service
	return b
}

func RegisterScoped[Service any](b Builder, scope ScopeID, creator func(c Dic) Service) Builder {
	if scope == ScopeSingleton {
		return RegisterSingleton(b, creator)
	}
	if scope == ScopeTransient {
		return RegisterTransient(b, creator)
	}
	key := typeKey[Service]()
	if _, ok := b.b.services[key]; ok {
		var t Service
		log.Panicf("registered service already exists '%s'", reflect.TypeOf(t).String())
	}
	service := newScoped(scope, func(c Dic) any { return creator(c) })
	b.b.services[key] = service
	return b
}

func RegisterTransient[Service any](b Builder, creator func(c Dic) Service) Builder {
	key := typeKey[Service]()
	if _, ok := b.b.services[key]; ok {
		var t Service
		log.Panicf("registered service already exists '%s'", reflect.TypeOf(t).String())
	}
	service := newTransient(func(c Dic) any { return creator(c) })
	b.b.services[key] = service
	return b
}

// wraps with the smallest id are applied first
// wraps with the same order are applied randomly
func WrapService[Service any](b Builder, wrap func(c Dic, s Service)) Builder {
	order := DefaultOrder
	key := typeKey[Service]()
	wraps := newCtorWrap(order, wrap)

	if _, ok := b.b.wraps[key]; !ok {
		b.b.wraps[key] = make([]ctorWrap, 0, 1)
	}

	b.b.wraps[key] = append(b.b.wraps[key], wraps)
	return b
}

func WrapServiceInOrder[Service any](b Builder, order Order, wrap func(c Dic, s Service)) Builder {
	key := typeKey[Service]()
	wraps := newCtorWrap(order, wrap)

	if _, ok := b.b.wraps[key]; !ok {
		b.b.wraps[key] = make([]ctorWrap, 0, 1)
	}

	b.b.wraps[key] = append(b.b.wraps[key], wraps)
	return b
}

// panics when attempting to regsiter ScopeSingleton or ScopeTransient
func (b Builder) RegisterScope(scope ScopeID) {
	if scope == ScopeSingleton {
		panic(ErrScopeDoesNotExist)
	}
	if scope == ScopeTransient {
		panic(ErrScopeDoesNotExist)
	}
	b.b.scopes[scope] = struct{}{}
}

func (b Builder) LazySingletonLoading() {
	b.b.eagerSingletonLoading = false
}
