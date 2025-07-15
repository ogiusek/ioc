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
	ioc.AddInit(builder, func(c ioc.Dic, onInit func(c ioc.Dic, service int) error) error {
		var i int = 7
		return onInit(c, i)
	})
	ioc.Build(builder, func(c ioc.Dic) error {
		b.ResetTimer()

		sum := 0
		for i := 0; i < b.N; i++ {
			ioc.Get(c, func(c ioc.Dic, service int) error {
				sum += service
				return nil
			})
		}
		if sum != b.N*7 {
			b.Errorf("sum != b.N * 7; sum == %d; b.N * 7 == %d;\n", sum, b.N*7)
		}
		return nil
	})
}

func BenchmarkWithTransientDependencyServiceInit(b *testing.B) {
	b.ReportAllocs()
	builder := ioc.NewBuilder()
	ioc.AddInit(builder, func(c ioc.Dic, onInit func(c ioc.Dic, service int32) error) error {
		return onInit(c, 7)
	})
	ioc.AddInit(builder, func(c ioc.Dic, onInit func(c ioc.Dic, service int) error) error {
		return ioc.Get(c, func(c ioc.Dic, service int32) error {
			return onInit(c, int(service))
		})
	})
	ioc.Build(builder, func(c ioc.Dic) error {
		b.ResetTimer()

		sum := 0
		for i := 0; i < b.N; i++ {
			ioc.Get(c, func(c ioc.Dic, service int) error {
				sum += service
				return nil
			})
		}
		if sum != b.N*7 {
			b.Errorf("sum != b.N * 7; sum == %d; b.N * 7 == %d;\n", sum, b.N*7)
		}
		return nil
	})
}

func BenchmarkWithSingletonDependencyServiceInit(b *testing.B) {
	b.ReportAllocs()
	builder := ioc.NewBuilder()
	ioc.AddInit(builder, func(c ioc.Dic, onInit func(c ioc.Dic, service int32) error) error {
		return onInit(c, 7)
	})
	ioc.AddInit(builder, func(c ioc.Dic, onInit func(c ioc.Dic, service int) error) error {
		return ioc.Get(c, func(c ioc.Dic, service int32) error {
			return onInit(c, int(service))
		})
	})
	ioc.Build(builder, func(c ioc.Dic) error {
		ioc.Get(c, func(c ioc.Dic, _ int32) error {
			b.ResetTimer()

			sum := 0
			for i := 0; i < b.N; i++ {
				ioc.Get(c, func(c ioc.Dic, service int) error {
					sum += service
					return nil
				})
			}
			if sum != b.N*7 {
				b.Errorf("sum != b.N * 7; sum == %d; b.N * 7 == %d;\n", sum, b.N*7)
			}
			return nil
		})

		return nil
	})
}

func addSingleton[Service any](b ioc.Builder, service Service) {
	ioc.AddInit(b, func(c ioc.Dic, getter func(c ioc.Dic, service Service) error) error {
		return getter(c, service)
	})
	ioc.MarkEagerSingleton[Service](b)
}

func BenchmarkInitWithSingletons(b *testing.B) {
	b.ReportAllocs()
	builder := ioc.NewBuilder()
	addSingleton[float32](builder, 0)
	addSingleton[float64](builder, 0)
	addSingleton[int32](builder, 0)
	addSingleton[int64](builder, 0)
	addSingleton[uint](builder, 0)
	addSingleton[uint8](builder, 0)
	addSingleton[uint16](builder, 0)
	addSingleton[uint32](builder, 0)
	addSingleton[uint64](builder, 0)

	ioc.AddInit(builder, func(c ioc.Dic, onInit func(c ioc.Dic, service int) error) error {
		var i int = 7
		return onInit(c, i)
	})
	ioc.Build(builder, func(c ioc.Dic) error {
		b.ResetTimer()

		sum := 0
		for i := 0; i < b.N; i++ {
			ioc.Get(c, func(c ioc.Dic, service int) error {
				sum += service
				return nil
			})
		}
		if sum != b.N*7 {
			b.Errorf("sum != b.N * 7; sum == %d; b.N * 7 == %d;\n", sum, b.N*7)
		}
		return nil
	})
}

func BenchmarkInitListeners(b *testing.B) {
	b.ReportAllocs()
	builder := ioc.NewBuilder()
	for i := 0; i < b.N; i++ {
		ioc.AddOnInit(builder, 0, func(c ioc.Dic, service *int, next func(c ioc.Dic) error) error {
			*service += 7
			return next(c)
		})
	}
	ioc.Build(builder, func(c ioc.Dic) error {
		b.ResetTimer()

		var sum int = 0
		ioc.Init(c, &sum)

		if sum != b.N*7 {
			b.Errorf("sum != b.N * 7; sum == %d; b.N * 7 == %d;\n", sum, b.N*7)
		}
		return nil
	})
}

func BenchmarkManualInit(b *testing.B) {
	b.ReportAllocs()
	builder := ioc.NewBuilder()
	ioc.AddOnInit(builder, 0, func(c ioc.Dic, service *int, next func(c ioc.Dic) error) error {
		*service += 1
		return next(c)
	})
	ioc.Build(builder, func(c ioc.Dic) error {
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			ioc.Init(c, i)
		}
		return nil
	})
}

