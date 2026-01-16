package context

type Repository interface {
	Find(id string) (*Context, bool)
	Insert(owner string, collection string, context *Context) *Context
	Update(owner string, context *Context) (*Context, bool)
	Delete(context *Context) *Context
}
