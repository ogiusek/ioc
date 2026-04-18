# **ioc - a simple and ergonomic dependency injection container for go**

ioc is a lightweight and opinionated dependency injection (di) container for go,
designed to minimize boilerplate and maximize performance, scalability and developer experience.\
It provides a straightforward way to manage and inject dependencies into your go applications,
promoting loosely coupled and testable code.

This package is thread safe.

## opinionated choices
### lifetimes
Only service lifetime is singleton.\
DI container should be a dependency manager.\
It's responsible for managing services with their dependencies.\
Scope or transient aren't services. They're data.

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

### reflection
We use reflection instead of compile time for syntax sugar and developer velocity.

### no wraping order
If order is necessary it should be in service.\
If we pre-bake wrapping order into every service even where order doesn't matter then interface becomes convoluted.

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

#### Other methods
These can be used for either more performant access (`Get` is most important) or
for more granural access but `GetServices` is on the biggest level of abstraction and
reduces most boilerplate with least costs.
Only cost is startup time, but it doesn't affect runtime performance.
- `TryGetServices` works like `GetServices` but returns error instead of panicing
- `InjectServices` works like `TryGetServices` but it takes pointer to a struct as an `any` argument
- `Get` retrieves specific service. Panics if service isn't registered
- `TryGet` retrieves specific service. Returns error if service isn't registered
- `Inject` takes pointer to a service and fills it with a service. When service isn't registered returns error

## Contributing
Contact us we are open for suggestions

## License
MIT