func BenchmarkGetInitialized(b *testing.B) {
	b.ReportAllocs()
	builder := ioc.NewBuilder()
	ioc.AddInit(builder, func(c ioc.Dic, onInit func(c ioc.Dic, service int) error) error {
		var i int = 0
		ioc.Init(c, &i)
		return onInit(c, i)
	})
	ioc.MarkEagerSingleton[int](builder)
	ioc.AddOnInit(builder, 0, func(c ioc.Dic, service *int, next func(c ioc.Dic) error) error {
		*service += 7
		return next(c)
	})

	ioc.Build(builder, func(c ioc.Dic) error {
		b.ResetTimer()

		sum := 0
		for i := 0; i < b.N; i++ {
			ioc.Get(c, func(c ioc.Dic, service int) error {
				sum += service
				return nil
			})
		}
		if sum != b.N*7 {
			b.Errorf("sum != b.N * 7; sum == %d; b.N * 7 == %d;\n", sum, b.N*7)
		}
		return nil
	})
}

func BenchmarkGetTInitialized(b *testing.B) {
	b.ReportAllocs()
	builder := ioc.NewBuilder()
	ioc.AddInit(builder, func(c ioc.Dic, onInit func(c ioc.Dic, service int) error) error {
		var i int = 0
		ioc.Init(c, &i)
		return onInit(c, i)
	})
	ioc.MarkEagerSingleton[int](builder)
	ioc.AddOnInit(builder, 0, func(c ioc.Dic, service *int, next func(c ioc.Dic) error) error {
		*service += 7
		return next(c)
	})

	ioc.Build(builder, func(c ioc.Dic) error {
		t := reflect.TypeFor[int]()
		b.ResetTimer()

		sum := 0
		for i := 0; i < b.N; i++ {
			ioc.GetAny(c, t, func(c ioc.Dic, service any) error {
				sum += service.(int)
				return nil
			})
		}
		if sum != b.N*7 {
			b.Errorf("sum != b.N * 7; sum == %d; b.N * 7 == %d;\n", sum, b.N*7)
		}
		return nil
	})
}

func getManyBenchmark(b *testing.B, getter any) {
	b.ReportAllocs()
	builder := ioc.NewBuilder()
	ioc.AddInit(builder, func(c ioc.Dic, onInit func(c ioc.Dic, service int) error) error {
		var i int = 0
		ioc.Init(c, &i)
		return onInit(c, i)
	})
	ioc.MarkEagerSingleton[int](builder)

	ioc.Build(builder, func(c ioc.Dic) error {
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			ioc.GetMany(c, getter)
		}
		return nil
	})
}

func BenchmarkGetManyExactly1Initialized(b *testing.B) { getManyBenchmark(b, func(s int) {}) }
func BenchmarkGetManyExactly2Initialized(b *testing.B) { getManyBenchmark(b, func(s, _ int) {}) }
func BenchmarkGetManyExactly3Initialized(b *testing.B) { getManyBenchmark(b, func(s, _, _ int) {}) }
func BenchmarkGetManyExactly4Initialized(b *testing.B) { getManyBenchmark(b, func(s, _, _, _ int) {}) }

func getServicesBenchmark[Services any](b *testing.B) {
	b.ReportAllocs()
	builder := ioc.NewBuilder()
	ioc.AddInit(builder, func(c ioc.Dic, onInit func(c ioc.Dic, service int) error) error {
		var i int = 0
		ioc.Init(c, &i)
		return onInit(c, i)
	})
	ioc.MarkEagerSingleton[int](builder)

	ioc.Build(builder, func(c ioc.Dic) error {
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			ioc.GetServices(c, func(c ioc.Dic, s Services) error { return nil })
		}
		return nil
	})
}

func BenchmarkGetServicesExactly1Field(b *testing.B) {
	type Services struct {
		S1 int `inject:"1"`
	}
	getServicesBenchmark[Services](b)
}
func BenchmarkGetServicesExactly2Field(b *testing.B) {
	type Services struct {
		S1 int `inject:"1"`
		S2 int `inject:"1"`
	}
	getServicesBenchmark[Services](b)
}
func BenchmarkGetServicesExactly3Field(b *testing.B) {
	type Services struct {
		S1 int `inject:"1"`
		S2 int `inject:"1"`
		S3 int `inject:"1"`
	}
	getServicesBenchmark[Services](b)
}
func BenchmarkGetServicesExactly4Field(b *testing.B) {
	type Services struct {
		S1 int `inject:"1"`
		S2 int `inject:"1"`
		S3 int `inject:"1"`
		S4 int `inject:"1"`
	}
	getServicesBenchmark[Services](b)
}

func BenchmarkGetNewScope(b *testing.B) {
	b.ReportAllocs()
	builder := ioc.NewBuilder()
	ioc.AddInit(builder, func(c ioc.Dic, onInit func(c ioc.Dic, service int) error) error {
		var i int = 0
		if err := ioc.Init(c, &i); err != nil {
			return err
		}
		return onInit(c, i)
	})
	ioc.MarkParallelScope[int](builder)
	ioc.AddOnInit(builder, 0, func(c ioc.Dic, service *int, next func(c ioc.Dic) error) error {
		*service += 7
		return next(c)
	})

	ioc.Build(builder, func(c ioc.Dic) error {
		sum := 0

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			ioc.Get(c, func(c ioc.Dic, service int) error {
				sum += service
				return nil
			})
		}
		if sum != b.N*7 {
			b.Errorf("sum != b.N * 7; sum == %d; b.N * 7 == %d;\n", sum, b.N*7)
		}
		return nil
	})
}
