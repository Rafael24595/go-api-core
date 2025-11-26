package action

import "github.com/Rafael24595/go-api-core/src/domain"

type NodeRequest struct {
	Order   int     `json:"order"`
	Request Request `json:"request"`
}

func ToRequestNodes(nodes []NodeRequest) []domain.NodeReference {
	references := make([]domain.NodeReference, len(nodes))

	for i := range nodes {
		references[i] = domain.NodeReference{
			Order: nodes[i].Order,
			Item:  nodes[i].Request.Id,
		}
	}

	return references
}

type NodeRequestLite struct {
	Order   int         `json:"order"`
	Request RequestLite `json:"request"`
}

func ToNodeRequestLite(requests []NodeRequest) []NodeRequestLite {
	nodes := make([]NodeRequestLite, len(requests))

	for i := range requests {
		nodes[i] = NodeRequestLite{
			Order:   requests[i].Order,
			Request: *ToLiteRequest(&requests[i].Request),
		}
	}

	return nodes
}