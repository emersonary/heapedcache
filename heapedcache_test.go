package util

import (
    "github.com/stretchr/testify/assert"
    "math"
    "strconv"
    "sync"
    "testing"
    "time"
)

type AccountTest struct {
    Id     int
    Name   string
    Phone  string
    Filler [200]rune
}

func NewAccountTest(id int) *AccountTest {

    return &AccountTest{
        Id:     id,
        Name:   "EMERSON " + strconv.Itoa(id),
        Phone:  "PHONE " + strconv.Itoa(id),
        Filler: [200]rune{},
    }

}

func TestCachedHeapLen(t *testing.T) {

    t.Log("validating TestCachedHeapLen")

    heapedCache := NewHeapedCache[AccountTest](10)

    for i := range 9 {

        heapedCache.Push(i, NewAccountTest(i))

    }

    assert.Equal(t, 9, heapedCache.Len())

}

func TestCachedHeapLenOverFlow(t *testing.T) {

    t.Log("validating TestCachedHeapLenOverFlow")

    heapedCache := NewHeapedCache[AccountTest](10)

    for i := range 20 {

        heapedCache.Push(i, NewAccountTest(i))

    }

    assert.Equal(t, 10, heapedCache.Len())

}

func TestCachedHeapGet(t *testing.T) {

    t.Log("validating TestCachedHeapGet")

    heapedCache := NewHeapedCache[AccountTest](10)

    for i := range 9 {

        heapedCache.Push(i, NewAccountTest(i))

    }

    for i := range 9 {

        assert.Equal(t, i, heapedCache.Get(i).Id)
        assert.Equal(t, "EMERSON "+strconv.Itoa(i), heapedCache.Get(i).Name)
        assert.Equal(t, "PHONE "+strconv.Itoa(i), heapedCache.Get(i).Phone)

    }

}

func TestCachedHeap1Million(t *testing.T) {

    t.Log("validating TestCachedHeap1Million")

    heapedCache := NewHeapedCache[AccountTest](1000000)

    now1 := time.Now()

    for i := range 1000000 {

        heapedCache.Push(i, NewAccountTest(i))

    }

    duration := time.Since(now1)

    t.Log("time to create 1 million records: " + strconv.Itoa(int(duration.Milliseconds())) + " milliseconds (" + strconv.Itoa(int(duration.Nanoseconds())/1000000) + " nanoseconds per item)")

    now2 := time.Now()

    for i := range 1000000 {

        if i != heapedCache.Get(i).Id {

            t.Fatalf("test failed, expected %d, got %d", i, heapedCache.Get(i).Id)

        }

    }

    duration = time.Since(now2)

    t.Log("time to read 1 million records: " + strconv.Itoa(int(duration.Milliseconds())) + " milliseconds (" + strconv.Itoa(int(duration.Nanoseconds())/1000000) + " nanoseconds per item)")

}

func TestCachedHeap1MillionWithOverFlow(t *testing.T) {

    t.Log("validating TestCachedHeap1Million")

    heapedCache := NewHeapedCache[AccountTest](10000)

    now1 := time.Now()

    for i := range 1000000 {

        heapedCache.Push(i, NewAccountTest(i))

    }

    duration := time.Since(now1)

    t.Log("time to create 1 million records: " + strconv.Itoa(int(duration.Milliseconds())) + " milliseconds (" + strconv.Itoa(int(duration.Nanoseconds())/1000000) + " nanoseconds per item)")

    now2 := time.Now()

    min, max, count := math.MaxInt, math.MinInt, 0

    for i := range 1000000 {

        item := heapedCache.Get(i)

        if item != nil {

            if item.Id > max {
                max = item.Id

            }

            if item.Id < min {
                min = item.Id

            }

            count++

        }

    }

    t.Log("max = " + strconv.Itoa(max) + " min = " + strconv.Itoa(min) + " count = " + strconv.Itoa(count))

    assert.Equal(t, 10000, count)

    duration = time.Since(now2)

    t.Log("time to read 1 million records: " + strconv.Itoa(int(duration.Milliseconds())) + " milliseconds (" + strconv.Itoa(int(duration.Nanoseconds())/1000000) + " nanoseconds per item)")

}

