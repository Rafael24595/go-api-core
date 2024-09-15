package translator

import (
	"github.com/Rafael24595/go-api-core/src/commons/collection"
)

type ResourceNexus struct {
	key     string
	root    bool
	nodes   collection.CollectionList[ResourceGroup]
}

func newNexus(key string, root bool, nodes []ResourceGroup) ResourceNexus {
	return ResourceNexus{
		key: key,
		root: root,
		nodes: *collection.FromList(nodes),
	}
}

func (r *ResourceNexus) get(position int) (*ResourceGroup, bool) {
	return r.nodes.Get(position)
}
