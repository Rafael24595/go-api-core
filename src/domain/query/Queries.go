package query

type Queries struct {
	Queries map[string]Query `json:"queries"`
}

func NewQueries() *Queries {
	return &Queries{
		Queries: make(map[string]Query),
	}
}

func (q *Queries) Add(query Query) *Queries {
	param, ok := q.Queries[query.Key]
	if !ok {
		q.Queries[query.Key] = query
		return q
	}

	param.Query = append(param.Query, query.Query...)
	
	if query.Active && !param.Active {
		param.Active = query.Active
	}

	return q
}