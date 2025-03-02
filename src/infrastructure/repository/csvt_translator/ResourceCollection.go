package csvt_translator

import "github.com/Rafael24595/go-collections/collection"

type ResourceCollection struct {
	nexus collection.Dictionary[string, ResourceNexus]
}

func newCollection(nexus map[string]ResourceNexus) ResourceCollection {
	collection := collection.DictionaryFromMap(nexus)
	return ResourceCollection{
		nexus: *collection,
	}
}

func (r *ResourceCollection) root() (*ResourceNexus, bool) {
	return r.nexus.FindOne(func(k string, rn ResourceNexus) bool {
		return rn.root
	})
}

func (r *ResourceCollection) Find(node *ResourceNode) (*ResourceGroup, bool) {
	value, exists := r.nexus.Get(node.key())
	if !exists {
		return nil, false
	}
	if node.index != -1 {
		return value.nodes.Get(node.index)
		/*if !exists {
			return nil, false
		}
		if position.category == "OBJ" {
			parse := position.group.(ResourceNode)
			return parse, true
		}*/
	}
	return nil, false
}