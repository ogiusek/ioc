package ioc

import (
	"log"
	"maps"
	"reflect"
	"sync"
)

type serviceID any

type builder struct {
	wraps           map[serviceID][]ctorWrap
	services        map[serviceID]Service
	servicesOrdered []serviceID
}

type Builder struct {
	b *builder
}

func NewBuilder() Builder {
	return Builder{
		b: &builder{
			wraps:    map[serviceID][]ctorWrap{},
			services: map[serviceID]Service{},
		},
	}
}

func (b Builder) Clone() Builder {
	clonedB := Builder{
		b: &builder{
			wraps:    make(map[serviceID][]ctorWrap, len(b.b.wraps)),
			services: nil,
		},
	}
	for key, val := range b.b.wraps {
		wraps := make([]ctorWrap, len(val))
		copy(wraps, val)
		clonedB.b.wraps[key] = wraps
	}
	clonedB.b.services = maps.Clone(b.b.services)
	return clonedB
}

func (b Builder) Build() Dic {
	services := b.b.services
	for key, service := range services {
		wraps, ok := b.b.wraps[key]
		if !ok || len(wraps) == 0 {
			continue
		}
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

			creationMapMutex: sync.Mutex{},
			creationMap:      make(map[serviceID]struct{}),
		},
	}
	for _, key := range b.b.servicesOrdered {
		service := b.b.services[key]
		if *service.instance != nil {
			continue
		}
		c.tryLock(key)
		instance := service.creator(c)
		*service.instance = instance
		b.b.services[key] = service
		c.unlock(key)
		service.wraps(c, instance)
	}
	return c
}

func Register[Service any](b Builder, creator func(c Dic) Service) {
	key := typeKey[Service]()
	if _, ok := b.b.services[key]; ok {
		var t Service
		log.Panicf("registered service already exists '%s'", reflect.TypeOf(t).String())
	}
	service := newSingleton(func(c Dic) any { return creator(c) })
	b.b.services[key] = service

	lazyKey := typeKey[Lazy[Service]]()
	lazyService := newSingleton(func(c Dic) any {
		var service Service
		ok := false
		var lazy Lazy[Service] = func() Service {
			if ok {
				return service
			}
			service = Get[Service](c)
			ok = true
			return service
		}
		return lazy
	})
	b.b.services[lazyKey] = lazyService

	b.b.servicesOrdered = append(b.b.servicesOrdered, key, lazyKey)
}

// wraps with the smallest id are applied first
// wraps with the same order are applied randomly
func Wrap[Service any](b Builder, wrap func(c Dic, s Service)) {
	key := typeKey[Service]()
	wraps := newCtorWrap(wrap)

	if _, ok := b.b.wraps[key]; !ok {
		b.b.wraps[key] = make([]ctorWrap, 0, 1)
	}

	b.b.wraps[key] = append(b.b.wraps[key], wraps)
}
