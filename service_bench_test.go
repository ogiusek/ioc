package ioc_test

import (
	"reflect"
	"testing"

	"github.com/ogiusek/ioc/v2"
)

// func BenchmarkReflection(b *testing.B) {
// 	for i := 0; i < b.N; i++ {
// 		reflect.TypeFor[int]()
// 	}
// }
//
// func BenchmarkDeepFunction(b *testing.B) {
// 	fun := func() {}
// 	for i := 0; i < 10; i++ {
// 		f := fun
// 		fun = func() { f() }
// 	}
// 	b.ResetTimer()
// 	for i := 0; i < b.N; i++ {
// 		fun()
// 	}
// }
//
// func Generize[T any](fn func(t T)) func(t any) {
// 	return func(t any) { fn(t.(T)) }
// }
//
// func BenchmarkGenericClousure(b *testing.B) {
// 	sum := 0
// 	for i := 0; i < b.N; i++ {
// 		var i int = 0
// 		Generize(func(t int) {
// 			sum += t
// 		})(i)
// 	}
// }
//
// func BenchmarkNestedDeepFunction(b *testing.B) {
// 	fun1 := func() {}
// 	fun2 := func() {
// 		b.ResetTimer()
// 		for i := 0; i < b.N; i++ {
// 			fun1()
// 		}
// 	}
// 	// call
// 	for i := 0; i < 10; i++ {
// 		f1 := fun1
// 		fun1 = func() { f1() }
// 	}
// 	// nest
// 	for i := 0; i < 3000; i++ {
// 		f2 := fun2
// 		if i%100 == 0 {
// 			fun2 = func() { f2() }
// 		} else {
// 			fun2 = func() { f2() }
// 		}
// 	}
// 	fun2()
// }
//
// func BenchmarkAnyAsT(b *testing.B) {
// 	function := func(x any) int {
// 		return x.(int)
// 	}
// 	var x any = (int)(7)
// 	b.ResetTimer()
// 	for i := 0; i < b.N; i++ {
// 		function(x)
// 		// a = x.(int)
// 	}
// }
//
// func BenchmarkCopyingMap(b *testing.B) {
// 	x1 := map[string]string{}
// 	x1["1"] = "1"
// 	// keys := []string{"1"}
// 	b.ResetTimer()
// 	for i := 0; i < b.N; i++ {
// 		x2 := make(map[string]string, len(x1))
// 		// maps.Copy(x2, x1)
// 		// x2 := make(map[string]string, len(x1))
// 		// for _, key := range keys {
// 		// 	x2[key] = x1[key]
// 		// }
// 		for key, value := range x1 {
// 			x2[key] = value
// 		}
// 		x2["1"] = "2"
// 	}
// }

func BenchmarkServiceInit(b *testing.B) {
	b.ReportAllocs()
	builder := ioc.NewBuilder()
	ioc.AddInit(builder, func(c ioc.Dic, onInit func(c ioc.Dic, service int)) {
		var i int = 7
		onInit(c, i)
	})
	ioc.Build(builder, func(c ioc.Dic) {
		b.ResetTimer()

		sum := 0
		for i := 0; i < b.N; i++ {
			ioc.Get(c, func(c ioc.Dic, service int) {
				sum += service
			})
		}
		if sum != b.N*7 {
			b.Errorf("sum != b.N * 7; sum == %d; b.N * 7 == %d;\n", sum, b.N*7)
		}
	})
}

func BenchmarkWithTransientDependencyServiceInit(b *testing.B) {
	b.ReportAllocs()
	builder := ioc.NewBuilder()
	ioc.AddInit(builder, func(c ioc.Dic, onInit func(c ioc.Dic, service int32)) {
		onInit(c, 7)
	})
	ioc.AddInit(builder, func(c ioc.Dic, onInit func(c ioc.Dic, service int)) {
		ioc.Get(c, func(c ioc.Dic, service int32) {
			onInit(c, int(service))
		})
	})
	ioc.Build(builder, func(c ioc.Dic) {
		b.ResetTimer()

		sum := 0
		for i := 0; i < b.N; i++ {
			ioc.Get(c, func(c ioc.Dic, service int) {
				sum += service
			})
		}
		if sum != b.N*7 {
			b.Errorf("sum != b.N * 7; sum == %d; b.N * 7 == %d;\n", sum, b.N*7)
		}
	})
}

func BenchmarkWithSingletonDependencyServiceInit(b *testing.B) {
	b.ReportAllocs()
	builder := ioc.NewBuilder()
	ioc.AddInit(builder, func(c ioc.Dic, onInit func(c ioc.Dic, service int32)) {
		onInit(c, 7)
	})
	ioc.AddInit(builder, func(c ioc.Dic, onInit func(c ioc.Dic, service int)) {
		ioc.Get(c, func(c ioc.Dic, service int32) {
			onInit(c, int(service))
		})
	})
	ioc.Build(builder, func(c ioc.Dic) {
		ioc.Get(c, func(c ioc.Dic, _ int32) {
			b.ResetTimer()

			sum := 0
			for i := 0; i < b.N; i++ {
				ioc.Get(c, func(c ioc.Dic, service int) {
					sum += service
				})
			}
			if sum != b.N*7 {
				b.Errorf("sum != b.N * 7; sum == %d; b.N * 7 == %d;\n", sum, b.N*7)
			}
		})
	})
}

