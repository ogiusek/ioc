package ioc_test

import (
	"sync"
	"testing"

	"github.com/ogiusek/ioc/v2"
)

func BenchmarkNewContainerWith3Services(b *testing.B) {
	pkg := ioc.NewPkg(func(b ioc.Builder) {
		ioc.Register(b, func(c ioc.Dic) int16 { return 0 })
		ioc.Register(b, func(c ioc.Dic) int32 { return 0 })
		ioc.Register(b, func(c ioc.Dic) int64 { return 0 })
	})
	for b.Loop() {
		ioc.NewContainer(pkg)
	}
}

func BenchmarkGet(b *testing.B) {
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

func BenchmarkLazyGet(b *testing.B) {
	initial := 1
	c := ioc.NewContainer(
		func(b ioc.Builder) {
			ioc.Register(b, func(d ioc.Dic) int { return initial })
		},
	)

	b.ResetTimer()
	for b.Loop() {
		ioc.Get[ioc.Lazy[int]](c)()
	}
}

func BenchmarkLazy(b *testing.B) {
	initial := 1
	c := ioc.NewContainer(
		func(b ioc.Builder) {
			ioc.Register(b, func(d ioc.Dic) int { return initial })
		},
	)

	getter := ioc.Get[ioc.Lazy[int]](c)
	b.ResetTimer()
	for b.Loop() {
		getter()
	}
}

func BenchmarkGetServices(b *testing.B) {
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

func BenchmarkGetInMapWithMutexForComparison(b *testing.B) {
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
