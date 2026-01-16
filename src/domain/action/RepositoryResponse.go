package action

type RepositoryResponse interface {
	Find(key string) (*Response, bool)
	FindMany(ids []string) []Response
	Insert(owner string, response *Response) *Response
	Delete(response *Response) *Response
	DeleteMany(responses ...Response) []Response
}
