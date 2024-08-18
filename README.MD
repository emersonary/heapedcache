# HeapedCache

`HeapedCache` is a generic, extremely fast, thread-safe in-memory cache for Go that maintains a fixed size by automatically removing the least recently accessed items when the cache limit is reached. This package leverages a min-heap to efficiently manage cache eviction based on item age.

## Features

---
- **Fixed Size**: The cache has a maximum number of items (`maxRows`). When this limit is reached, the oldest item is automatically removed to make space for new entries. The cache is intended to be initialized at the start of the application, allocating the desired amount of memory upfront. There is no capacity resizing, which ensures maximum performance.
<br><br>
- **Time Complexity**: The cache structure comprises a hash map and a slice. All operations on the map have a time complexity of **O(1)**. The slice is managed using Go's `container/heap` library, with all its operations having a time complexity of **O(log n)**. Both the map and the slice store pointers to the cached objects, facilitating efficient access and management.
  <br><br>
- **Thread-Safe**: The cache is safe for concurrent use by multiple goroutines, thanks to internal mutex locks that manage concurrent read/write operations.
  <br><br>
- **Generic**: Implemented using Go's type parameters (`[T any]`), the cache can store any type of object, providing flexibility and type safety.
  <br><br>
- **Automatic Eviction**: When the cache reaches its maximum capacity, items with the oldest timestamps are automatically evicted to make room for new entries. Note that the items in the slice are not kept in a sorted order in memory to avoid additional overhead.
  <br><br>
- **Customizable Loading**: If an item is not found in the cache, a custom loading function can be executed to generate and add the item to the cache seamlessly.
  <br><br>
- **Efficient Searching**: The structure provides rapid access to items through unique IDs, leveraging the efficiency of the underlying hash map.


## Memory allocation

---
In a 64 bit application, it consumes 16 bytes per item. That means that if you store 1 million items, it allocates 16 MB, represented by a list of pointers to the desired structure, whose (unknown) allocated memory is not referenced on this estimation.  

## Known use cases (so far)

---
#### 1. It can be used as a cache for any structure.

#### 2. It can be used to create a stream buffer guaranteed to transmit older messages first.

## Speed

---
Tests were performed in the environment below:

- Processor: 13th Gen Intel(R) Core(TM) i7-1355U   1.70 GHz
- 16GB RAM
- OS: Windows

#### 1. Creation of 1 million items (200 bytes struct) on an 1 million capacity heaped cache: 
 
- average of 700 nanoseconds per item on creation, 90 nanoseconds on reading
- 60% of creation duration were related to the memory allocation of the item itself. 
- So storing duration constituted an average of 280 nanoseconds.

#### 2. Creation of 1 million items (200 bytes struct) on an 100 thousand capacity heaped cache:

- average of 650 nanoseconds per item on creation, 30 nanoseconds on reading
- 60% of creation duration were related to the memory allocation of the item itself.
- So storing duration constituted an average of 260 nanoseconds.

---
## Usage

### 1. Creating a New Cache

```go
import "path/to/util"

type Person struct {
    Id     int
    Name   string
    Phone  string
}

cache := util.NewHeapedCache[Person]( 1000000 )  // Cache with 1 million items
```

### 2. Adding Items to the Cache

```go
obj := &Person{/* ... */}
cache.Push(obj.Id, obj)
```

### 3. Retrieving Items from the Cache

```go
obj := cache.Get(itemId)
if obj != nil {
    // Use the cached object
}
```

### 4. Fetching or Adding Items with a Loading Function

If the item is not found in the cache, `GetOrPush` can be used to fetch or create it:

```go
obj := cache.GetOrPush(itemId, func(id any) *Person {
    // Logic to load/create the item if it's not in the cache
    return &Person{/* ... */}
})
```

### 5. Removing Items from the Cache (Invalidation)

```go
removed := cache.Remove(itemId)
if removed {
    // Item was successfully removed
}
```

### 6. Checking the Size of the Cache

```go
size := cache.Len()
```

## API Reference

---
### `NewHeapedCache[T any](maxRows int) *HeapedCache[T]`
Creates a new `HeapedCache` with a fixed maximum size.

### `Push(id any, item *T) *T`
Adds an item to the cache or updates it if it already exists. If the cache is full, the oldest item is evicted.

### `Get(id any) *T`
Retrieves an item from the cache by its ID. Returns `nil` if the item is not found.

### `GetOrPush(id any, fn func(id any) *T) *T`
Retrieves an item from the cache by its ID. If the item does not exist, the provided function `fn` is called to create it, and the new item is added to the cache.

### `Remove(id any) bool`
Removes an item from the cache by its ID. Returns `true` if the item was successfully removed.

### `Len() int`
Returns the number of items currently stored in the cache.

## Understanding Priority Queues

---

### https://www.youtube.com/watch?v=wptevk0bshY

## Contributing

---
Contributions are welcome! Please feel free to submit a pull request or open an issue to discuss any changes.
