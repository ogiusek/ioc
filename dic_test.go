package ioc_test

import (
	"encoding/json"
	"fmt"
	"reflect"
	"runtime/debug"
	"testing"

	"github.com/ogiusek/ioc/v2"
)

// go test .

func afterPanic() {
	print("\033[1A") // go 1 line up
	print("\033[2K") // clear line
}

type ExInterface interface {
	Get() int
}

type ExInterfaceImplementation struct {
	Prop int
}

func (impl *ExInterfaceImplementation) Get() int {
	return impl.Prop
}
func (impl *ExInterfaceImplementation) Error() string {
	return fmt.Sprintf("%d", impl.Prop)
}

func TestContainerForDifferentTypes(t *testing.T) {
	marshal := func(element any) string {
		val, _ := json.Marshal(element)
		return string(val)
	}
	RunContainerTestsForType[int](t, 1, 2, func(a, b int) bool { return a == b })
	{
		a, b := 1, 2
		RunContainerTestsForType[*int](t, &a, &b, func(a, b *int) bool { return a == b })
	}

	RunContainerTestsForType[[]int](t, []int{1}, []int{2}, func(a, b []int) bool { return marshal(a) == marshal(b) })
	RunContainerTestsForType[[1]int](t, [1]int{1}, [1]int{2}, func(a, b [1]int) bool { return marshal(a) == marshal(b) })
	RunContainerTestsForType[map[int]int](t, map[int]int{1: 1}, map[int]int{2: 2}, func(a, b map[int]int) bool { return marshal(a) == marshal(b) })

	RunContainerTestsForType[uintptr](t, uintptr(0x100), uintptr(0x200), func(a, b uintptr) bool { return a == b })
	RunContainerTestsForType[complex64](t, complex(1, 2), complex(3, 4), func(a, b complex64) bool { return a == b })
	RunContainerTestsForType[complex128](t, complex(1.0, 2.0), complex(3.0, 4.0), func(a, b complex128) bool { return a == b })

	RunContainerTestsForType[chan int](t, make(chan int), make(chan int), func(a, b chan int) bool { return a == b })
	RunContainerTestsForType[chan int](t, (chan int)(nil), make(chan int), func(a, b chan int) bool { return a == b })

	RunContainerTestsForType[any](t, 1, 2, func(a, b any) bool { return marshal(a) == marshal(b) })
	RunContainerTestsForType[any](t, 1, "two", func(a, b any) bool { return marshal(a) == marshal(b) })

	{
		val := int(42)
		unsafePtr1 := uintptr(reflect.ValueOf(&val).Pointer())
		unsafePtr2 := uintptr(reflect.ValueOf(new(int)).Pointer())
		RunContainerTestsForType[uintptr](t, unsafePtr1, unsafePtr2, func(a, b uintptr) bool { return a == b })
	}

	{
		RunContainerTestsForType[ExInterface](
			t,
			&ExInterfaceImplementation{Prop: 1},
			&ExInterfaceImplementation{Prop: 2},
			func(a, b ExInterface) bool { return marshal(a) == marshal(b) },
		)
	}

	{
		type WrapperInterface interface{ error }
		a, b := &ExInterfaceImplementation{Prop: 1}, &ExInterfaceImplementation{Prop: 2}
		RunContainerTestsForType[WrapperInterface](t, a, b, func(a, b WrapperInterface) bool { return marshal(a) == marshal(b) })
	}
}