func BenchmarkInitWithSingletons(b *testing.B) {
	b.ReportAllocs()
	builder := ioc.NewBuilder()
	ioc.AddInit(builder, func(c ioc.Dic, onInit func(c ioc.Dic, service float32)) { onInit(c, 0) })
	ioc.MarkEagerSingleton[float32](builder)
	ioc.AddInit(builder, func(c ioc.Dic, onInit func(c ioc.Dic, service float64)) { onInit(c, 0) })
	ioc.MarkEagerSingleton[float64](builder)
	ioc.AddInit(builder, func(c ioc.Dic, onInit func(c ioc.Dic, service int32)) { onInit(c, 0) })
	ioc.MarkEagerSingleton[int32](builder)
	ioc.AddInit(builder, func(c ioc.Dic, onInit func(c ioc.Dic, service int64)) { onInit(c, 0) })
	ioc.MarkEagerSingleton[int64](builder)
	ioc.AddInit(builder, func(c ioc.Dic, onInit func(c ioc.Dic, service uint)) { onInit(c, 0) })
	ioc.MarkEagerSingleton[uint](builder)
	ioc.AddInit(builder, func(c ioc.Dic, onInit func(c ioc.Dic, service uint8)) { onInit(c, 0) })
	ioc.MarkEagerSingleton[uint8](builder)
	ioc.AddInit(builder, func(c ioc.Dic, onInit func(c ioc.Dic, service uint16)) { onInit(c, 0) })
	ioc.MarkEagerSingleton[uint16](builder)
	ioc.AddInit(builder, func(c ioc.Dic, onInit func(c ioc.Dic, service uint32)) { onInit(c, 0) })
	ioc.MarkEagerSingleton[uint32](builder)
	ioc.AddInit(builder, func(c ioc.Dic, onInit func(c ioc.Dic, service uint64)) { onInit(c, 0) })
	ioc.MarkEagerSingleton[uint64](builder)

	ioc.AddInit(builder, func(c ioc.Dic, onInit func(c ioc.Dic, service int)) {
		var i int = 7
		onInit(c, i)
	})
	ioc.Build(builder, func(c ioc.Dic) {
		b.ResetTimer()

		sum := 0
		for i := 0; i < b.N; i++ {
			ioc.Get(c, func(c ioc.Dic, service int) {
				sum += service
			})
		}
		if sum != b.N*7 {
			b.Errorf("sum != b.N * 7; sum == %d; b.N * 7 == %d;\n", sum, b.N*7)
		}
	})
}

func BenchmarkInitListeners(b *testing.B) {
	b.ReportAllocs()
	builder := ioc.NewBuilder()
	for i := 0; i < b.N; i++ {
		ioc.AddOnInit(builder, 0, func(c ioc.Dic, service *int, next func(c ioc.Dic)) {
			*service += 7
			next(c)
		})
	}
	ioc.Build(builder, func(c ioc.Dic) {
		b.ResetTimer()

		var sum int = 0
		ioc.Init(c, &sum)

		if sum != b.N*7 {
			b.Errorf("sum != b.N * 7; sum == %d; b.N * 7 == %d;\n", sum, b.N*7)
		}
	})
}

func BenchmarkManualInit(b *testing.B) {
	b.ReportAllocs()
	builder := ioc.NewBuilder()
	ioc.AddOnInit(builder, 0, func(c ioc.Dic, service *int, next func(c ioc.Dic)) {
		*service += 1
		next(c)
	})
	ioc.Build(builder, func(c ioc.Dic) {
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			ioc.Init(c, i)
		}
	})
}

func BenchmarkGetInitialized(b *testing.B) {
	b.ReportAllocs()
	builder := ioc.NewBuilder()
	ioc.AddInit(builder, func(c ioc.Dic, onInit func(c ioc.Dic, service int)) {
		var i int = 0
		ioc.Init(c, &i)
		onInit(c, i)
	})
	ioc.MarkEagerSingleton[int](builder)
	ioc.AddOnInit(builder, 0, func(c ioc.Dic, service *int, next func(c ioc.Dic)) {
		*service += 7
		next(c)
	})

	ioc.Build(builder, func(c ioc.Dic) {
		b.ResetTimer()

		sum := 0
		for i := 0; i < b.N; i++ {
			ioc.Get(c, func(c ioc.Dic, service int) {
				sum += service
			})
		}
		if sum != b.N*7 {
			b.Errorf("sum != b.N * 7; sum == %d; b.N * 7 == %d;\n", sum, b.N*7)
		}
	})
}

