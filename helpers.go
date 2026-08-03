package ioc

import (
	"log"
	"reflect"
	"slices"
)

// Transient is just a factory which can be registered
type Transient[Service any] func() Service

//

// For every service its lazy getter is automatically registered
type Lazy[Service any] func() Service

//

// pkg is an interface recommended to use
type Pkg func(b Builder)

func NewPkg(r func(b Builder)) Pkg { return r }
func NewPkgT[Config any](r func(Builder, Config)) func(Config) Pkg {
	return func(c Config) Pkg { return func(b Builder) { r(b, c) } }
}

//

// allows to register multiple services in single one
type ServiceRegistry[Key, Service any] interface {
	// panics upon registering copy
	Register(Key, Service)
	// panics upon calling non-existant service
	Get(Key) Service
	Keys() []Key
}

// errors
func AlreadyRegisteredInServiceRegistry[Service any](key any) {
	log.Panicf("registered service '%s' key '%s' already exists", reflect.TypeFor[Service]().String(), key)
}
func MissingKeyInServiceRegistry[Service any](key any) {
	log.Panicf("service of type '%s' with key '%s' is not registered", reflect.TypeFor[Service]().String(), key)
}

// map service registry
type mapServiceRegistry[Key comparable, Service any] struct {
	keys     []Key
	services map[Key]Service
}

func newMapServiceRegistry[Key comparable, Service any]() ServiceRegistry[Key, Service] {
	return &mapServiceRegistry[Key, Service]{services: make(map[Key]Service)}
}

func (r *mapServiceRegistry[Key, Service]) Register(key Key, service Service) {
	if _, ok := r.services[key]; ok {
		AlreadyRegisteredInServiceRegistry[Service](key)
	}
	r.services[key] = service
	r.keys = append(r.keys, key)
}
func (r *mapServiceRegistry[Key, Service]) Get(key Key) Service {
	service, ok := r.services[key]
	if !ok {
		MissingKeyInServiceRegistry[Service](key)
	}
	return service
}
func (r *mapServiceRegistry[Key, Service]) Keys() []Key {
	return slices.Clone(r.keys)
}

func MapServiceRegistryPkg[Key comparable, Service any](b Builder) {
	Register(b, func(c Dic) ServiceRegistry[Key, Service] { return newMapServiceRegistry[Key, Service]() })
}
