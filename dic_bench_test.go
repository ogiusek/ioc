package ioc_test

import (
	"sync"
	"testing"

	"github.com/ogiusek/ioc/v2"
)

// go test -bench=.

func BenchmarkGetTransient(b *testing.B) {
	initial := 1
	c := ioc.NewBuilder().
		Wrap(func(b ioc.Builder) ioc.Builder {
			return ioc.RegisterTransient(b, func(c ioc.Dic) int { return initial })
		}).Build()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ioc.Get[int](c)
	}
}

func BenchmarkGetWrappedTransient(b *testing.B) {
	initial := 1
	builder := ioc.NewBuilder()
	ioc.RegisterTransient(builder, func(c ioc.Dic) int { return initial })
	for i := 0; i < 10; i++ {
		ioc.WrapService(builder, 0, func(c ioc.Dic, s int) int { return s + 1 })
	}
	c := builder.Build()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ioc.Get[int](c)
	}
}

func BenchmarkGetScoped(b *testing.B) {
	scope := ioc.ScopeID("")
	initial := 1
	builder := ioc.NewBuilder()
	ioc.RegisterScoped(builder, ioc.EagerLoading, scope, func(c ioc.Dic) int { return initial })
	builder.RegisterScope(scope)
	c := builder.Build()
	s := c.Scope(scope)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ioc.Get[int](s)
	}
}

func BenchmarkScopeCreation(b *testing.B) {
	scope := ioc.ScopeID("")
	initial := 1
	builder := ioc.NewBuilder()
	ioc.RegisterScoped(builder, ioc.EagerLoading, scope, func(d ioc.Dic) int { return initial })
	builder.RegisterScope(scope)
	c := builder.Build()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Scope(scope)
	}
}

func BenchmarkInjectSingleton(b *testing.B) {
	initial := 1
	c := ioc.NewBuilder().
		Wrap(func(b ioc.Builder) ioc.Builder {
			return ioc.RegisterSingleton(b, func(d ioc.Dic) int { return initial })
		}).Build()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var service int
		c.Inject(&service)
	}
}

func BenchmarkGetSingleton(b *testing.B) {
	initial := 1
	c := ioc.NewBuilder().
		Wrap(func(b ioc.Builder) ioc.Builder {
			return ioc.RegisterSingleton(b, func(d ioc.Dic) int { return initial })
		}).Build()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ioc.Get[int](c)
	}
}

func BenchmarkGetSingletonServices(b *testing.B) {
	type Services struct {
		Service int `inject:"1"`
	}
	c := ioc.NewBuilder().
		Wrap(func(b ioc.Builder) ioc.Builder {
			return ioc.RegisterSingleton(b, func(c ioc.Dic) int { return 7 })
		}).Build()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ioc.GetServices[Services](c)
	}
}

func BenchmarkMapForComparison(b *testing.B) {
	key := "item"
	testedMap := map[string]int{
		key: 1,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = testedMap[key]
	}
}

func BenchmarkMapWithMutexForComparison(b *testing.B) {
	key := "item"
	testedMap := map[string]int{
		key: 1,
	}
	mutex := &sync.Mutex{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mutex.Lock()
		_ = testedMap[key]
		mutex.Unlock()
	}
}

func BenchmarkMapPtrWithMutexForComparison(b *testing.B) {
	key := "item"
	testedMap := &map[string]int{
		key: 1,
	}
	mutex := &sync.Mutex{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mutex.Lock()
		_ = (*testedMap)[key]
		mutex.Unlock()
	}
}
