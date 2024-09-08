package collection

type Pair[T, K any] struct {
	key T
	value K
}

func NewPair[T, K any](key T, value K) Pair[T, K] {
	return Pair[T, K]{
		key: key,
		value: value,
	}
}

func (p Pair[T, K]) Key() T {
	return p.key
}

func (p Pair[T, K]) Value() K {
	return p.value
}