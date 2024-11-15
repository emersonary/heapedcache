package utils

import (
	"container/heap"
	"sync"
	"time"
)

// struct to represent the cached item
type HeapedCacheItem[TId any, TObj any] struct {
	Id        TId
	index     int
	Refreshed time.Time
	obj       *TObj
}

// this type wraps the array of HeapedCacheItem
// in order to define methods
type HeapedCacheItems[TId any, TObj any] []*HeapedCacheItem[TId, TObj]

// type that represents the cache
type HeapedCache[TId any, TObj any] struct {
	mu         sync.RWMutex
	maxRows    int
	mapItems   map[any]*HeapedCacheItem[TId, TObj]
	sliceItems HeapedCacheItems[TId, TObj]
}

// conctructor of the HeapedCache
// this cache is meant to have a fixed sized in memory.
// The higher the data volume, the lower the range of the cache
func NewHeapedCache[TId any, TObj any](maxRows int) *HeapedCache[TId, TObj] {

	return &HeapedCache[TId, TObj]{
		maxRows:    maxRows,
		mapItems:   make(map[any]*HeapedCacheItem[TId, TObj], maxRows+1),
		sliceItems: make(HeapedCacheItems[TId, TObj], 0, maxRows+1),
	}

}

// removes the oldest cached item from the list (private)
func (t *HeapedCache[Tid, TObj]) pop() *TObj {

	item := heap.Pop(&t.sliceItems).(*HeapedCacheItem[Tid, TObj])
	delete(t.mapItems, item.Id)
	return item.obj

}

// removes the oldest cached item from the list (public)
func (t *HeapedCache[TId, TObj]) Pop() *TObj {

	t.mu.Lock()
	defer t.mu.Unlock()

	return t.pop()

}

func (t *HeapedCache[TId, TObj]) PopWithRefreshed() (*TObj, time.Time) {

	t.mu.Lock()
	defer t.mu.Unlock()

	return t.popWithRefreshed()

}

func (t *HeapedCache[Tid, TObj]) popWithRefreshed() (*TObj, time.Time) {

	item := heap.Pop(&t.sliceItems).(*HeapedCacheItem[Tid, TObj])
	delete(t.mapItems, item.Id)
	return item.obj, item.Refreshed

}

// returns the cached item of a given id
// returns nil if it does not exist
func (t *HeapedCache[Tid, TObj]) Get(id any) *TObj {

	t.mu.Lock()
	defer t.mu.Unlock()

	item := t.mapItems[id]

	if item == nil {
		return nil
	}

	return item.obj

}

// returns the cached item of a given id
// if it does not exist, fn is executed and returned in the function
// while the new item is placed on the cache
func (t *HeapedCache[TId, TObj]) GetOrAdd(id TId, fn func(id TId) *TObj) *TObj {

	t.mu.Lock()
	defer t.mu.Unlock()

	findItem := t.mapItems[id]

	if findItem == nil {

		result := fn(id)

		if result == nil {
			return nil
		}

		return t.push(id, result)

	} else {

		return findItem.obj

	}

}

func (t *HeapedCache[TId, TObj]) Len() int {

	t.mu.Lock()
	defer t.mu.Unlock()

	return len(t.mapItems)

}

// Adds new item to the cache when it does not exist (public)
// Updates the item when it does exist
func (t *HeapedCache[TId, TObj]) push(id TId, item *TObj) *TObj {

	if item == nil {
		return nil
	}

	findItem := t.mapItems[id]

	if findItem == nil {

		newItem := &HeapedCacheItem[TId, TObj]{
			Id:        id,
			index:     len(t.sliceItems),
			Refreshed: time.Now(),
			obj:       item,
		}

		t.mapItems[id] = newItem

		heap.Push(&t.sliceItems, newItem)

		if len(t.sliceItems) > t.maxRows {
			t.pop()
		}

	} else {

		findItem.obj = item
		findItem.Refreshed = time.Now()
		heap.Fix(&t.sliceItems, findItem.index)

	}

	return item

}

// Adds new item to the cache when it does not exist (public)
// Updates the item when it does exist
func (t *HeapedCache[TId, TObj]) Push(id TId, item *TObj) *TObj {

	t.mu.Lock()
	defer t.mu.Unlock()

	return t.push(id, item)

}

// Remove items from the list (cache invalidation)
func (t *HeapedCache[TId, TObj]) Remove(id TId) bool {

	t.mu.Lock()
	defer t.mu.Unlock()

	findItem := t.mapItems[id]

	if findItem != nil {

		// removes item from the slice
		t.sliceItems.Swap(findItem.index, len(t.sliceItems)-1)
		t.sliceItems[len(t.sliceItems)-1] = nil // don't stop the GC from reclaiming the item eventually
		heap.Fix(&t.sliceItems, findItem.index)
		t.sliceItems = t.sliceItems[:len(t.sliceItems)-1]

		// remove item from the map
		delete(t.mapItems, id)

		return true

	}

	return false

}

// returns the size of the cache in lines
func (h *HeapedCacheItems[TId, TObj]) Len() int {

	return len(*h)

}

// returns true if the cached item from the second index is smaller than the first one
func (h *HeapedCacheItems[TId, TObj]) Less(i int, j int) bool {

	return (*h)[i].Refreshed.Compare((*h)[j].Refreshed) < 0

}

// swaps items of given indexes
func (h *HeapedCacheItems[TId, TObj]) Swap(i int, j int) {

	if i != j {

		(*h)[i], (*h)[j] = (*h)[j], (*h)[i]

		(*h)[i].index = i
		(*h)[j].index = j

	}

}

// Adds item in the cache
func (h *HeapedCacheItems[TId, TObj]) Push(x any) {

	*h = append(*h, x.(*HeapedCacheItem[TId, TObj]))

}

// Removes last item (older) from the cache and returns it
func (h *HeapedCacheItems[TId, TObj]) Pop() any {

	n := len(*h)
	item := (*h)[n-1]
	(*h)[n-1] = nil // don't stop the GC from reclaiming the item eventually
	item.index = -1 // for safety
	*h = (*h)[0 : n-1]

	return item

}
