package collection

type CollectionMap [T comparable, K any] struct {
	items map[T]K
}

func FromMap[T comparable, K any](items map[T]K) *CollectionMap[T, K] {
	return &CollectionMap[T, K]{
		items,
	}
}

func (collection *CollectionMap[T, K]) Find(key T) (*K, bool) {
	value, exists := collection.items[key]
	return &value, exists
}

func (collection *CollectionMap[T, K]) Exists(key T) bool {
	_, exists := collection.items[key]
	return exists
}

func (collection *CollectionMap[T, K]) Merge(other map[T]K) *CollectionMap[T, K] {
    for key := range other {
        collection.items[key] = other[key]
    }
	return collection
}

func (collection CollectionMap[T, K]) Keys() []T {
	keys := make([]T, 0, len(collection.items))
    for key := range collection.items {
        keys = append(keys, key)
    }
	return keys
}

func (collection *CollectionMap[T, K]) Values() []K {
	values := make([]K, 0, len(collection.items))
    for key := range collection.items {
        values = append(values, collection.items[key])
    }
	return values
}

func (collection *CollectionMap[T, K]) Pairs() []Pair[T, K] {
	pairs := make([]Pair[T, K], 0, len(collection.items))
    for key := range collection.items {
        pairs = append(pairs, Pair[T, K]{
			key: key,
			Value: collection.items[key],
		})
    }
	return pairs
}

func (collection CollectionMap[T, K]) Collect() map[T]K {
	return collection.items
}
