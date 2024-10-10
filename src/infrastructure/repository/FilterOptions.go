package repository

type FilterOptions[T any] struct {
	Predicate func(T) bool
	From      int
	To        int
	Sort      func(i, j T) bool
}