func RunContainerTestsForType[Service any](
	t *testing.T,
	serviceA Service,
	serviceB Service,
	equal func(a, b Service) bool,
) {
	if equal(serviceA, serviceB) {
		el := reflect.TypeOf(&serviceA).Elem()
		t.Errorf("Invalid test arguments for %s", el)
	}

	// this method should be called right after initialization of the container
	testBeforeRegisteredService := func(b ioc.Builder) {
		// test retriving not registered service
		t.Run("panics", func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					afterPanic()
				} else {
					t.Errorf("container should panic when retriving not existing service")
				}
			}()
			ioc.Get[Service](b.Build())
		})

		// test injecting not registered service
		t.Run("panics", func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					afterPanic()
				} else {
					t.Errorf("container shouldn't panic when injecting not existing service: %s\n%s", r, debug.Stack())
				}
			}()
			_ = ioc.Get[Service](b.Build())
		})
	}

	// in this container should be registered service A of any lifetime
	testsOnRegisteredService := func(b ioc.Builder) {
		// test retriving service
		t.Run("panics", func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					afterPanic()
					t.Errorf("container panics when retriving registered service: %s", r)
				}
			}()
			b := b.Clone()

			c := b.Build()
			s := ioc.Get[Service](c)

			if !equal(s, serviceA) {
				t.Errorf("retrieved service is not equal to registered service")
			}
		})

		// test injecting service
		t.Run("panics", func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					afterPanic()
					t.Errorf("container panics when injecting registered service")
				}
			}()
			b := b.Clone()

			c := b.Build()
			service := ioc.Get[Service](c)

			if !equal(service, serviceA) {
				t.Errorf("retrieved service is not equal to registered service")
			}
		})

		// test getting service during resitstration of singleton service
		t.Run("panics", func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					afterPanic()
					t.Errorf("container panics when injecting singleton service requiring service: %s\n%s", r, debug.Stack())
				}
			}()

			b := b.Clone()
			type RequiringService struct{ Service Service }
			ioc.RegisterSingleton(b, func(c ioc.Dic) RequiringService { return RequiringService{Service: serviceA} })
			c := b.Build()
			service := ioc.Get[RequiringService](c)

			if !equal(service.Service, serviceA) {
				t.Errorf("retrieved service is not equal to registered service")
			}
		})

		// test getting service during resitstration of scoped service
		t.Run("panics", func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					afterPanic()
					t.Errorf("container panics when injecting scoped service requiring service: %s\n%s", r, debug.Stack())
				}
			}()

			b := b.Clone()
			scope := ioc.ScopeID("")
			type RequiringService struct{ Service Service }
			ioc.RegisterScoped(b, scope, func(c ioc.Dic) RequiringService { return RequiringService{Service: ioc.Get[Service](c)} })
			b.RegisterScope(scope)
			c := b.Build()
			service := ioc.Get[RequiringService](c)

			if !equal(service.Service, serviceA) {
				t.Errorf("retrieved service is not equal to registered service")
			}
		})

		// test getting service during resitstration of transient service
		t.Run("panics", func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					afterPanic()
					t.Errorf("container panics when injecting transient service requiring service\n%s", debug.Stack())
				}
			}()

			b := b.Clone()
			type RequiringService struct{ Service Service }
			ioc.RegisterTransient(b, func(c ioc.Dic) RequiringService { return RequiringService{Service: ioc.Get[Service](c)} })
			c := b.Build()
			service := ioc.Get[RequiringService](c)

			if !equal(service.Service, serviceA) {
				t.Errorf("retrieved service is not equal to registered service")
			}
		})

		// test injecting services
		t.Run("panics", func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					afterPanic()
					t.Errorf("container panics when injecting registered services: %s\n%s", r, debug.Stack())
				}
			}()

			b := b.Clone()
			type Services struct {
				A Service `inject:"1"`
				B Service `inject:"0"`
				C Service
			}

			c := b.Build()
			services := ioc.GetServices[Services](c)
			var defaultServices Services

			if !equal(services.A, serviceA) {
				t.Errorf("injected service is not equal to registered service")
			}

			if !equal(services.B, defaultServices.B) {
				t.Errorf("injected service is not equal to default service")
			}

			if !equal(services.C, defaultServices.C) {
				t.Errorf("injected service is not equal to default service")
			}
		})

		// test retriving services
		t.Run("panics", func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					afterPanic()
					t.Errorf("container panics when retriving registered services: %s", r)
				}
			}()

			b := b.Clone()
			type Services struct {
				A Service `inject:"1"`
				B Service `inject:"0"`
				C Service
			}

			c := b.Build()
			var defaultServices Services
			services := ioc.GetServices[Services](c)

			if !equal(services.A, serviceA) {
				t.Errorf("retrieved service is not equal to registered service")
			}

			if !equal(services.B, defaultServices.B) {
				t.Errorf("retrieved service is not equal to default service")
			}

			if !equal(services.C, defaultServices.C) {
				t.Errorf("retrieved service is not equal to default service")
			}
		})
	}

	// test universal behaviour (shared for every lifetime)
	// second line is done for each container in case of some funny side effects
	{
		b := ioc.NewBuilder()
		testBeforeRegisteredService(b)
		ioc.RegisterSingleton(b, func(c ioc.Dic) Service { return serviceA })
		testsOnRegisteredService(b)
	}
	{
		b := ioc.NewBuilder()
		scope := ioc.ScopeID("injected scope")
		b.RegisterScope(scope)
		testBeforeRegisteredService(b)
		ioc.RegisterScoped(b, scope, func(c ioc.Dic) Service { return serviceA })
		testsOnRegisteredService(b)
	}
	{
		b := ioc.NewBuilder()
		testBeforeRegisteredService(b)
		ioc.RegisterTransient(b, func(c ioc.Dic) Service { return serviceA })
		testsOnRegisteredService(b)
	}

	register := func(toggler *bool) Service {
		defer func() { *toggler = !*toggler }()
		if !*toggler {
			return serviceA
		}
		return serviceB
	}

	// test lifetime specific behaviour
	{ // test singleton specific behaviour
		b := ioc.NewBuilder()
		testBeforeRegisteredService(b)
		var toggler bool
		b = ioc.RegisterSingleton(b, func(c ioc.Dic) Service { return register(&toggler) })

		// test retriving service
		t.Run("panics", func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					afterPanic()
					t.Errorf("container panics when retriving registered singleton service")
				}
			}()

			test := func(c ioc.Dic) {
				s := ioc.Get[Service](c)

				if !equal(s, serviceA) {
					t.Errorf("singleton service got initialized twice")
				}
			}
			c := b.Build()

			for i := 0; i < 10; i++ {
				test(c)
			}

			for i := 0; i < 10; i++ {
				test(c)
			}
		})
	}
	{ // test scoped specific behaviour
		b := ioc.NewBuilder()
		scope := ioc.ScopeID("")
		testBeforeRegisteredService(b)
		ioc.RegisterSingleton(b, func(c ioc.Dic) *bool {
			toggler := false
			return &toggler
		})
		b.RegisterScope(scope)
		ioc.RegisterScoped(b, scope, func(c ioc.Dic) Service { return register(ioc.Get[*bool](c)) })

		// test retriving service
		t.Run("panics", func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					afterPanic()
					t.Errorf("container panics when retriving registered scoped service: %s", r)
				}
			}()

			expectA := func(c ioc.Dic) {
				s := ioc.Get[Service](c)

				if !equal(s, serviceA) {
					t.Errorf("unexpected scoped service initialization\nexpected %v\ngot %v\n", serviceA, serviceB)
				}
			}

			expectB := func(c ioc.Dic) {
				s := ioc.Get[Service](c)

				if !equal(s, serviceB) {
					t.Errorf("unexpected scoped service initialization\nexpected %v\ngot %v\n", serviceB, serviceA)
				}
			}

			c := b.Build()

			for i := 0; i < 10; i++ {
				if i%2 == 0 {
					c := c.Scope(scope)
					expectA(c)
					expectA(c)
				} else {
					c := c.Scope(scope)
					expectB(c)
					expectB(c)
				}
			}
		})
	}
	{ // test transient spefic behaviour
		b := ioc.NewBuilder()
		scope := ioc.ScopeID("tt")
		testBeforeRegisteredService(b)
		var toggler bool
		b.RegisterScope(scope)
		b = ioc.RegisterTransient(b, func(c ioc.Dic) Service { return register(&toggler) })
		c := b.Build()

		// test retriving service
		t.Run("panics", func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					afterPanic()
					t.Errorf("container panics when retriving registered transient service: %s", r)
				}
			}()

			expectA := func(c ioc.Dic) {
				s := ioc.Get[Service](c)

				if !equal(s, serviceA) {
					t.Errorf("unexpected transient initialization")
				}
			}

			expectB := func(c ioc.Dic) {
				s := ioc.Get[Service](c)

				if !equal(s, serviceB) {
					t.Errorf("unexpected transient initialization")
				}
			}

			for i := 0; i < 10; i++ {
				if i%2 == 0 {
					expectA(c)
				} else {
					expectB(c)
				}
			}

			for i := 0; i < 10; i++ {
				c := c.Scope(scope)
				expectA(c)
				expectB(c)
			}
		})
	}
}

