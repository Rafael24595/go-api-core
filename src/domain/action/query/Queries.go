package query

type Queries struct {
	Queries map[string][]Query `json:"queries"`
}

func NewQueries() *Queries {
	return &Queries{
		Queries: make(map[string][]Query),
	}
}

func (q *Queries) Add(key, value string) *Queries {
	return q.AddStatus(key, value, true)
}

func (q *Queries) AddStatus(key, value string, status bool) *Queries {
	return q.AddQuery(key, Query{
		Order:  int64(len(q.Queries)),
		Status: status,
		Value:  value,
	})
}

func (q *Queries) AddQuery(key string, query Query) *Queries {
	if _, ok := q.Queries[key]; !ok {
		q.Queries[key] = make([]Query, 0)
	}

	q.Queries[key] = append(q.Queries[key], query)

	return q
}

func (q *Queries) SizeOf(key string) int {
	if queries, ok := q.Queries[key]; ok {
		return len(queries)
	}
	return 0
}
