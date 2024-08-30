package collection

import "sort"

type CollectionMap [T comparable, K any] struct {
	items map[T]K
}

func FromMap[T comparable, K any](items map[T]K) *CollectionMap[T, K] {
	return &CollectionMap[T, K]{
		items,
	}
}

func (collection *CollectionMap[T, K]) Sort(less func(a, b T) bool) *CollectionMap[T, K] {
	keys := collection.Keys()
	sort.Slice(keys, func(i, j int) bool {
		return less(keys[i], keys[j])
	})
	sorted := map[T]K{}
    for _, key := range keys {
		sorted[key] = collection.items[key]
    }
	collection.items = sorted
	return collection
}

func (collection CollectionMap[T, K])  Keys() []T {
	keys := make([]T, 0, len(collection.items))
    for key := range collection.items {
        keys = append(keys, key)
    }
	return keys
}

func (collection CollectionMap[T, K]) Collect() map[T]K {
	return collection.items
}
