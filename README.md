# **ioc - a simple and ergonomic dependency injection container for go**

ioc is a lightweight and opinionated dependency injection (di) container for go,
designed to minimize boilerplate and maximize performance, scalability and developer experience.\
It provides a straightforward way to manage and inject dependencies into your go applications,
promoting loosely coupled and testable code.

## thread safety
This package is thread safe because it uses mutexes but there should be no scenario where this is needed.\
On startup services should be deterministic and initialized in order and during runtime di container isn't used because everything is already wired.

## opinionated choices
### lifetimes
Only service lifetime is singleton.\
DI container should be a dependency manager.\
It's responsible for managing services with their dependencies.

#### Transient services
Transient is just a factory not a separate lifetime.\
There is a built in way to define objects with transient lifetime but its because its just a factory.

#### Scoped services
Scoped services are just data.\
Scoped services usually refer to requestID or connectionID.\
Therefor we should create a singleton service to store these ids.\
Use event bus to emit creation and release of resource with specific id.\
And add other services subscribing to these events and storing all data related to scope id.

### eager loading
All services are eagerly loaded to ensure runtime safety.
If container isn't wired properly application panics.
Its idiomatic because it follows "fail fast" instead of starting with broken service

### reflection
We use reflection instead of compile time for syntax sugar and developer velocity.

### no wraping order
If order is necessary it should be in service.\
If we pre-bake wrapping order into every service even where order doesn't matter then interface becomes convoluted.

## benchmarks

```sh
$ go test . -bench=.
goos: linux
goarch: amd64
pkg: github.com/ogiusek/ioc/v2
cpu: Intel(R) Core(TM) i5-8350U CPU @ 1.70GHz
BenchmarkNewContainerWith3Services-8              382906              3043 ns/op
BenchmarkGet-8                                  58340506                20.61 ns/op
BenchmarkGetServices-8                           5002544               237.3 ns/op
BenchmarkGetInMapWithMutexForComparison-8       54038545                21.72 ns/op
PASS
```

## documentation
### what is package
```go
type Pkg func(b Builder)
```

### package ctors
#### default
```go
func NewPkg(r func(b Builder)) Pkg
```

Example usage.
```go
var Pkg = ioc.NewPkg(func(b ioc.Builder) {
    // `ioc.Register` and/or `ioc.Wrap` calls
}
```

#### parametrized
```go
func NewPkgT[Config any](r func(Builder, Config)) func(Config) Pkg
```

Example usage.
```go
var Pkg = ioc.NewPkgT(func(b ioc.Builder, initial int) {
    // `ioc.Register` and/or `ioc.Wrap` calls
}
```

### service regisration
#### registrations
Registers service `T` and its `ioc.Lazy[T]` getter.
```go
// registers service and its lazy getter with singleton lifetimes
func Register[Service any](b Builder, creator func(c Dic) Service)
```

Example usage.
```go
func _(b ioc.Builder) {
	ioc.Register(b, func(c ioc.Dic) Service {
        // initialize service with its dependencies
		return NewService(/**/)
	})
}
```

#### wrapping
```go
// wraps are applied in addition order after service initialization.
// if there is circular dependency betewen `ServiceA` wrapper and `ServiceB` wrapper one is going to be applied first
func Wrap[Service any](b Builder, wrap func(c Dic, s Service))
```

Example usage.
```go
func _(b ioc.Builder) {
	ioc.Wrap(b, func(c ioc.Dic, s Service) {
        // call anything what should be called on service initialization like
        // Add, Register, Queue
	})
}
```

### service retrieval
#### `GetServices` reccomended
Its most developer friendly approach.\
Its not most performant approach but performance cost is only payed at startup.
```go
// GetServices creates a new instance of type T where T is struct or pointer to a struct and
// injects services into every public property with inject struct tag.
// This function panics when injection fails.
// This is an intentional choice, its go idiomatic because its on startup.
// It ensures application consistency and there is no proper way to handle invalid application wiring.
func GetServices[T any](c Dic) T
```

Example usage.
```go
type Other struct {
    OtherService OtherService `inject:""`
}
type newService struct {
    // we can also inject other structs
    // If struct isn't registered but has fields which can be injected it is injected
    Other `inject:""`
	ServiceA ServiceA `inject:""`
    // On registration also Lazy[T] is automatically registered
    // This allows for circular dependencies
	ServiceB ioc.Lazy[ServiceB] `inject:""`
}
func _(c ioc.Dic) NewService {
    // Parameter T has to be a struct or a pointer to a struct.
    newService := ioc.GetServices[*newService](c)
    // do something with the newService
    return newService
}
```

#### other retrieval methods
These can be used for either more performant access (`Get` is most important) or
for more granural access but `GetServices` is on the biggest level of abstraction and
reduces most boilerplate with least costs.
Only cost is startup time, but it doesn't affect runtime performance.
- `TryGetServices` works like `GetServices` but returns error instead of panicing
- `InjectServices` works like `TryGetServices` but it takes pointer to a struct as an `any` argument
- `Get` retrieves specific service. Panics if service isn't registered
- `TryGet` retrieves specific service. Returns error if service isn't registered
- `Inject` takes pointer to a service and fills it with a service. When service isn't registered returns error

### transients
There is a built in service factory.
```go
// Transient is just a factory which can be registered
type Transient[Service any] func() Service
```

To register a transient you'll need to register `Transient[Service]`

## Contributing
Contact us we are open for suggestions

## License
MIT
