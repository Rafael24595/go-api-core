package collection

import (
	"math"
	"sort"
)

type CollectionList [T any] struct {
	items []T
}

func FromList[T any](items []T) *CollectionList[T] {
	return &CollectionList[T]{
		items,
	}
}

func (c *CollectionList[T]) Append(items ...T) *CollectionList[T] {
    c.items = append(c.items, items...)
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
    return Map(c, predicate)
}

func (c *CollectionList[T]) Map(predicate func(T) any) *CollectionList[interface{}] {
    return Map(c, predicate)
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

func Map[T, K any](c *CollectionList[T], predicate func(T) K) *CollectionList[K] {
    var mapped []K
    for _, item := range c.items {
		mapped = append(mapped, predicate(item))
    }
    return &CollectionList[K]{
		items: mapped,
	}
}

func (collection CollectionList[T]) Collect() []T {
	return collection.items
}
