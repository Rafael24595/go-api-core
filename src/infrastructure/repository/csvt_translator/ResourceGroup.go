package csvt_translator

import (
	"strconv"

	"github.com/Rafael24595/go-api-core/src/commons/collection"
)

type ResourceGroup struct {
	category GroupCategory
	headers  collection.CollectionList[string]
	group    any
}

func newGroup[T any](category GroupCategory, headers []string, group T) ResourceGroup {
	return ResourceGroup{
		category: category,
		headers: *collection.FromList(headers),
		group: group,
	}
}

func (r *ResourceGroup) findField(key string) (*ResourceNode, bool) {
	switch v := r.group.(type) {
    case []ResourceNode:
        index, exists := r.headers.IndexOf(func(s string) bool {
			return s == key
		})
		if !exists || index > len(v) {
			return nil, false
		}
		return &v[index], true
    default:
		return nil, false
    }
}

func (r *ResourceGroup) findFields() []collection.Pair[string, ResourceNode] {
	pairs := []collection.Pair[string, ResourceNode]{}
	switch v := r.group.(type) {
    case map[string]ResourceNode:
		for k, v := range v {
			pairs = append(pairs, collection.NewPair(k, v))
		}
	case []ResourceNode:
		for i, v := range v {
			pairs = append(pairs, collection.NewPair(strconv.Itoa(i), v))
		}
    }
	return pairs
}

func (r *ResourceGroup) findValue() (*ResourceNode, bool) {
	switch v := r.group.(type) {
    case ResourceNode:
        return &v, true
    default:
		return nil, false
    }
}
