package ioc

import (
	"errors"
	"fmt"
	"reflect"
)

var (
	ErrInitMethodAlreadyExists error = errors.New("init method already exists")
)

//

type Order uint

const (
	DefaultOrder Order = iota
)

//

type onInit struct {
	order  Order
	onInit func(c Dic, service any, next func(c Dic))
}

type initialize func(c Dic, onInit func(c Dic, service any))

//

type builder struct {
	errors []error

	onInit map[reflect.Type][]onInit
	init   map[reflect.Type]initialize

	eagerSingletons map[reflect.Type]struct{}
	parallelScopes  map[reflect.Type]struct{}

	defaultInit   func(c Dic, t reflect.Type, getter func(c Dic, service any)) error
	defaultOnInit func(c Dic, t reflect.Type, service any)
	errorHandler  func(c Dic, err error) error
}

type Builder *builder

func NewBuilder() Builder {
	return &builder{
		errors:          make([]error, 0),
		onInit:          map[reflect.Type][]onInit{},
		init:            map[reflect.Type]initialize{},
		eagerSingletons: map[reflect.Type]struct{}{},
		parallelScopes:  map[reflect.Type]struct{}{},
		defaultInit: func(c Dic, t reflect.Type, getter func(c Dic, service any)) error {
			// cannot try to get not inited function.
			// service didn't get manually initialized or its init method is missing
			return fmt.Errorf("there is no init function for %s\n", t.String())
		},
		defaultOnInit: func(c Dic, t reflect.Type, service any) {
			// do nothing. there can be services which have nothing done on their start
		},
		errorHandler: func(c Dic, err error) error { panic(err.Error()) },
	}
}

func AddInit[Service any](b *builder, init func(c Dic, getter func(c Dic, service Service))) {
	t := reflect.TypeFor[Service]()
	if _, ok := b.init[t]; ok {
		err := errors.Join(
			ErrInitMethodAlreadyExists,
			fmt.Errorf("tried second time to add init method for \"%s\"", t.String()),
		)
		b.errors = append(b.errors, err)
		return
	}

	tInit := func(c Dic, onInit func(c Dic, service any)) {
		init(c, func(c Dic, service Service) { onInit(c, service) })
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
func SetMissingInit(b *builder, missingInit func(c Dic, t reflect.Type, getter func(c Dic, service any)) error) {
	b.defaultInit = missingInit
}

func AddOnInit[Service any](b *builder, order Order, onInitListener func(c Dic, service Service, next func(c Dic))) {
	t := reflect.TypeFor[Service]()
	if _, ok := b.onInit[t]; !ok {
		b.onInit[t] = make([]onInit, 0, 1)
	}
	init := onInit{
		order:  order,
		onInit: func(c Dic, service any, next func(c Dic)) { onInitListener(c, service.(Service), next) },
	}
	b.onInit[t] = append(b.onInit[t], init)
}

// when on init is missing program by default does nothing.
// when on init is missing we can:
// - do nothing
// - do something on init by default
func SetMissingOnInit(b *builder, missingOnInit func(c Dic, t reflect.Type, service any)) {
	b.defaultOnInit = missingOnInit
}

func SetErrorHandler(b *builder, handler func(c Dic, err error) error) {
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
