package collection

import (
	"fmt"
	"math"
	"sort"
	"strings"
)

type CollectionList [T any] struct {
	items []T
}

func FromList[T any](items []T) *CollectionList[T] {
	return &CollectionList[T]{
		items,
	}
}

func EmptyList[T any]() *CollectionList[T] {
	return FromList(make([]T, 0))
}

func (c *CollectionList[T]) Append(items ...T) *CollectionList[T] {
    c.items = append(c.items, items...)
	return c
}

func (c *CollectionList[T]) Merge(other CollectionList[T]) *CollectionList[T] {
    c.items = append(c.items, other.items...)
	return c
}

func (c *CollectionList[T]) Size() int {
	return len(c.items)
}

func (c *CollectionList[T]) Get(index int) (*T, bool) {
    if index >= 0 && index < len(c.items) {
        return &c.items[index], true
    }
    return nil, false
}

func (c *CollectionList[T]) First() (*T, bool) {
    return c.Get(0)
}

func (c *CollectionList[T]) Last() (*T, bool) {
    return c.Get(c.Size() - 1)
}

func (c *CollectionList[T]) Sort(less func(i, j T) bool) *CollectionList[T] {
	sort.Slice(c.items, func(i, j int) bool {
        return less(c.items[i], c.items[j])
    })
	return c
}

func (c *CollectionList[T]) IndexOf(predicate func(T) bool) (int, bool) {
	for i, item := range c.items {
        if predicate(item) {
            return i, true
        }
    }
	return -1, false
}

func (c *CollectionList[T]) Find(predicate func(T) bool) (*T, bool) {
	for _, item := range c.items {
        if predicate(item) {
            return &item, true
        }
    }
	return nil, false
}

func (c *CollectionList[T]) FindPredicate(predicate func (T) bool) []T {
	filter := []T{}
	for _, v := range c.items {
		if predicate(v) {
			filter = append(filter, v)
		}
	}
	return filter
}

func (c *CollectionList[T]) Filter(predicate func(T) bool) *CollectionList[T] {
    var filtered []T
    for _, item := range c.items {
        if predicate(item) {
            filtered = append(filtered, item)
        }
    }
	c.items = filtered
    return c
}

func (c *CollectionList[T]) MapSelf(predicate func(T) T) *CollectionList[T] {
    return MapList(c, predicate)
}

func (c *CollectionList[T]) ForEach(predicate func(int, T)) *CollectionList[T] {
    for i, v := range c.items {
        predicate(i, v)
    }
    return c
}

func (c *CollectionList[T]) Map(predicate func(T) any) *CollectionList[interface{}] {
    return MapList(c, predicate)
}

func (c *CollectionList[T]) Pages(size int) int {
    len := float64(len(c.items))
    fSize := float64(size)
	return int(math.Ceil(len / fSize))
}

func (c *CollectionList[T]) Page(page, size int) *CollectionList[T] {
    if page == 0 {
        page = 1
    }
    start := (page - 1) * size;
    end := page * size;
	return c.Slice(start, end)
}

func (c *CollectionList[T]) Slice(start, end int) *CollectionList[T] {
    if start < 0 {
        start = 0
    }
    if start > len(c.items) -1 {
        start = len(c.items)
    }
    if end > len(c.items) -1 {
        end = len(c.items)
    }
    c.items = c.items[start:end]
	return c
}

func (c *CollectionList[T]) Join(separator string) string {
    if items, ok := interface{}(c.items).([]string); ok {
        return strings.Join(items, separator)
    }
    return MapList(c, func(i T) string {
        return fmt.Sprintf("%v", i)
    }).Join(separator)
}

func (c *CollectionList[T]) JoinBy(indexer func(T) string, predicate func(i, j T) T) *CollectionList[T] {
    dict := map[string]T{}
    for _, item := range c.items {
        key := indexer(item)
        aux := item
        if found, ok := dict[key]; ok {
            aux = predicate(found, item)
        }
        dict[key] = aux
    }

    c.items = make([]T, 0)
    for _, v := range dict {
        c.items = append(c.items, v)
    }

    return c
}

func (c *CollectionList[T]) Clean() *CollectionList[T] {
	c.items = make([]T, 0)
	return c
}

func (c *CollectionList[T]) Clone() *CollectionList[T] {
	return FromList(c.items)
}

func (c CollectionList[T]) Collect() []T {
	return c.items
}

func MapListFrom[T, K any](items []T, predicate func(T) K) *CollectionList[K] {
    return MapList(FromList(items), predicate)
}

func MapList[T, K any](c *CollectionList[T], predicate func(T) K) *CollectionList[K] {
    var mapped []K
    for _, item := range c.items {
		mapped = append(mapped, predicate(item))
    }
    return &CollectionList[K]{
		items: mapped,
	}
}

func MapperList[T comparable, K any](coll CollectionList[K], mapper func(K) T) *CollectionMap[T, K] {
    return Mapper(coll.items, mapper)
}

func Mapper[T comparable, K any](coll []K, mapper func(K) T) *CollectionMap[T, K] {
    mapp := EmptyMap[T, K]()
    for _, v := range coll {
        mapp.Put(mapper(v), v)
    }
    return mapp
}