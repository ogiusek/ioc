package ioc

import (
	"errors"
	"fmt"
	"reflect"
	"sort"
)

var (
	ErrCircularDependency error = errors.New("circular dependency. cannot request pending service")
	ErrNotAFunction       error = errors.New("argument is not a function")
)

func Build(b *builder, getter func(c Dic) error) []error {
	if len(b.errors) != 0 {
		return b.errors
	}
	c := Dic(&dic{
		onInit:          make(map[reflect.Type]containerOnInitFunc),
		defaultInit:     b.defaultInit,
		init:            make(map[reflect.Type]initFunc),
		defaultOnInit:   b.defaultOnInit,
		errHandler:      b.errorHandler,
		services:        make(map[reflect.Type]any),
		pendingServices: make(map[reflect.Type]struct{}),
	})

	for serviceType, onInits := range b.onInit {
		sort.Slice(onInits, func(i, j int) bool { return onInits[i].order < onInits[j].order })

		c.onInit[serviceType] = func(c Dic, service any) error {
			var onInit func(c Dic) error
			for i, oi := range onInits {
				if i == 0 {
					onInit = func(c Dic) error {
						oi.onInit(c, service, func(c Dic) error { return nil })
						return nil
					}
					continue
				}
				next := onInit
				onInit = func(c Dic) error {
					return oi.onInit(c, service, next)
				}
			}
			if onInit != nil {
				return onInit(c)
			}
			return nil
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
	var load getterFunc = func(c Dic, service any) error {
		return getter(c)
	}
	for _, t := range eagerSingletons {
		loadEager := load
		load = func(c Dic, service any) error { return GetAny(c, t, loadEager) }
	}
	errs := []error{}
	if err := load(c, nil); err != nil {
		errs = append(errs, err)
	}
	return errs
}

var dicType = reflect.TypeFor[Dic]()

type containerOnInitFunc func(c Dic, service any) error

type Dic *dic

type dic struct {
	// add shared
	onInit        map[reflect.Type]containerOnInitFunc
	defaultOnInit defaultOnInitFunc

	init        map[reflect.Type]initFunc
	defaultInit defaultInitFunc

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
func Init[Service any](c Dic, s Service) error {
	t := reflect.TypeFor[Service]()
	return initT(c, t, s)
}

// calls on init listeners
func InitAny(c Dic, s any) error {
	t := reflect.TypeOf(s)
	return initT(c, t, s)
}

func initT(c Dic, t reflect.Type, s any) error {
	var err error
	withServiceT(c, t, s, func(c Dic) {
		if onInit, ok := c.onInit[t]; ok {
			err = onInit(c, s)
			return
		}
		err = c.defaultOnInit(c, t, s)
	})
	return err
}

// gets existing service or tries to initialize service.
// note: onInit isn't called
// this method can return ErrCircularDependency
func Get[Service any](c Dic, getter func(c Dic, service Service) error) error {
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
	onInit := func(c Dic, service any) error {
		var err error
		withServiceT(c, t, service, func(c Dic) { err = getter(c, service.(Service)) })
		return err
	}
	if !ok {
		err := c.defaultInit(c, t, onInit)
		if err != nil {
			return c.errHandler(c, err)
		}
		return nil
	}
	err := init(c, onInit)
	delete(c.pendingServices, t)
	return err
}

// gets existing service or tries to initialize service.
// note: onInit isn't called
func GetAny(c Dic, t reflect.Type, getter getterFunc) error {
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
	var onInit getterFunc = func(c Dic, service any) error {
		var err error
		withServiceT(c, t, service, func(c Dic) { err = getter(c, service) })
		return err
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

// gets existing services or tries to initialize service.
// note: onInit isn't called
// example getter: `func(c Dic, serviceA Service, serviceB ...)`
// argument can be any service or `Dic`
func GetMany(c Dic, getter any) error {
	getterValue := reflect.ValueOf(getter)
	getterType := getterValue.Type()

	if getterType.Kind() != reflect.Func {
		return errors.Join(
			ErrNotAFunction,
			fmt.Errorf("argument type is \"%s\"", getterType.String()),
		)
	}

	numArgs := getterType.NumIn()
	in := make([]reflect.Value, numArgs)
	dicIndicies := []int{}

	call := func(c Dic) error {
		cValue := reflect.ValueOf(c)
		for _, i := range dicIndicies {
			in[i] = cValue
		}
		results := getterValue.Call(in)
		if len(results) != 1 {
			return nil
		}
		res := results[0]
		if err, ok := res.Interface().(error); ok {
			return err
		}
		return nil
	}

	for i := 0; i < numArgs; i++ {
		argType := getterType.In(i)

		if argType == dicType {
			dicIndicies = append(dicIndicies, i)
			continue
		}

		prevCall := call
		call = func(c Dic) error {
			return GetAny(c, argType, func(c Dic, service any) error {
				in[i] = reflect.ValueOf(service)
				return prevCall(c)
			})
		}
	}

	if err := call(c); err != nil {
		return c.errHandler(c, err)
	}
	return nil
}
