package query

type Queries struct {
	Queries map[string][]Query `json:"queries"`
}

func NewQueries() *Queries {
	return &Queries{
		Queries: make(map[string][]Query),
	}
}

func (q *Queries) Add(query Query) *Queries {
	if _, ok := q.Queries[query.Key]; !ok {
		q.Queries[query.Key] = make([]Query, 0)
	}

	q.Queries[query.Key] = append(q.Queries[query.Key], query)

	return q
}