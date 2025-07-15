package ioc

import (
	"errors"
	"fmt"
	"reflect"
)

var (
	// package assiciated error
	ErrIoc error = errors.New("ioc error")

	ErrInitMethodAlreadyExists error = errors.Join(ErrIoc, errors.New("init method already exists"))
)

//

type Order uint

const (
	DefaultOrder Order = iota
)

//

type getterTFunc[Service any] func(c Dic, service Service) error
type getterFunc func(c Dic, service any) error

// type initTFunc[Service any] func(c Dic, getter getterTFunc[Service])
// type initFunc func(c Dic, getter getterFunc)
type initTFunc[Service any] func(c Dic, getter func(c Dic, service Service) error) error
type initFunc func(c Dic, getter func(c Dic, service any) error) error

type onInitTFunc[Service any] func(c Dic, service Service, next func(c Dic) error) error
type onInitFunc func(c Dic, service any, next func(c Dic) error) error

type defaultInitFunc func(c Dic, t reflect.Type, getter getterFunc) error
type defaultOnInitFunc func(c Dic, t reflect.Type, service any) error
type errorHandlerFunc func(c Dic, err error) error

type onInit struct {
	order  Order
	onInit onInitFunc
}

type initialize func(c Dic, onInit func(c Dic, service any))

//

type builder struct {
	errors []error

	onInit map[reflect.Type][]onInit
	init   map[reflect.Type]initFunc

	eagerSingletons map[reflect.Type]struct{}
	parallelScopes  map[reflect.Type]struct{}

	defaultInit   defaultInitFunc
	defaultOnInit defaultOnInitFunc
	errorHandler  errorHandlerFunc
}

type Builder *builder

func NewBuilder() Builder {
	return &builder{
		errors:          make([]error, 0),
		onInit:          map[reflect.Type][]onInit{},
		init:            map[reflect.Type]initFunc{},
		eagerSingletons: map[reflect.Type]struct{}{},
		parallelScopes:  map[reflect.Type]struct{}{},
		defaultInit: func(c Dic, t reflect.Type, getter getterFunc) error {
			// cannot try to get not inited function.
			// service didn't get manually initialized or its init method is missing
			return fmt.Errorf("there is no init function for %s\n", t.String())
		},
		defaultOnInit: func(c Dic, t reflect.Type, service any) error {
			// do nothing. there can be services which have nothing done on their start
			return nil
		},
		errorHandler: func(c Dic, err error) error { panic(err.Error()) },
	}
}

func AddInit[Service any](b *builder, init initTFunc[Service]) {
	t := reflect.TypeFor[Service]()
	if _, ok := b.init[t]; ok {
		err := errors.Join(
			ErrInitMethodAlreadyExists,
			fmt.Errorf("tried second time to add init method for \"%s\"", t.String()),
		)
		b.errors = append(b.errors, err)
		return
	}

	var tInit initFunc = func(c Dic, getter func(c Dic, service any) error) error {
		return init(c, func(c Dic, service Service) error { return getter(c, service) })
	}

	b.init[t] = tInit
}

// when init is missing program by default panics.
// when init is missing we can:
// - panic
// - ignore getter
// - log that getter is missing
// - call getter with some default implementation
// - return error to be handled by package
func SetMissingInit(b *builder, missingInit defaultInitFunc) {
	b.defaultInit = missingInit
}

func AddOnInit[Service any](b *builder, order Order, onInitListener onInitTFunc[Service]) {
	t := reflect.TypeFor[Service]()
	if _, ok := b.onInit[t]; !ok {
		b.onInit[t] = make([]onInit, 0, 1)
	}
	init := onInit{
		order: order,
		onInit: func(c Dic, service any, next func(c Dic) error) error {
			return onInitListener(c, service.(Service), next)
		},
	}
	b.onInit[t] = append(b.onInit[t], init)
}

// when on init is missing program by default does nothing.
// when on init is missing we can:
// - do nothing
// - do something on init by default
func SetMissingOnInit(b *builder, missingOnInit defaultOnInitFunc) {
	b.defaultOnInit = missingOnInit
}

func SetContainerErrorHandler(b *builder, handler errorHandlerFunc) {
	b.errorHandler = handler
}

// on build all eager sinletons are started
func MarkEagerSingleton[Service any](b *builder) {
	t := reflect.TypeFor[Service]()
	b.eagerSingletons[t] = struct{}{}
}

// parallel scope copies all initialized services before appending itself.
// normal service just appends itself and removes upon function end but
// parallel scope needs to copy everything to ensure everything matches when other parallel scope is active
func MarkParallelScope[Service any](b *builder) {
	t := reflect.TypeFor[Service]()
	b.parallelScopes[t] = struct{}{}
}
