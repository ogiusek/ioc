# **ioc - a simple and ergonomic dependency injection container for go**

ioc is a lightweight and opinionated dependency injection (di) container for go,
designed to minimize boilerplate and maximize developer experience. it provides
a straightforward way to manage and inject dependencies into your go applications,
promoting loosely coupled and testable code.

This package is thread safe.

## code examples

### shared code
this code is shared across all code snippets
```go
package main

import (
	"fmt"
	"log"
	"reflect"

	"github.com/ogiusek/ioc/v2"
)

var MyScope ioc.ScopeID = "my scope"

type ExSingleton int
type ExScoped int
type ExTransient int
type ExampleServices struct {
	Singleton   ExSingleton `inject:"1"`
	ExSingleton `inject:"1"`
}
```

### example registration

```go
func exampleRegistration(b ioc.Builder) {
	// registering and wraping can be done in any order
	// ioc.Register(Transient, Scoped, Singleton) panics when service is already registered

    // `Builder.Wrap` just wraps and does nothing
    // example first wrap replace
    // b = ioc.WrapService[ExSingleton](b, ioc.DefaultOrder, func(c ioc.Dic, s ExSingleton) ExSingleton { return s + 1 })

    // can wrap not yet registered service
    // if some service has custom order than default order shouldn't be used
	ioc.WrapService[ExSingleton](b, ioc.DefaultOrder, func(c ioc.Dic, s ExSingleton) ExSingleton {
		return ExSingleton(int(s) + int(ioc.Get[ExTransient](c)))
    })
    // this is optional. it adds compile time safety.
    // when we add dependencies container can panic with ErrCircularDependency or ErrMissingDependency
	ioc.RegisterDependencies[ExSingleton](b, reflect.TypeFor[ExTransient]())
    // other way of adding dependencies
	ioc.RegisterDependency[ExSingleton, ExTransient](b)
    // registers service
	ioc.RegisterSingleton(b, func(c ioc.Dic) ExSingleton { return 7 })
    // example scoped service registration
	ioc.RegisterScoped(b, MyScope, func(c ioc.Dic) ExScoped { return 1 })
    // example transient
	ioc.RegisterTransient(b, func(c ioc.Dic) ExTransient { return 1 })

	// currently registered services do not need lifetime because they do not use pointers
}
```

### example scope

```go
func exampleScope(c ioc.Dic) {
	scope := c.Scope(MyScope) // example scope creation (its useless in current example. its just an example)
	ioc.Get[ExSingleton](scope)
}
```

### example get


```go
func exampleGet(c ioc.Dic) {
	// ways to get service
	// all of these do the same
	{ // ioc.Get[T]
		// wraps ioc.TryGet but panics upon error (ioc.ErrServiceIsntRegistered)
		service := ioc.Get[ExSingleton](c)
		fmt.Printf("%d\n", service) // expected 7 + 1 (8)
	}
	{ // ioc.TryGet[T]
		// this is the fastest way to get service
		// using generics allows for fastest rerival (use go test -bench=.)
		service, err := ioc.TryGet[ExSingleton](c)
		if err != nil {
			panic(err.Error())
		}
		fmt.Printf("%d\n", service) // expected 7 + 1 (8)
	}
	{ // c.Inject
		// this is much slower than ioc.Get[T] and ioc.TryGet[T]
		var service ExSingleton
		if err := c.Inject(&service); err != nil {
			panic(err.Error())
		}
		fmt.Printf("%d\n", service) // expected 7 + 1 (8)
	}
}
```

### example get services

when defining struct specify which properties to inject by using struct tag `inject:"1"`

```go
func exampleGetServices(c ioc.Dic) {
	// ways to get services.
	// this is much slower than injecting manually because this uses reflection
	{ // inject services
		var services ExampleServices
		if err := c.InjectServices(&services); err != nil {
			panic(err.Error())
		}
		fmt.Printf("%d\n", services.Singleton)
	}
	{ // try get services
		services, err := ioc.TryGetServices[ExampleServices](c)
		if err != nil {
			panic(err.Error())
		}
		fmt.Printf("%d\n", services.Singleton)
	}
	{ // get services
		services := ioc.GetServices[ExampleServices](c)
		fmt.Printf("%d\n", services.Singleton)
	}
}
```

### register pkg

```go
type Pkg interface {
	Register(b Builder) Builder
}
```

## notes

all singletons are initialized upon build

## Contributing

Contact us we are open for suggestions

## License

MIT
