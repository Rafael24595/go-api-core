package query

type Queries struct {
	Queries map[string][]Query `json:"queries"`
}

func NewQueries() *Queries {
	return &Queries{
		Queries: make(map[string][]Query),
	}
}

func (q *Queries) Add(key string, query Query) *Queries {
	if _, ok := q.Queries[key]; !ok {
		q.Queries[key] = make([]Query, 0)
	}

	q.Queries[key] = append(q.Queries[key], query)

	return q
}