func TestCachedHeap1MillionMultiThread(t *testing.T) {

    t.Log("validating TestCachedHeap1MillionMultiThread")

    heapedCache := NewHeapedCache[AccountTest](1000000)

    now1 := time.Now()

    var wg sync.WaitGroup

    wg.Add(10)

    for j := range 10 {

        go func() {

            for i := range 100000 {

                id := i*10 + j

                heapedCache.Push(id, NewAccountTest(i))

            }

            wg.Done()

        }()

    }

    wg.Wait()

    duration := time.Since(now1)

    t.Log("time to create 1 million records: " + strconv.Itoa(int(duration.Milliseconds())) + " milliseconds (" + strconv.Itoa(int(duration.Nanoseconds())/1000000) + " nanoseconds per item)")

    assert.Equal(t, 1000000, heapedCache.Len())

    now2 := time.Now()

    wg.Add(10)

    for range 10 {

        go func() {

            for range 100000 {

                heapedCache.Pop()

            }

            wg.Done()

        }()

    }

    wg.Wait()

    duration = time.Since(now2)

    t.Log("time to pop 1 million records: " + strconv.Itoa(int(duration.Milliseconds())) + " milliseconds (" + strconv.Itoa(int(duration.Nanoseconds())/1000000) + " nanoseconds per item)")

    assert.Equal(t, 0, heapedCache.Len())

}

func TestCachedHeap1MillionGetOrPush(t *testing.T) {

    t.Log("validating TestCachedHeap1MillionGetOrPush")

    heapedCache := NewHeapedCache[AccountTest](1000000)

    now1 := time.Now()

    fn := func(id any) *AccountTest {

        i := id.(int)

        if i == 0 {
            return nil
        }

        return NewAccountTest(i)
    }

    for i := range 1000000 {

        heapedCache.GetOrPush(i, fn)

    }

    duration := time.Since(now1)

    t.Log("time to create 1 million records: " + strconv.Itoa(int(duration.Milliseconds())) + " milliseconds (" + strconv.Itoa(int(duration.Nanoseconds())/1000000) + " nanoseconds per item)")

    assert.Equal(t, 999999, heapedCache.Len())

    now2 := time.Now()

    for i := range 1000000 {

        heapedCache.GetOrPush(i, fn)

    }

    duration = time.Since(now2)

    t.Log("time to read 1 million records: " + strconv.Itoa(int(duration.Milliseconds())) + " milliseconds (" + strconv.Itoa(int(duration.Nanoseconds())/1000000) + " nanoseconds per item)")

    assert.Equal(t, 999999, heapedCache.Len())

}

func TestDuration(t *testing.T) {

    t.Log("validating TestDuration")

    heapedCache := NewHeapedCache[AccountTest](1000000)

    now1 := time.Now()

    for i := range 1000000 {

        item := NewAccountTest(i)
        item.Id = item.Id
        item = nil

    }

    duration1 := time.Since(now1)

    t.Log("time to create 1 million isolated records: " + strconv.Itoa(int(duration1.Milliseconds())) + " milliseconds (" + strconv.Itoa(int(duration1.Nanoseconds())/1000000) + " nanoseconds per item)")

    now2 := time.Now()

    for i := range 1000000 {

        heapedCache.Push(i, NewAccountTest(i))

    }

    duration2 := time.Since(now2)
    duration3 := duration2 - duration1

    t.Log("total time to create 1 million records: " + strconv.Itoa(int(duration2.Milliseconds())) + " milliseconds (" + strconv.Itoa(int(duration2.Nanoseconds())/1000000) + " nanoseconds per item)")
    t.Log("time to store 1 million records: " + strconv.Itoa(int(duration3.Milliseconds())) + " milliseconds (" + strconv.Itoa(int(duration3.Nanoseconds())/1000000) + " nanoseconds per item)")
    t.Log("% of creation duration regarding the whole process: " + strconv.Itoa(int(float64(duration1.Milliseconds())/float64(duration2.Milliseconds())*100)) + "%")

}

// Test Cases to be implemented
// - Remove any positions
// - Remove first and last position
// - GetOrAdd
// - Pop Order Assert
// - Add same ID (check len)
// - compare slice len to map len
// - remove performance
// - sync lenght between map and slice after each operation
