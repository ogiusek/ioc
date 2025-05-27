package ioc_test

import (
	"testing"

	"github.com/ogiusek/ioc"
)

// go test .

func afterPanic() {
	print("\033[1A") // go 1 line up
	print("\033[2K") // clear line
}

func TestCannotRegisterTwice(t *testing.T) {
	c := ioc.NewContainer()
	ioc.RegisterSingleton(c, func(d ioc.Dic) int { return 1 })

	t.Run("panics", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("function should panic")
			}
			afterPanic()
		}()
		ioc.RegisterSingleton(c, func(d ioc.Dic) int { return 1 })
	})

	t.Run("panics", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("function should panic")
			}
			afterPanic()
		}()
		ioc.RegisterScoped(c, func(d ioc.Dic) int { return 1 })
	})

	t.Run("panics", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("function should panic")
			}
			afterPanic()
		}()
		ioc.RegisterTransient(c, func(d ioc.Dic) int { return 1 })
	})
}

func TestCannotGetNotRegisteredService(t *testing.T) {
	c := ioc.NewContainer()

	t.Run("panics", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("function should panic")
			}
			afterPanic()
		}()
		ioc.Get[int](c)
	})
}
func TestScoped(t *testing.T) {
	initial := 1
	c := ioc.NewContainer()
	ioc.RegisterScoped(c, func(d ioc.Dic) *int {
		service := initial
		return &service
	})

	for i := 0; i <= 5; i++ {
		scope := ioc.Scope(c)
		service := ioc.Get[*int](scope)
		if *service != initial {
			t.Error("scoped service is singleton")
			return
		}
		*service += 1

		service = ioc.Get[*int](scope)
		if *service != initial+1 {
			t.Error("scoped service is transient")
			return
		}
	}
}

func TestSingletion(t *testing.T) {
	initial := 1
	c := ioc.NewContainer()
	ioc.RegisterSingleton(c, func(d ioc.Dic) *int {
		service := initial
		return &service
	})

	for i := 0; i <= 5; i++ {
		scope := ioc.Scope(c)
		service := ioc.Get[*int](scope)
		if *service-i != initial {
			t.Error("singleton service is scoped or transient")
			return
		}
		*service += 1
	}
}

func TestTransient(t *testing.T) {
	initial := 1
	c := ioc.NewContainer()
	ioc.RegisterTransient(c, func(d ioc.Dic) *int {
		service := initial
		return &service
	})

	for i := 0; i <= 5; i++ {
		scope := ioc.Scope(c)
		service := ioc.Get[*int](scope)
		*service += 1
		service = ioc.Get[*int](scope)

		if *service != initial {
			t.Error("transient service is scoped or singleton")
			return
		}
	}
}

func TestServices(t *testing.T) {
	type Services struct {
		Service   int `inject:"1"`
		NoService int
	}

	c := ioc.NewContainer()
	service := 7
	ioc.RegisterTransient(c, func(ioc.Dic) int { return service })

	services := Services{}
	c.InjectServices(&services)

	if services.Service != service {
		t.Error("service isn't properly injected")
	}

	if services.NoService == service {
		t.Error("servie is injected without specifying it")
	}
}
