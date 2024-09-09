package collection

type CollectionMap [T comparable, K any] struct {
	items map[T]K
}

func (c *CollectionMap[T, K]) Size() int {
	return len(c.items)
}

func FromMap[T comparable, K any](items map[T]K) *CollectionMap[T, K] {
	return &CollectionMap[T, K]{
		items,
	}
}

func (c *CollectionMap[T, K]) Find(key T) (*K, bool) {
	value, exists := c.items[key]
	return &value, exists
}

func (c *CollectionMap[T, K]) FindOnePredicate(predicate func (K) bool) (*K, bool) {
	items := c.FindPredicate(predicate)
	if len(items) == 0 {
		return nil, false
	}
	return &items[0], true
}

func (c *CollectionMap[T, K]) FindPredicate(predicate func (K) bool) []K {
	filter := []K{}
	for _, v := range c.items {
		if predicate(v) {
			filter = append(filter, v)
		}
	}
	return filter
}

func (c *CollectionMap[T, K]) Exists(key T) bool {
	_, exists := c.items[key]
	return exists
}

func (c *CollectionMap[T, K]) Merge(other map[T]K) *CollectionMap[T, K] {
    for key := range other {
        c.items[key] = other[key]
    }
	return c
}

func (collection CollectionMap[T, K]) Keys() []T {
	keys := make([]T, 0, len(collection.items))
    for key := range collection.items {
        keys = append(keys, key)
    }
	return keys
}

func (c *CollectionMap[T, K]) Values() []K {
	values := make([]K, 0, len(c.items))
    for key := range c.items {
        values = append(values, c.items[key])
    }
	return values
}

func (c *CollectionMap[T, K]) Pairs() []Pair[T, K] {
	pairs := make([]Pair[T, K], 0, len(c.items))
    for key := range c.items {
        pairs = append(pairs, Pair[T, K]{
			key: key,
			value: c.items[key],
		})
    }
	return pairs
}

func (collection CollectionMap[T, K]) Collect() map[T]K {
	return collection.items
}
