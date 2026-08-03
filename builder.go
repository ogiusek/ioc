package ioc

import (
	"log"
	"reflect"
	"sync"
)

var Registry []Pkg

type serviceID any

type builder struct {
	wraps           map[serviceID][]ctorWrap
	services        map[serviceID]service
	servicesOrdered []serviceID
}

type Builder struct {
	b *builder
}

func NewContainer(pkgs ...Pkg) Dic {
	b := Builder{
		b: &builder{
			wraps:    map[serviceID][]ctorWrap{},
			services: map[serviceID]service{},
		},
	}
	registered := map[uintptr]struct{}{}
	for _, pkg := range pkgs {
		k := reflect.ValueOf(pkg).Pointer()
		if _, ok := registered[k]; ok {
			continue
		}
		registered[k] = struct{}{}
		pkg(b)
	}
	return b.build()
}

func (b Builder) build() Dic {
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

// registers service and its lazy getter with singleton lifetimes
func Register[Service any](b Builder, creator func(c Dic) Service) {
	key := typeKey[Service]()
	if _, ok := b.b.services[key]; ok {
		log.Panicf("registered service already exists '%s'", reflect.TypeFor[Service]().String())
	}

	lazyKey := typeKey[Lazy[Service]]()
	if _, ok := b.b.services[lazyKey]; ok {
		log.Panicf("registered service already exists '%s'", reflect.TypeFor[Service]().String())
	}

	b.b.services[key] = newService(func(c Dic) any { return creator(c) })

	b.b.services[lazyKey] = newService(func(c Dic) any {
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

	b.b.servicesOrdered = append(b.b.servicesOrdered, key, lazyKey)
}

// wraps are applied in addition order after service initialization.
// if there is circular dependency betewen `ServiceA` wrapper and `ServiceB` wrapper one is going to be applied first
func Wrap[Service any](b Builder, wrap func(c Dic, s Service)) {
	key := typeKey[Service]()
	wraps := newCtorWrap(wrap)

	if _, ok := b.b.wraps[key]; !ok {
		b.b.wraps[key] = make([]ctorWrap, 0, 1)
	}

	b.b.wraps[key] = append(b.b.wraps[key], wraps)
}
