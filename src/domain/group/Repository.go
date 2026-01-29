package group

type Repository interface {
	Find(id string) (*Group, bool)
	Insert(owner string, group *Group) *Group
	Delete(group *Group) *Group
}
