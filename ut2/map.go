package ut2

import (
	"container/list"
	"sync"
	"time"
)

type IMap[K Comparable, V any] interface {
	Load(k K) (V, bool)
	Store(k K, v V)
	Delete(k K)
	Clear()
	Len() int
	Each(fn func(k K, v V))
}

//

type FakeMap[K Comparable, V any] struct {
	f V
}

func NewFakeMap[K Comparable, V any]() *FakeMap[K, V] {
	return &FakeMap[K, V]{}
}

func (m *FakeMap[K, V]) Load(k K) (V, bool)     { return m.f, false }
func (m *FakeMap[K, V]) Store(k K, v V)         {}
func (m *FakeMap[K, V]) Delete(k K)             {}
func (m *FakeMap[K, V]) Clear()                 {}
func (m *FakeMap[K, V]) Len() int               { return 0 }
func (m *FakeMap[K, V]) Each(fn func(k K, v V)) {}

//

type Map[K Comparable, V any] struct {
	m map[K]V
}

func NewMap[K Comparable, V any]() *Map[K, V] { return &Map[K, V]{m: map[K]V{}} }

func (m *Map[K, V]) Load(k K) (V, bool) {
	v, ok := m.m[k]
	return v, ok
}

func (m *Map[K, V]) Store(k K, v V) {
	m.m[k] = v
}

func (m *Map[K, V]) Delete(k K) {
	delete(m.m, k)
}

func (m *Map[K, V]) Clear() {
	m.m = map[K]V{}
}

func (m *Map[K, V]) Len() int {
	return len(m.m)
}

func (m *Map[K, V]) Each(fn func(k K, v V)) {
	for k, v := range m.m {
		fn(k, v)
	}
}

//

type SyncMap[K Comparable, V any] struct {
	m    map[K]V
	lock sync.Mutex
}

func NewSyncMap[K Comparable, V any]() *SyncMap[K, V] {
	return &SyncMap[K, V]{m: map[K]V{}}
}

func (m *SyncMap[K, V]) Load(k K) (V, bool) {
	m.lock.Lock()
	defer m.lock.Unlock()

	v, ok := m.m[k]
	return v, ok
}

func (m *SyncMap[K, V]) Store(k K, v V) {
	m.lock.Lock()
	defer m.lock.Unlock()

	m.m[k] = v
}

func (m *SyncMap[K, V]) Delete(k K) {
	m.lock.Lock()
	defer m.lock.Unlock()

	delete(m.m, k)
}

func (m *SyncMap[K, V]) Clear() {
	m.lock.Lock()
	defer m.lock.Unlock()

	m.m = map[K]V{}
}

func (m *SyncMap[K, V]) Len() int {
	m.lock.Lock()
	defer m.lock.Unlock()

	return len(m.m)
}

func (m *SyncMap[K, V]) Each(fn func(k K, v V)) {
	m.lock.Lock()
	defer m.lock.Unlock()

	for k, v := range m.m {
		fn(k, v)
	}
}

// 这个map 用于缓存活跃数据。
// 不活跃的数据将自动删除

type ActiveMap[K Comparable, V any] struct {
	lst  *list.List
	mp   map[K]*activeNode[K, V]
	lock sync.Mutex
}

type activeNode[K Comparable, V any] struct {
	activeTime time.Time
	k          K
	v          V
	ele        *list.Element
}

// inteval  检查过期的时钟周期
// lifetime 不活跃数据生命时长，超过将删除
func NewActiveMap[K Comparable, V any](inteval, lifetime time.Duration) *ActiveMap[K, V] {
	r := &ActiveMap[K, V]{mp: map[K]*activeNode[K, V]{}, lst: list.New()}

	go r.autoDelete(inteval, lifetime)

	return r
}

func (m *ActiveMap[K, V]) active(one *activeNode[K, V]) {
	one.activeTime = time.Now()
	m.lst.MoveToBack(one.ele)
}

func (m *ActiveMap[K, V]) autoDelete(inteval, lifetime time.Duration) {
	t := time.NewTicker(inteval)
	for {

		select {
		case <-t.C:
			m.DeleteInactive(lifetime)
		}
	}
}

func (m *ActiveMap[K, V]) DeleteInactive(lifetime time.Duration) {
	m.lock.Lock()
	defer m.lock.Unlock()

	head := m.lst.Front()
	now := time.Now()

	for head != m.lst.Back() {
		val := head.Value.(*activeNode[K, V])

		if now.Sub(val.activeTime) < lifetime {
			return
		}

		tpl := head

		head = head.Next()

		// fmt.Println("delete", val.k)

		delete(m.mp, val.k)
		m.lst.Remove(tpl)
	}
}

func (m *ActiveMap[K, V]) Load(k K) (v V, ok bool) {
	m.lock.Lock()
	defer m.lock.Unlock()

	one, ok := m.mp[k]
	if ok {
		v = one.v
		m.active(one)
	}

	return
}

func (m *ActiveMap[K, V]) Store(k K, v V) {
	// list 去重
	m.Delete(k)

	m.lock.Lock()
	defer m.lock.Unlock()

	one := &activeNode[K, V]{
		activeTime: time.Now(),
		v:          v,
		k:          k,
	}

	ele := m.lst.PushBack(one)
	one.ele = ele

	m.mp[k] = one
}

func (m *ActiveMap[K, V]) Delete(k K) {
	m.lock.Lock()
	defer m.lock.Unlock()

	one, ok := m.mp[k]
	if !ok {
		return
	}

	delete(m.mp, k)
	m.lst.Remove(one.ele)
}

func (m *ActiveMap[K, V]) Clear() {
	m.lock.Lock()
	defer m.lock.Unlock()

	m.mp = map[K]*activeNode[K, V]{}
	m.lst = list.New()
}

func (m *ActiveMap[K, V]) Len() int {
	m.lock.Lock()
	defer m.lock.Unlock()

	return len(m.mp)
}

func (m *ActiveMap[K, V]) Each(fn func(k K, v V)) {
	m.lock.Lock()
	defer m.lock.Unlock()

	for k, v := range m.mp {
		fn(k, v.v)
	}
}
