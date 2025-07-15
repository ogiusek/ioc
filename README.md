# **ioc - a simple and ergonomic dependency injection container for go**

ioc is a safe and opinionated dependency injection (di) container for go,
designed to minimize boilerplate and maximize developer experience. it provides
a straightforward way to manage and inject dependencies into your go applications,
promoting loosely coupled and testable code.

This di container embraces lifetimes.

This package is thread safe.

## short explanation

changes from default ioc container and short rules:

1. this code adds lifetime to services represented by function calls
2. scopes do not exists because each service is its own scope (lifetime). function represents scope
3. services are attached to `Dic` upon:
- `Get`
- `WrapService`
Attached services are returned upon retrival attempt.
4. there are different types of attach.
- `ParallelScope` copies all services. this ensures nothing collides if 2 scopes exist at once.
- default attach only attaches removes service before function start and after function end.
5. services do not have to have `Init` method and can be specified using `WrapService`

## code examples

example usage

```go
package main

import (
	"fmt"
	"runtime/debug"

	"github.com/ogiusek/ioc/v2"
)

func main() {
	b := ioc.NewBuilder()

	// errors like circular dependency can be here handled
	ioc.SetErrorHandler(b, func(c ioc.Dic, err error) error {
		stack := debug.Stack()
		fmt.Printf("error: %s\n    stack trace: %s\n\n", err.Error(), string(stack))
		return err
	})

	ioc.AddInit[int](b, func(c ioc.Dic, getter func(c ioc.Dic, service int)) {
		var s int = 1
		ioc.Init(c, &s)
		getter(c, s)
		// clean up
	})

	ioc.AddOnInit[*int32](b, ioc.DefaultOrder, func(c ioc.Dic, service *int32, next func(c ioc.Dic)) {
		*service += 1
		next(c)
		// clean up
	})

	ioc.AddOnInit[*int](b, ioc.DefaultOrder, func(c ioc.Dic, service *int, next func(c ioc.Dic)) {
		ioc.Get(c, func(c ioc.Dic, s int32) {
			*service += int(s)
			next(c)
			// clean up
		})
	})

	ioc.AddInit(b, func(c ioc.Dic, getter func(c ioc.Dic, service int64)) {
		// example circular dependency source
		// ioc.Get(c, func(c ioc.Dic, i int) {
		getter(c, 1)
		// })
	})
	ioc.MarkEagerSingleton[int64](b) // initializes service on application start

	ioc.AddOnInit[*int](b, ioc.DefaultOrder, func(c ioc.Dic, service *int, next func(c ioc.Dic)) {
		ioc.Get(c, func(c ioc.Dic, s int64) {
			*service += int(s)
			next(c)
		})
	})

	// this returns errors like registered init method twice
	errs := ioc.Build(b, func(c ioc.Dic) {
		var i32 int32 = 1
		ioc.Init(c, &i32)
		ioc.WithService(c, i32, func(c ioc.Dic) { // this service is cached and each time anything uses get this service will be returned
			// if getting service faills getter is not called
			// main error handler is called and then error is returned
			// system can be done to never need handling error during getting it
			// but its still needed to handle here error to cleanup somehow else
			/* err := */
			ioc.Get(c, func(c ioc.Dic, i int) {
				fmt.Printf("service int is %d\n", i)
			})
			// do nothing with error
		})
	})
	if len(errs) != 0 {
		panic(fmt.Sprintf("%v\n", errs))
	}
}
```

### builder

```go
var (
	ErrInitMethodAlreadyExists error = errors.New("init method already exists")
)
```

```go
type Order uint

const (
	DefaultOrder Order = iota
)
```

```go
type Builder *builder
```

```go
func NewBuilder() Builder
```

Adds new service initializer.
Initializer is called when `Get` is called and `Service` isn't already present.
```go
func AddInit[Service any](b *builder, init func(c Dic, getter func(c Dic, service Service)))
```

sets function to call when:
- `Get` is called and needs to initialize not yet registered service
```go
// when init is missing program by default panics.
// when init is missing we can:
// - panic
// - ignore getter
// - log that getter is missing
// - call getter with some default implementation
// - return error to be handled by package
func SetMissingInit(b *builder, missingInit func(c Dic, t reflect.Type, getter func(c Dic, service any)) error)
```

adds function to be called on `Init`
```go
func AddOnInit[Service any](b *builder, order Order, onInitListener func(c Dic, service Service, next func(c Dic)))
```

sets function to call when:
- `Init` is called and there is no listeners
```go
// when on init is missing program by default does nothing.
// when on init is missing we can:
// - do nothing
// - do something on init by default
func SetMissingOnInit(b *builder, missingOnInit func(c Dic, t reflect.Type, service any))
```

sets error handler which is called when `Dic` method would return an error
```go
func SetErrorHandler(b *builder, handler func(c Dic, err error) error)
```

initializes `Service` on build.
```go
// on build all eager sinletons are started
func MarkEagerSingleton[Service any](b *builder)
```

when service is used container ensures it can be called in parallel.
```go
// parallel scope copies all initialized services before appending itself.
// normal service just appends itself and removes upon function end but
// parallel scope needs to copy everything to ensure everything matches when other parallel scope is active
func MarkParallelScope[Service any](b *builder)
```

### di container

```go
var (
	ErrCircularDependency error = errors.New("circular dependency. cannot request pending service")
)
```

```go
type Dic *dic
```

it builds everything and checks for errors
```go
// returns all errors which occured during registering.
// can return errors:
// - ErrInitMethodAlreadyExists
func Build(b *builder, getter func(c Dic)) []error
```

attaches service to `Dic` in scope
```go
// specifies that everything inside should use this service
func WithService[Service any](c Dic, service Service, getter func(c Dic))
```

attaches service to `Dic` in scope
```go
// specifies that everything inside should use this service
func WithAnyService(c Dic, service any, getter func(c Dic))
```

calls init listeners
```go
// calls on init listeners
func Init[Service any](c Dic, s Service)
```

calls init listeners
```go
// calls on init listeners
func InitAny(c Dic, s any)
```

gets service in scope.
If service already is attached to `Dic` it just returns service
Else creates service
```go
// gets existing service or tries to initialize service.
// note: onInit isn't called
// this method can return ErrCircularDependency
func Get[Service any](c Dic, getter func(c Dic, service Service)) error
```

gets service in scope.
If service already is attached to `Dic` it just returns service
Else creates service
```go
// gets existing service or tries to initialize service.
// note: onInit isn't called
func GetT(c Dic, t reflect.Type, getter func(c Dic, service any)) error
```

## Contributing

Contact us we are open for suggestions

## License

MIT
