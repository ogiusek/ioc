package ioc_test

import (
	"testing"

	"github.com/ogiusek/ioc"
)

// go test -bench=.

var iterations int = 1000000

func BenchmarkTransientRetreive(b *testing.B) {
	initial := 1
	c := ioc.NewContainer()
	ioc.RegisterTransient(c, func(d ioc.Dic) *int {
		service := initial
		return &service
	})

	b.ResetTimer()
	for i := 0; i <= iterations; i++ {
		ioc.Get[*int](c)
	}
}

func BenchmarkScopedRetrieve(b *testing.B) {
	initial := 1
	c := ioc.NewContainer()
	ioc.RegisterScoped(c, func(d ioc.Dic) *int {
		service := initial
		return &service
	})
	scope := ioc.Scope(c)

	b.ResetTimer()
	for i := 0; i <= iterations; i++ {
		ioc.Get[*int](scope)
	}
}

func BenchmarkScopeCreation(b *testing.B) {
	initial := 1
	c := ioc.NewContainer()
	ioc.RegisterScoped(c, func(d ioc.Dic) *int {
		service := initial
		return &service
	})
	scope := ioc.Scope(c)

	b.ResetTimer()
	for i := 0; i <= iterations; i++ {
		ioc.Get[*int](scope)
	}
}

func BenchmarkSingletonRetrieve(b *testing.B) {
	initial := 1
	c := ioc.NewContainer()
	ioc.RegisterSingleton(c, func(d ioc.Dic) *int {
		service := initial
		return &service
	})

	b.ResetTimer()
	for i := 0; i <= iterations; i++ {
		ioc.Get[*int](c)
	}
}
