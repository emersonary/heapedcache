package util

import (
    "container/heap"
    "sync"
    "time"
)

// struct to represent the cached item
type HeapedCacheItem[T any] struct {
    Id        any
    index     int
    Refreshed time.Time
    obj       *T
}

// this type wraps the array of HeapedCacheItem
// in order to define methods
type HeapedCacheItems[T any] []*HeapedCacheItem[T]

// type that represents the cache
type HeapedCache[T any] struct {
    mu         sync.Mutex
    maxRows    int
    mapItems   map[any]*HeapedCacheItem[T]
    sliceItems HeapedCacheItems[T]
}

// conctructor of the HeapedCache
// this cache is meant to have a fixed sized in memory.
// The higher the data volume, the lower the range of the cache
func NewHeapedCache[T any](maxRows int) *HeapedCache[T] {

    tolerance := 5

    return &HeapedCache[T]{
        maxRows:    maxRows,
        mapItems:   make(map[any]*HeapedCacheItem[T], maxRows+tolerance),
        sliceItems: make(HeapedCacheItems[T], 0, maxRows+tolerance),
    }

}

// removes the oldest cached item from the list (private)
func (t *HeapedCache[T]) pop() *T {

    item := heap.Pop(&t.sliceItems).(*HeapedCacheItem[T])
    delete(t.mapItems, item.Id)
    return item.obj

}

// removes the oldest cached item from the list (public)
func (t *HeapedCache[T]) Pop() *T {

    t.mu.Lock()
    defer t.mu.Unlock()

    return t.pop()

}

// returns the cached item of a given id
// returns nil if it does not exist
func (t *HeapedCache[T]) Get(id any) *T {

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
func (t *HeapedCache[T]) GetOrPush(id any, fn func(id any) *T) *T {

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

func (t *HeapedCache[T]) Len() int {

    t.mu.Lock()
    defer t.mu.Unlock()

    return len(t.mapItems)

}

// Adds new item to the cache when it does not exist (public)
// Updates the item when it does exist
func (t *HeapedCache[T]) push(id any, item *T) *T {

    if item == nil {
        return nil
    }

    findItem := t.mapItems[id]

    if findItem == nil {

        newItem := &HeapedCacheItem[T]{
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
func (t *HeapedCache[T]) Push(id any, item *T) *T {

    t.mu.Lock()
    defer t.mu.Unlock()

    return t.push(id, item)

}

// Remove items from the list (cache invalidation)
func (t *HeapedCache[T]) Remove(id any) bool {

    t.mu.Lock()
    defer t.mu.Unlock()

    findItem := t.mapItems[id]

    if findItem != nil {

        // removes item from the slice
        t.sliceItems.Swap(findItem.index, len(t.sliceItems)-1)
        t.sliceItems[len(t.sliceItems)-1] = nil // avoid memory leaks
        t.sliceItems = t.sliceItems[:len(t.sliceItems)-1]
        heap.Fix(&t.sliceItems, findItem.index)

        // remove item from the map
        delete(t.mapItems, id)

        return true

    }

    return false

}

// returns the size of the cache in lines
func (h *HeapedCacheItems[T]) Len() int {

    return len(*h)

}

// returns true if the cached item from the second index is smaller than the first one
func (h *HeapedCacheItems[T]) Less(i int, j int) bool {

    return (*h)[i].Refreshed.Compare((*h)[j].Refreshed) < 0

}

// swaps items of given indexes
func (h *HeapedCacheItems[T]) Swap(i int, j int) {

    if i != j {

        (*h)[i], (*h)[j] = (*h)[j], (*h)[i]

        (*h)[i].index = i
        (*h)[j].index = j

    }

}

// Adds item in the cache
func (h *HeapedCacheItems[T]) Push(x any) {

    *h = append(*h, x.(*HeapedCacheItem[T]))

}

// Removes last item (older) from the cache and returns it
func (h *HeapedCacheItems[T]) Pop() any {

    n := len(*h)
    item := (*h)[n-1]
    (*h)[n-1] = nil // avoid memory leak
    *h = (*h)[0 : n-1]

    return item

}
