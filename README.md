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
	ioc.SetContainerErrorHandler(b, func(c ioc.Dic, err error) error {
		stack := debug.Stack()
		fmt.Printf("error: %s\n    stack trace: %s\n\n", err.Error(), string(stack))
		return err
	})

	ioc.AddInit[int](b, func(c ioc.Dic, getter func(c ioc.Dic, service int) error) error {
		var s int = 1
		if err := ioc.Init(c, &s); err != nil {
			return err
		}
		err := getter(c, s)
		// clean up
		return err
	})

	ioc.AddOnInit[*int32](b, ioc.DefaultOrder, func(c ioc.Dic, service *int32, next func(c ioc.Dic) error) error {
		*service += 1
		err := next(c)
		// clean up
		return err
	})

	ioc.AddOnInit[*int](b, ioc.DefaultOrder, func(c ioc.Dic, service *int, next func(c ioc.Dic) error) error {
		return ioc.Get(c, func(c ioc.Dic, s int32) error {
			*service += int(s)
			err := next(c)
			// clean up
			return err
		})
	})

	ioc.AddInit(b, func(c ioc.Dic, getter func(c ioc.Dic, service int64) error) error {
		// example circular dependency source
		// ioc.Get(c, func(c ioc.Dic, i int) {
		return getter(c, 64)
		// })
	})
	ioc.MarkEagerSingleton[int64](b) // initializes service on application start

	ioc.AddOnInit[*int](b, ioc.DefaultOrder, func(c ioc.Dic, service *int, next func(c ioc.Dic) error) error {
		return ioc.Get(c, func(c ioc.Dic, s int64) error {
			*service += int(s)
			return next(c)
		})
	})

	// this returns errors like registered init method twice
	errs := ioc.Build(b, func(c ioc.Dic) error {
		var i32 int32 = 32
		if err := ioc.Init(c, &i32); err != nil {
			return err
		}
		ioc.WithService(c, i32, func(c ioc.Dic) { // this service is cached and each time anything uses get this service will be returned
			// if getting service faills getter is not called
			// main error handler is called and then error is returned
			// system can be done to never need handling error during getting it
			// but its still needed to handle here error to cleanup somehow else
			/* err := */
			ioc.Get(c, func(c ioc.Dic, i int) error {
				fmt.Printf("service int is %d\n", i)
				return nil
			})
			// do nothing with error

			ioc.GetMany(c, func(c ioc.Dic, i64 int64, i32 int32, i int) {
				fmt.Printf("int64 %d; int32 %d; int %d\n", i64, i32, i)
			})
		})

		return nil
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
func AddInit[Service any](b *builder, init func(c Dic, getter func(c Dic, service Service) error) error)
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
func SetMissingInit(b *builder, missingInit func(c Dic, t reflect.Type, getter func(c Dic, service any) error) error)
```

adds function to be called on `Init`
```go
func AddOnInit[Service any](b *builder, order Order, onInitListener func(c Dic, service Service, next func(c Dic) error) error)
```

sets function to call when:
- `Init` is called and there is no listeners
```go
// when on init is missing program by default does nothing.
// when on init is missing we can:
// - do nothing
// - do something on init by default
func SetMissingOnInit(b *builder, missingOnInit func(c Dic, t reflect.Type, service any) error)
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
func Build(b *builder, getter func(c Dic) error) []error
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
func Init[Service any](c Dic, s Service) error
```

calls init listeners
```go
// calls on init listeners
func InitAny(c Dic, s any) error
```

gets service in scope.
If service already is attached to `Dic` it just returns service
Else creates service
```go
// gets existing service or tries to initialize service.
// note: onInit isn't called
// this method can return ErrCircularDependency
func Get[Service any](c Dic, getter func(c Dic, service Service) error) error
```

gets service in scope.
If service already is attached to `Dic` it just returns service
Else creates service
```go
// gets existing service or tries to initialize service.
// note: onInit isn't called
func GetT(c Dic, t reflect.Type, getter func(c Dic, service any) error) error
```

calls getter where arguments are injected from container.
if `Dic` is requested then its injected not as service but as latest list of services.
```go
// gets existing services or tries to initialize service.
// note: onInit isn't called
// example getter: `func(c Dic, serviceA Service, serviceB ...)`
// argument can be any service or `Dic`
// if getter returns only error its returned by GetMany
func GetMany(c Dic, getter any) error
```

## Contributing

Contact us we are open for suggestions

## License

MIT
