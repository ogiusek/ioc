# **ioc - a simple and ergonomic dependency injection container for go**

ioc is a lightweight and opinionated dependency injection (di) container for go,
designed to minimize boilerplate and maximize developer experience. it provides
a straightforward way to manage and inject dependencies into your go applications,
promoting loosely coupled and testable code.

This package is thread safe.

## opinionated choices
### only service lifetime is singleton
DI container should be a dependency manager.\
It's responsible for managing services with their dependencies.\
Scope or transient aren't services. They're data.

#### Transient services
Transient is just data which should either have public ctor or
have a singleton factory if you want to use wraps for transient.

#### Scoped services
Scoped services are usually services refering to requestID or connectionID.\
Then create a singleton service to store these ids.\
Use event bus to emit resource with id creation and release.\
And add other services subscribing to these events and storing all scopes data.

### no wraping order
If order is necessary it should be in service.\
If we pre-bake wrapping order into every service even where order doesn't matter then interface becomes convoluted.

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

type ExSingleton int
type ExampleServices struct {
	Singleton   ExSingleton `inject:"1"`
	ExSingleton `inject:"1"`
}
```

### example registration

```go
func exampleRegistration(b ioc.Builder) {
	ioc.Wrap[ExSingleton](b, ioc.DefaultOrder, func(c ioc.Dic, s ExSingleton) ExSingleton {
		return ExSingleton(int(s) + int(ioc.Get[ExTransient](c)))
    })
	ioc.Register(b, func(c ioc.Dic) ExSingleton { return 7 })
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
