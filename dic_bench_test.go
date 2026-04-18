package ioc_test

import (
	"sync"
	"testing"

	"github.com/ogiusek/ioc/v2"
)

// go test -bench=.

func BenchmarkInjectSingleton(b *testing.B) {
	initial := 1
	c := ioc.NewContainer(
		func(b ioc.Builder) {
			ioc.Register(b, func(d ioc.Dic) int { return initial })
		},
	)

	b.ResetTimer()
	for b.Loop() {
		_ = ioc.Get[int](c)
	}
}

func BenchmarkGetSingleton(b *testing.B) {
	initial := 1
	c := ioc.NewContainer(
		func(b ioc.Builder) {
			ioc.Register(b, func(d ioc.Dic) int { return initial })
		},
	)

	b.ResetTimer()
	for b.Loop() {
		ioc.Get[int](c)
	}
}

func BenchmarkGetSingletonServices(b *testing.B) {
	type Services struct {
		Service int `inject:"1"`
	}
	c := ioc.NewContainer(
		func(b ioc.Builder) {
			ioc.Register(b, func(c ioc.Dic) int { return 7 })
		},
	)

	b.ResetTimer()
	for b.Loop() {
		ioc.GetServices[Services](c)
	}
}

func BenchmarkMapForComparison(b *testing.B) {
	key := "item"
	testedMap := map[string]int{
		key: 1,
	}

	b.ResetTimer()
	for b.Loop() {
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
	for b.Loop() {
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
	for b.Loop() {
		mutex.Lock()
		_ = (*testedMap)[key]
		mutex.Unlock()
	}
}
