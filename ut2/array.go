package ut2

import "sync"

type Array[V Comparable] struct {
	arr []V
}

func (a *Array[V]) PushBack(v V) {
	a.arr = append(a.arr, v)
}

func (a *Array[V]) IndexOf(v V) int {
	return IndexOf(a.arr, v)
}

func (a *Array[V]) Len() int {
	return len(a.arr)
}

func (a *Array[V]) RemoveByValue(v V) int {
	return TryRemoveByValue(&a.arr, v)
}

func (a *Array[V]) RemoveByIndex(index int) {
	RemoveByIdx(&a.arr, index)
}

func (a *Array[V]) Each(fn func(V)) {
	for _, v := range a.arr {
		fn(v)
	}
}

////

type ArrayAnySync[V any] struct {
	arr []V
	mu  sync.Mutex
}

func (a *ArrayAnySync[V]) PushBack(v V) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.arr = append(a.arr, v)
}

func (a *ArrayAnySync[V]) Len() int {
	a.mu.Lock()
	defer a.mu.Unlock()

	return len(a.arr)
}

func (a *ArrayAnySync[V]) At(i int) V {
	a.mu.Lock()
	defer a.mu.Unlock()

	return a.arr[i]
}

func (a *ArrayAnySync[V]) Each(fn func(V)) {
	a.mu.Lock()
	defer a.mu.Unlock()

	for _, v := range a.arr {
		fn(v)
	}
}
