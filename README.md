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

	"github.com/ogiusek/ioc"
)

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
func exampleRegistration(c ioc.Builder) ioc.Builder {
	// registering and wraping can be done in any order
	// ioc.Register(Transient, Scoped, Singleton) panics when service is already registered
	return c.
		Wrap(func(b ioc.Builder) ioc.Builder { // wraps not yet registered service
			return ioc.WrapService[ExSingleton](c, func(c ioc.Dic, s ExSingleton) ExSingleton { return s + 1 })
		}).
		Wrap(func(b ioc.Builder) ioc.Builder { // registers service
			return ioc.RegisterSingleton(c, func(c ioc.Dic) ExSingleton { return 7 })
		}).
		Wrap(func(b ioc.Builder) ioc.Builder { // example scoped service registration
			return ioc.RegisterScoped(c, func(c ioc.Dic) ExScoped { return 1 })
		}).
		Wrap(func(b ioc.Builder) ioc.Builder { // example transient
			return ioc.RegisterTransient(c, func(c ioc.Dic) ExTransient { return 1 })
		})
	// currently registered services do not need lifetime because they do not use pointers
}
```

### example scope

```go
func exampleScope(c ioc.Dic) {
	scope := c.Scope() // example scope creation (its useless in current example. its just an example)
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

## Contributing

Contact us we are open for suggestions

## License

MIT
