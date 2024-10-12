package repository

type FilterMapOptions[T comparable, K any] struct {
	Predicate func(T, K) bool
	Mapper    func(K) T
	Filter    *FilterOptions[K]
}
