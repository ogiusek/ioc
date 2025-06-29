package ioc

import (
	"log"
	"reflect"
	"sort"
	"sync"

	"github.com/ogiusek/lockset"
)

type ScopeID any
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
		wraps := make([]ctorWrap, 0, len(val))
		for _, val := range val {
			wraps = append(wraps, val)
		}
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
			for _, wrap := range w {
				s = wrap.wraps(d, s)
			}
			return s
		}
		services[key] = service
	}
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
	return b
}

func RegisterScoped[T any](b Builder, scope ScopeID, creator func(c Dic) T) Builder {
	key := typeKey[T]()
	if _, ok := b.b.services[key]; ok {
		var t T
		log.Panicf("registered service already exists '%s'", reflect.TypeOf(t).String())
	}
	service := newScoped(scope, func(c Dic) any { return creator(c) })
	b.b.services[key] = service
	return b
}

func RegisterTransient[T any](b Builder, creator func(c Dic) T) Builder {
	key := typeKey[T]()
	if _, ok := b.b.services[key]; ok {
		var t T
		log.Panicf("registered service already exists '%s'", reflect.TypeOf(t).String())
	}
	service := newTransient(func(c Dic) any { return creator(c) })
	b.b.services[key] = service
	return b
}

// wraps with the smallest id are applied first
// wraps with the same order are applied randomly
func WrapService[T any](b Builder, order Order, wrap func(c Dic, s T) T) Builder {
	key := typeKey[T]()
	wraps := newCtorWrap(order, wrap)

	if _, ok := b.b.wraps[key]; !ok {
		b.b.wraps[key] = make([]ctorWrap, 0, 1)
	}

	b.b.wraps[key] = append(b.b.wraps[key], wraps)
	return b
}

func (b Builder) RegisterScope(scope ScopeID) {
	b.b.scopes[scope] = struct{}{}
}

func (b Builder) LazySingletonLoading() {
	b.b.eagerSingletonLoading = false
}
