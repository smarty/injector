package injector

import (
	"testing"

	"github.com/smarty/benchy"
	"github.com/smarty/benchy/options"
	. "github.com/smarty/injector/internal/test"
)

//	AVERAGE |    MEDIAN |       MIN |       MAX |   STD DEV |   STD ERR |        4σ
//
// USING MAP:
// 0.390 µs |  0.391 µs |  0.388 µs |  0.393 µs |  1.624 ns |  0.513 ns |  0.397 µs
//
// USING BUBBLE LIST:
// 0.382 µs |  0.382 µs |  0.380 µs |  0.383 µs |  0.580 ns |  0.205 ns |  0.384 µs
func BenchmarkRegistering(b *testing.B) {
	var di *Injector

	benchy.New(b, options.Medium).
		RegisterBenchmark("registering", func() {
			RegisterTransient[Car](di, NewRegularCar)
			RegisterTransient[Driver](di, NewRegularDriver)
		}).
		RegisterSetup("registering", func() {
			di = New()
		}).
		Run()
}

//	AVERAGE  |     MEDIAN |        MIN |        MAX |   STD DEV |   STD ERR |        4σ
//
// USING MAP:
// 0.144 µs  |   0.144 µs |   0.144 µs |   0.146 µs |  0.109 ns |  0.041 ns |  0.146 µs
//
// USING BUBBLE LIST:
// 98.131 ns |  98.048 ns |  97.858 ns |  99.500 ns |  0.252 ns |  0.084 ns |  99.687 ns
func BenchmarkVerify(b *testing.B) {
	var di *Injector

	benchy.New(b, options.Medium).
		RegisterBenchmark("verifying", func() {
			Verify(di)
		}).
		RegisterSetup("verifying", func() {
			di = New()
			RegisterTransient[Car](di, NewRegularCar)
			RegisterTransient[Driver](di, NewRegularDriver)
		}).
		Run()
}

//	AVERAGE |    MEDIAN |       MIN |       MAX |   STD DEV |   STD ERR |        4σ | ALLOCATIONS | MEMORY GROWTH
//
// USING MAP:
// 0.370 µs |  0.370 µs |  0.369 µs |  0.378 µs |  0.625 ns |  0.221 ns |  0.379 µs |       7.000 |         0.031
//
// USING BUBBLE LIST:
// 0.358 µs |  0.358 µs |  0.356 µs |  0.362 µs |  1.025 ns |  0.362 ns |  0.364 µs |       7.000 |         0.030
func BenchmarkGetTransient(b *testing.B) {
	di := New()
	RegisterTransient[Car](di, NewRegularCar)
	RegisterTransient[Driver](di, NewRegularDriver)
	Verify(di)

	benchy.New(b, options.Medium).
		ShowMemoryStats().
		RegisterBenchmark("getting-transient", func() {
			Get[Car](di)
		}, options.OverheadSampling).
		Run()
}

//	AVERAGE  |     MEDIAN |        MIN |        MAX |   STD DEV |   STD ERR |         4σ | ALLOCATIONS | MEMORY GROWTH
//
// USING MAP:
// 20.293 ns |  20.290 ns |  20.214 ns |  20.567 ns |  0.057 ns |  0.019 ns |  20.656 ns |       0.000 |         0.000
//
// USING BUBBLE LIST:
// 15.630 ns |  15.603 ns |  15.495 ns |  15.905 ns |  0.138 ns |  0.044 ns |  16.183 ns |       0.000 |         0.000
func BenchmarkGetSingleton(b *testing.B) {
	di := New()
	RegisterSingleton[Car](di, NewRegularCar)
	RegisterSingleton[Driver](di, NewRegularDriver)
	Verify(di)

	benchy.New(b, options.Medium).
		ShowMemoryStats().
		RegisterBenchmark("getting-singleton", func() {
			Get[Car](di)
		}, options.OverheadSampling).
		Run()
}