func BenchmarkGetTInitialized(b *testing.B) {
	b.ReportAllocs()
	builder := ioc.NewBuilder()
	ioc.AddInit(builder, func(c ioc.Dic, onInit func(c ioc.Dic, service int)) {
		var i int = 0
		ioc.Init(c, &i)
		onInit(c, i)
	})
	ioc.MarkEagerSingleton[int](builder)
	ioc.AddOnInit(builder, 0, func(c ioc.Dic, service *int, next func(c ioc.Dic)) {
		*service += 7
		next(c)
	})

	ioc.Build(builder, func(c ioc.Dic) {
		t := reflect.TypeFor[int]()
		b.ResetTimer()

		sum := 0
		for i := 0; i < b.N; i++ {
			ioc.GetT(c, t, func(c ioc.Dic, service any) {
				sum += service.(int)
			})
		}
		if sum != b.N*7 {
			b.Errorf("sum != b.N * 7; sum == %d; b.N * 7 == %d;\n", sum, b.N*7)
		}
	})
}

func BenchmarkGetManyExactly2Initialized(b *testing.B) {
	b.ReportAllocs()
	builder := ioc.NewBuilder()
	ioc.AddInit(builder, func(c ioc.Dic, onInit func(c ioc.Dic, service int)) {
		var i int = 0
		ioc.Init(c, &i)
		onInit(c, i)
	})
	ioc.MarkEagerSingleton[int](builder)
	ioc.AddOnInit(builder, 0, func(c ioc.Dic, service *int, next func(c ioc.Dic)) {
		*service += 7
		next(c)
	})

	ioc.Build(builder, func(c ioc.Dic) {
		b.ResetTimer()

		sum := 0
		for i := 0; i < b.N; i++ {
			ioc.GetMany(c, func(service, _ int) {
				sum += service
			})
		}
		if sum != b.N*7 {
			b.Errorf("sum != b.N * 7; sum == %d; b.N * 7 == %d;\n", sum, b.N*7)
		}
	})
}

func BenchmarkGetManyExactly3Initialized(b *testing.B) {
	b.ReportAllocs()
	builder := ioc.NewBuilder()
	ioc.AddInit(builder, func(c ioc.Dic, onInit func(c ioc.Dic, service int)) {
		var i int = 0
		ioc.Init(c, &i)
		onInit(c, i)
	})
	ioc.MarkEagerSingleton[int](builder)
	ioc.AddOnInit(builder, 0, func(c ioc.Dic, service *int, next func(c ioc.Dic)) {
		*service += 7
		next(c)
	})

	ioc.Build(builder, func(c ioc.Dic) {
		b.ResetTimer()

		sum := 0
		for i := 0; i < b.N; i++ {
			ioc.GetMany(c, func(service, _, _ int) {
				sum += service
			})
		}
		if sum != b.N*7 {
			b.Errorf("sum != b.N * 7; sum == %d; b.N * 7 == %d;\n", sum, b.N*7)
		}
	})
}

func BenchmarkGetManyExactly4Initialized(b *testing.B) {
	b.ReportAllocs()
	builder := ioc.NewBuilder()
	ioc.AddInit(builder, func(c ioc.Dic, onInit func(c ioc.Dic, service int)) {
		var i int = 0
		ioc.Init(c, &i)
		onInit(c, i)
	})
	ioc.MarkEagerSingleton[int](builder)
	ioc.AddOnInit(builder, 0, func(c ioc.Dic, service *int, next func(c ioc.Dic)) {
		*service += 7
		next(c)
	})

	ioc.Build(builder, func(c ioc.Dic) {
		b.ResetTimer()

		sum := 0
		for i := 0; i < b.N; i++ {
			ioc.GetMany(c, func(service, _, _, _ int) {
				sum += service
			})
		}
		if sum != b.N*7 {
			b.Errorf("sum != b.N * 7; sum == %d; b.N * 7 == %d;\n", sum, b.N*7)
		}
	})
}

func BenchmarkGetNewScope(b *testing.B) {
	b.ReportAllocs()
	builder := ioc.NewBuilder()
	ioc.AddInit(builder, func(c ioc.Dic, onInit func(c ioc.Dic, service int)) {
		var i int = 0
		ioc.Init(c, &i)
		onInit(c, i)
	})
	ioc.MarkParallelScope[int](builder)
	ioc.AddOnInit(builder, 0, func(c ioc.Dic, service *int, next func(c ioc.Dic)) {
		*service += 7
		next(c)
	})

	ioc.Build(builder, func(c ioc.Dic) {
		sum := 0

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			ioc.Get(c, func(c ioc.Dic, service int) {
				sum += service
			})
		}
		if sum != b.N*7 {
			b.Errorf("sum != b.N * 7; sum == %d; b.N * 7 == %d;\n", sum, b.N*7)
		}
	})
}
