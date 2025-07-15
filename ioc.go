package ioc

import (
	"errors"
	"fmt"
	"reflect"
	"sort"
)

var (
	ErrCircularDependency error = errors.New("circular dependency. cannot request pending service")
)

func Build(b *builder, getter func(c Dic)) []error {
	if len(b.errors) != 0 {
		return b.errors
	}
	c := Dic(&dic{
		onInit:          make(map[reflect.Type]func(c Dic, service any)),
		defaultInit:     b.defaultInit,
		init:            make(map[reflect.Type]func(c Dic, getter func(c Dic, service any))),
		defaultOnInit:   b.defaultOnInit,
		errHandler:      b.errorHandler,
		services:        make(map[reflect.Type]any),
		pendingServices: make(map[reflect.Type]struct{}),
	})

	for serviceType, onInits := range b.onInit {
		sort.Slice(onInits, func(i, j int) bool { return onInits[i].order < onInits[j].order })

		c.onInit[serviceType] = func(c Dic, service any) {
			var onInit func(c Dic)
			for i, oi := range onInits {
				if i == 0 {
					onInit = func(c Dic) { oi.onInit(c, service, func(c Dic) {}) }
					continue
				}
				next := onInit
				onInit = func(c Dic) {
					oi.onInit(c, service, next)
				}
			}
			if onInit != nil {
				onInit(c)
			}
		}
	}

	for serviceType, initializer := range b.init {
		c.init[serviceType] = initializer
	}

	parallelScopes := map[reflect.Type]struct{}{}
	for key, val := range b.parallelScopes {
		parallelScopes[key] = val
	}
	c.parallelScopes = parallelScopes

	var eagerSingletons []reflect.Type
	for t := range b.eagerSingletons {
		eagerSingletons = append(eagerSingletons, t)
	}

	// this service is just to reduce amount of function calls
	load := func(c Dic, service any) {
		getter(c)
	}
	for _, t := range eagerSingletons {
		loadEager := load
		load = func(c Dic, service any) { GetT(c, t, loadEager) }
	}
	load(c, nil)
	return nil
}

type Dic *dic

type dic struct {
	onInit        map[reflect.Type]func(c Dic, service any)
	defaultOnInit func(c Dic, t reflect.Type, service any)

	init        map[reflect.Type]func(c Dic, getter func(c Dic, service any))
	defaultInit func(c Dic, t reflect.Type, getter func(c Dic, service any)) error

	errHandler func(c Dic, err error) error

	parallelScopes map[reflect.Type]struct{}

	services        map[reflect.Type]any
	pendingServices map[reflect.Type]struct{}
}

// specifies that everything inside should use this service
func WithService[Service any](c Dic, service Service, getter func(c Dic)) {
	withServiceT(c, reflect.TypeFor[Service](), service, getter)
}

// specifies that everything inside should use this service
func WithAnyService(c Dic, service any, getter func(c Dic)) {
	withServiceT(c, reflect.TypeOf(service), service, getter)
}

func withServiceT(c Dic, t reflect.Type, service any, getter func(c Dic)) {
	var services map[reflect.Type]any
	services = c.services
	_, isScoped := c.parallelScopes[t]

	if !isScoped {
		services = c.services
	} else {
		services = make(map[reflect.Type]any, len(c.services)+1)
		for key, val := range c.services {
			services[key] = val
		}
	}
	services[t] = service
	getter(Dic(&dic{
		onInit:          c.onInit,
		defaultOnInit:   c.defaultOnInit,
		init:            c.init,
		defaultInit:     c.defaultInit,
		errHandler:      c.errHandler,
		parallelScopes:  c.parallelScopes,
		services:        services,
		pendingServices: c.pendingServices,
	}))
	if !isScoped {
		delete(services, t)
	}
}

// calls on init listeners
func Init[Service any](c Dic, s Service) {
	t := reflect.TypeFor[Service]()
	initT(c, t, s)
}

// calls on init listeners
func InitAny(c Dic, s any) {
	t := reflect.TypeOf(s)
	initT(c, t, s)
}

func initT(c Dic, t reflect.Type, s any) {
	withServiceT(c, t, s, func(c Dic) {
		if onInit, ok := c.onInit[t]; ok {
			onInit(c, s)
			return
		}
		c.defaultOnInit(c, t, s)
	})
}

// gets existing service or tries to initialize service.
// note: onInit isn't called
// this method can return ErrCircularDependency
func Get[Service any](c Dic, getter func(c Dic, service Service)) error {
	t := reflect.TypeFor[Service]()
	if service, ok := c.services[t]; ok {
		getter(c, service.(Service))
		return nil
	}
	if _, pending := c.pendingServices[t]; pending {
		err := errors.Join(
			ErrCircularDependency,
			fmt.Errorf("\"%s\" type has circular dependency", t.String()),
		)
		return c.errHandler(c, err)
	}
	c.pendingServices[t] = struct{}{}
	init, ok := c.init[t]
	onInit := func(c Dic, service any) {
		withServiceT(c, t, service, func(c Dic) { getter(c, service.(Service)) })
	}
	if !ok {
		err := c.defaultInit(c, t, onInit)
		if err != nil {
			return c.errHandler(c, err)
		}
		return nil
	}
	init(c, onInit)
	delete(c.pendingServices, t)
	return nil
}

// gets existing service or tries to initialize service.
// note: onInit isn't called
func GetT(c Dic, t reflect.Type, getter func(c Dic, service any)) error {
	if service, ok := c.services[t]; ok {
		getter(c, service)
		return nil
	}
	if _, pending := c.pendingServices[t]; pending {
		err := errors.Join(
			ErrCircularDependency,
			fmt.Errorf("\"%s\" type has circular dependency", t.String()),
		)
		return c.errHandler(c, err)
	}
	init, ok := c.init[t]
	onInit := func(c Dic, service any) {
		withServiceT(c, t, service, func(c Dic) { getter(c, service) })
	}
	if !ok {
		err := c.defaultInit(c, t, onInit)
		if err != nil {
			return c.errHandler(c, err)
		}
		return nil
	}
	init(c, onInit)
	delete(c.pendingServices, t)
	return nil
}
