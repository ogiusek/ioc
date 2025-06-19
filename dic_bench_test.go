package ioc_test

import (
	"sync"
	"testing"

	"github.com/ogiusek/ioc"
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

func BenchmarkGetScoped(b *testing.B) {
	initial := 1
	c := ioc.NewBuilder().
		Wrap(func(b ioc.Builder) ioc.Builder {
			return ioc.RegisterScoped(b, func(c ioc.Dic) int { return initial })
		}).Build()
	scope := ioc.Scope(c)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ioc.Get[int](scope)
	}
}

func BenchmarkScopeCreation(b *testing.B) {
	initial := 1
	c := ioc.NewBuilder().
		Wrap(func(b ioc.Builder) ioc.Builder {
			return ioc.RegisterScoped(b, func(d ioc.Dic) int { return initial })
		}).Build()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ioc.Scope(c)
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
