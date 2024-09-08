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

func (collection *CollectionList[T]) Size() int {
	return len(collection.items)
}

func (collection *CollectionList[T]) Get(index int) (*T, bool) {
    if index >= 0 && index < len(collection.items) {
        return &collection.items[index], true
    }
    return nil, false
}

func (collection *CollectionList[T]) Sort(predicate func(i, j int) bool) *CollectionList[T] {
	sort.Slice(collection.items, predicate)
	return collection
}

func (collection *CollectionList[T]) IndexOf(predicate func(T) bool) (int, bool) {
	for i, item := range collection.items {
        if predicate(item) {
            return i, true
        }
    }
	return -1, false
}

func (collection *CollectionList[T]) Find(predicate func(T) bool) (*T, bool) {
	for _, item := range collection.items {
        if predicate(item) {
            return &item, true
        }
    }
	return nil, false
}

func (collection *CollectionList[T]) Filter(predicate func(T) bool) *CollectionList[T] {
    var filtered []T
    for _, item := range collection.items {
        if predicate(item) {
            filtered = append(filtered, item)
        }
    }
	collection.items = filtered
    return collection
}

func (collection *CollectionList[T]) MapSelf(predicate func(T) T) *CollectionList[T] {
    return Map(collection, predicate)
}

func (collection *CollectionList[T]) Map(predicate func(T) any) *CollectionList[interface{}] {
    return Map(collection, predicate)
}

func (collection *CollectionList[T]) Pages(size int) int {
    len := float64(len(collection.items))
    fSize := float64(size)
	return int(math.Ceil(len / fSize))
}

func (collection *CollectionList[T]) Page(page, size int) *CollectionList[T] {
    if page == 0 {
        page = 1
    }
    start := (page - 1) * size;
    end := page * size;
	return collection.Slice(start, end)
}

func (collection *CollectionList[T]) Slice(start, end int) *CollectionList[T] {
    if start < 0 {
        start = 0
    }
    if start > len(collection.items) -1 {
        start = len(collection.items)
    }
    if end > len(collection.items) -1 {
        end = len(collection.items)
    }
    collection.items = collection.items[start:end]
	return collection
}

func Map[T, K any](collection *CollectionList[T], predicate func(T) K) *CollectionList[K] {
    var mapped []K
    for _, item := range collection.items {
		mapped = append(mapped, predicate(item))
    }
    return &CollectionList[K]{
		items: mapped,
	}
}

func (collection CollectionList[T]) Collect() []T {
	return collection.items
}
