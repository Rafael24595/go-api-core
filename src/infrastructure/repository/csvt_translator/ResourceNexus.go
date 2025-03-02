package csvt_translator

import "github.com/Rafael24595/go-collections/collection"

type ResourceNexus struct {
	key   string
	root  bool
	nodes collection.Vector[ResourceGroup]
}

func newNexus(key string, root bool, nodes []ResourceGroup) ResourceNexus {
	return ResourceNexus{
		key:   key,
		root:  root,
		nodes: *collection.VectorFromList(nodes),
	}
}

func (r *ResourceNexus) get(position int) (*ResourceGroup, bool) {
	return r.nodes.Get(position)
}
