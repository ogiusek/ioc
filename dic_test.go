package ioc_test

import (
	"encoding/json"
	"errors"
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
		el := reflect.TypeFor[*Service]().Elem()
		t.Errorf("Invalid test arguments for %s", el)
	}

	// this method should be called right after initialization of the container
	testEmptyContainer := func(pkgs ...ioc.Pkg) {
		// test retriving not registered service
		t.Run("panics", func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					afterPanic()
				} else {
					t.Errorf("container should panic when retriving not existing service")
				}
			}()
			_ = ioc.Get[Service](ioc.NewContainer(pkgs...))
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
			_ = ioc.Get[Service](ioc.NewContainer(pkgs...))
		})
	}

	// in this container should be registered service A of any lifetime
	testConteinerWithPkg := func(pkgs ...ioc.Pkg) {
		// test retriving service
		t.Run("panics", func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					afterPanic()
					t.Errorf("container panics when retriving registered service: %s", r)
				}
			}()
			c := ioc.NewContainer(pkgs...)
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

			c := ioc.NewContainer(pkgs...)
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

			type RequiringService struct{ Service Service }
			c := ioc.NewContainer(append(
				pkgs,
				func(b ioc.Builder) {
					ioc.Register(b, func(c ioc.Dic) RequiringService { return RequiringService{Service: serviceA} })
				},
			)...)
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

			type Services struct {
				A Service `inject:""`
				B Service
			}
			c := ioc.NewContainer(pkgs...)
			services := ioc.GetServices[Services](c)
			var defaultServices Services

			if !equal(services.A, serviceA) {
				t.Errorf("injected service is not equal to registered service")
			}

			if !equal(services.B, defaultServices.B) {
				t.Errorf("injected service is not equal to default service")
			}
		})

		t.Run("panics", func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					afterPanic()
					t.Errorf("container panics when injecting registered services: %s\n%s", r, debug.Stack())
				}
			}()

			type Services struct {
				A Service `inject:""`
				B Service
			}

			c := ioc.NewContainer(pkgs...)
			services := *ioc.GetServices[*Services](c)
			var defaultServices Services

			if !equal(services.A, serviceA) {
				t.Errorf("injected service is not equal to registered service")
			}

			if !equal(services.B, defaultServices.B) {
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

			type Services struct {
				A Service `inject:""`
				B Service
			}

			c := ioc.NewContainer(pkgs...)
			var defaultServices Services
			services := ioc.GetServices[Services](c)

			if !equal(services.A, serviceA) {
				t.Errorf("retrieved service is not equal to registered service")
			}

			if !equal(services.B, defaultServices.B) {
				t.Errorf("retrieved service is not equal to default service")
			}
		})
	}

	// test universal behaviour (shared for every lifetime)
	// second line is done for each container in case of some funny side effects
	{
		testEmptyContainer()
		testConteinerWithPkg(
			func(b ioc.Builder) {
				ioc.Register(b, func(c ioc.Dic) Service { return serviceA })
			},
		)
	}

	register := func(toggler *bool) Service {
		defer func() { *toggler = !*toggler }()
		if !*toggler {
			return serviceA
		}
		return serviceB
	}

	var toggler bool
	pkg := ioc.NewPkg(func(b ioc.Builder) {
		ioc.Register(b, func(c ioc.Dic) Service { return register(&toggler) })
	})

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
		c := ioc.NewContainer(pkg)

		for range 10 {
			test(c)
		}

		for range 10 {
			test(c)
		}
	})
}

func TestGettingServices(t *testing.T) {
	type Service struct {
		value int
	}
	type Services struct {
		Service `inject:""`
	}

	val := 7

	c := ioc.NewContainer(func(b ioc.Builder) {
		ioc.Register(b, func(c ioc.Dic) Service { return Service{value: val} })
	})
	services := ioc.GetServices[Services](c)

	if services.value != val {
		t.Errorf("injected value is not equal to expected")
	}
}

func TestInjectServicesError(t *testing.T) {
	type Service struct{}
	type Services struct {
		Service `inject:""`
	}

	c := ioc.NewContainer()
	defer func() {
		r := recover()
		if r != nil {
			afterPanic()
		}
		if r == nil || errors.Is(errors.New(r.(string)), ioc.ErrServiceIsntRegistered) {
			t.Errorf("expected InjectServices to panic when service do not exist and didn't expect %v", r)
		}
	}()
	ioc.GetServices[*Services](c)
}

func TestNestedService(t *testing.T) {
	type Service struct{ Val int }
	type Wrapper struct{ Service Service }

	c := ioc.NewContainer(func(b ioc.Builder) {
		ioc.Register(b, func(c ioc.Dic) Service { return Service{Val: 1} })
		ioc.Register(b, func(c ioc.Dic) Wrapper { return Wrapper{Service: ioc.Get[Service](c)} })
	})

	wrapper := ioc.Get[Wrapper](c)
	if wrapper.Service.Val != 1 {
		t.Errorf("service inside other service isn't equal to its expected value")
	}
}

func TestRegister(t *testing.T) {
	type Service struct{ Val int }
	c := ioc.NewContainer(
		func(b ioc.Builder) {
			ioc.Register(b, func(c ioc.Dic) *Service {
				return &Service{7}
			})
		},
	)
	service := ioc.Get[*Service](c)
	if service.Val != 7 {
		t.Errorf("unexpected value expected %v and got %v", 7, service.Val)
	}
}

func TestCircularDependencyDetection(t *testing.T) {
	type ServiceA struct{ Val int }
	type ServiceB struct{ Val int }

	pkg := ioc.NewPkg(func(b ioc.Builder) {
		ioc.Register(b, func(c ioc.Dic) ServiceA { return ServiceA{ioc.Get[ServiceB](c).Val} })
		ioc.Register(b, func(c ioc.Dic) ServiceB { return ServiceB{ioc.Get[ServiceA](c).Val} })
	})

	t.Run("panics", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				afterPanic()
			}
		}()
		c := ioc.NewContainer(pkg)
		ioc.Get[ServiceA](c)
		t.Errorf("container should panic on circular dependency detenction")
	})
}

func TestCircularDependencyDetectionSafety(t *testing.T) {
	type ServiceA struct{ Val int }
	type ServiceB struct{ Val int }

	c := ioc.NewContainer(
		func(b ioc.Builder) {
			ioc.Register(b, func(c ioc.Dic) ServiceA { return ServiceA{} })
			ioc.Register(b, func(c ioc.Dic) ServiceB { return ServiceB{} })

			ioc.Wrap(b, func(c ioc.Dic, s ServiceA) { ioc.Get[ServiceB](c) })
			ioc.Wrap(b, func(c ioc.Dic, s ServiceB) { ioc.Get[ServiceA](c) })
		},
	)
	ioc.Get[ServiceA](c)
	ioc.Get[ServiceB](c)
}
