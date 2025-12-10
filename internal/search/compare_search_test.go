package search

import (
	"reflect"
	"testing"

	"github.com/smarty/benchy"
	"github.com/smarty/benchy/options"
	"github.com/smarty/benchy/providers"
)

func BenchmarkCompareSearches(b *testing.B) {
	provider := providers.
		New1(func(reflect.Type) {}).
		Add(reflect.TypeFor[S]()).
		Add(reflect.TypeFor[A]()).
		Add(reflect.TypeFor[V]()).
		Add(reflect.TypeFor[T]()).
		Add(reflect.TypeFor[S]()).
		Add(reflect.TypeFor[A]()).
		Add(reflect.TypeFor[W]()).
		Add(reflect.TypeFor[S]()).
		Add(reflect.TypeFor[A]()).
		Add(reflect.TypeFor[B]()).
		Add(reflect.TypeFor[Q]())

	bubbleList := new(BubbleList[reflect.Type, int])
	bubbleList.Add(reflect.TypeFor[A](), 0)
	bubbleList.Add(reflect.TypeFor[B](), 0)
	bubbleList.Add(reflect.TypeFor[C](), 0)
	bubbleList.Add(reflect.TypeFor[D](), 0)
	bubbleList.Add(reflect.TypeFor[E](), 0)
	bubbleList.Add(reflect.TypeFor[F](), 0)
	bubbleList.Add(reflect.TypeFor[G](), 0)
	bubbleList.Add(reflect.TypeFor[H](), 0)
	bubbleList.Add(reflect.TypeFor[I](), 0)
	bubbleList.Add(reflect.TypeFor[J](), 0)
	bubbleList.Add(reflect.TypeFor[K](), 0)
	bubbleList.Add(reflect.TypeFor[L](), 0)
	bubbleList.Add(reflect.TypeFor[M](), 0)
	bubbleList.Add(reflect.TypeFor[N](), 0)
	bubbleList.Add(reflect.TypeFor[O](), 0)
	bubbleList.Add(reflect.TypeFor[P](), 0)
	bubbleList.Add(reflect.TypeFor[Q](), 0)
	bubbleList.Add(reflect.TypeFor[R](), 0)
	bubbleList.Add(reflect.TypeFor[S](), 0)
	bubbleList.Add(reflect.TypeFor[T](), 0)
	bubbleList.Add(reflect.TypeFor[U](), 0)
	bubbleList.Add(reflect.TypeFor[V](), 0)
	bubbleList.Add(reflect.TypeFor[W](), 0)
	bubbleList.Add(reflect.TypeFor[X](), 0)
	bubbleList.Add(reflect.TypeFor[Y](), 0)
	bubbleList.Add(reflect.TypeFor[Z](), 0)

	myMap := NewMap[reflect.Type, int]()
	myMap.Add(reflect.TypeFor[A](), 0)
	myMap.Add(reflect.TypeFor[B](), 0)
	myMap.Add(reflect.TypeFor[C](), 0)
	myMap.Add(reflect.TypeFor[D](), 0)
	myMap.Add(reflect.TypeFor[E](), 0)
	myMap.Add(reflect.TypeFor[F](), 0)
	myMap.Add(reflect.TypeFor[G](), 0)
	myMap.Add(reflect.TypeFor[H](), 0)
	myMap.Add(reflect.TypeFor[I](), 0)
	myMap.Add(reflect.TypeFor[J](), 0)
	myMap.Add(reflect.TypeFor[K](), 0)
	myMap.Add(reflect.TypeFor[L](), 0)
	myMap.Add(reflect.TypeFor[M](), 0)
	myMap.Add(reflect.TypeFor[N](), 0)
	myMap.Add(reflect.TypeFor[O](), 0)
	myMap.Add(reflect.TypeFor[P](), 0)
	myMap.Add(reflect.TypeFor[Q](), 0)
	myMap.Add(reflect.TypeFor[R](), 0)
	myMap.Add(reflect.TypeFor[S](), 0)
	myMap.Add(reflect.TypeFor[T](), 0)
	myMap.Add(reflect.TypeFor[U](), 0)
	myMap.Add(reflect.TypeFor[V](), 0)
	myMap.Add(reflect.TypeFor[W](), 0)
	myMap.Add(reflect.TypeFor[X](), 0)
	myMap.Add(reflect.TypeFor[Y](), 0)
	myMap.Add(reflect.TypeFor[Z](), 0)

	priorityList := new(PriorityList[reflect.Type, int])
	priorityList.Add(reflect.TypeFor[A](), 0)
	priorityList.Add(reflect.TypeFor[B](), 0)
	priorityList.Add(reflect.TypeFor[C](), 0)
	priorityList.Add(reflect.TypeFor[D](), 0)
	priorityList.Add(reflect.TypeFor[E](), 0)
	priorityList.Add(reflect.TypeFor[F](), 0)
	priorityList.Add(reflect.TypeFor[G](), 0)
	priorityList.Add(reflect.TypeFor[H](), 0)
	priorityList.Add(reflect.TypeFor[I](), 0)
	priorityList.Add(reflect.TypeFor[J](), 0)
	priorityList.Add(reflect.TypeFor[K](), 0)
	priorityList.Add(reflect.TypeFor[L](), 0)
	priorityList.Add(reflect.TypeFor[M](), 0)
	priorityList.Add(reflect.TypeFor[N](), 0)
	priorityList.Add(reflect.TypeFor[O](), 0)
	priorityList.Add(reflect.TypeFor[P](), 0)
	priorityList.Add(reflect.TypeFor[Q](), 0)
	priorityList.Add(reflect.TypeFor[R](), 0)
	priorityList.Add(reflect.TypeFor[S](), 0)
	priorityList.Add(reflect.TypeFor[T](), 0)
	priorityList.Add(reflect.TypeFor[U](), 0)
	priorityList.Add(reflect.TypeFor[V](), 0)
	priorityList.Add(reflect.TypeFor[W](), 0)
	priorityList.Add(reflect.TypeFor[X](), 0)
	priorityList.Add(reflect.TypeFor[Y](), 0)
	priorityList.Add(reflect.TypeFor[Z](), 0)

	benchy.New(b, options.Medium).
		RegisterBenchmark("priority-list", provider.WrapBenchmarkFunc(func(t reflect.Type) {
			priorityList.Find(t, Reorder)
		})).
		RegisterBenchmark("bubble-list", provider.WrapBenchmarkFunc(func(t reflect.Type) {
			bubbleList.Find(t, Reorder)
		})).
		RegisterBenchmark("map", provider.WrapBenchmarkFunc(func(t reflect.Type) {
			myMap.Find(t, Reorder)
		})).
		Run()
}

// ---------------- some types -------------------

type A struct {
}

type B struct {
}

type C struct {
}

type D struct {
}

type E struct {
}

type F struct {
}

type G struct {
}

type H struct {
}

type I struct {
}

type J struct {
}

type K struct {
}

type L struct {
}

type M struct {
}

type N struct {
}

type O struct {
}

type P struct {
}

type Q struct {
}

type R struct {
}

type S struct {
}

type T struct {
}

type U struct {
}

type V struct {
}

type W struct {
}

type X struct {
}

type Y struct {
}

type Z struct {
}