func TestGettingServices(t *testing.T) {
	type Service struct {
		value int
	}
	type Services struct {
		Service `inject:"1"`
	}

	val := 7

	b := ioc.NewBuilder()
	b = ioc.RegisterSingleton(b, func(c ioc.Dic) Service { return Service{value: val} })

	c := b.Build()
	services := ioc.GetServices[Services](c)

	if services.value != val {
		t.Errorf("injected value is not equal to expected")
	}
}

func TestDoubleInjection(t *testing.T) {
	b := ioc.NewBuilder()

	type Service struct{ Val int }
	b = ioc.RegisterSingleton(b, func(c ioc.Dic) Service { return Service{Val: 1} })

	type Wrapper struct{ Service Service }
	b = ioc.RegisterSingleton(b, func(c ioc.Dic) Wrapper { return Wrapper{Service: ioc.Get[Service](c)} })
	c := b.Build()

	wrapper := ioc.Get[Wrapper](c)
	if wrapper.Service.Val != 1 {
		t.Errorf("service inside other service isn't equal to its expected value")
	}
}

func TestRegister(t *testing.T) {
	b := ioc.NewBuilder()
	type Service struct{ Val int }
	ioc.RegisterSingleton(b, func(c ioc.Dic) *Service {
		return &Service{7}
	})
	b.Build()
	c := b.Build()
	service := ioc.Get[*Service](c)
	if service.Val != 7 {
		t.Errorf("unexpected value expected %v and got %v", 7, service.Val)
	}
}

func TestCircularDependencyDetection(t *testing.T) {
	b := ioc.NewBuilder()

	type ServiceA struct{ Val int }
	type ServiceB struct{ Val int }

	ioc.RegisterSingleton(b, func(c ioc.Dic) ServiceA { return ServiceA{ioc.Get[ServiceB](c).Val} })
	ioc.RegisterSingleton(b, func(c ioc.Dic) ServiceB { return ServiceB{ioc.Get[ServiceA](c).Val} })

	t.Run("panics", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				afterPanic()
			}
		}()
		c := b.Build()
		ioc.Get[ServiceA](c)
		t.Errorf("container should panic on circular dependency detenction")
	})
}
