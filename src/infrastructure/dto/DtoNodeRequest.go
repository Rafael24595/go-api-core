package dto

import "github.com/Rafael24595/go-api-core/src/domain"

type DtoNodeRequest struct {
	Order   int        `json:"order"`
	Request DtoRequest `json:"request"`
}

func ToRequestNodes(dto []DtoNodeRequest) []domain.NodeReference {
	nodes := make([]domain.NodeReference, len(dto))

	for i := range dto {
		nodes[i] = domain.NodeReference{
			Order: dto[i].Order,
			Item:  dto[i].Request.Id,
		}
	}

	return nodes
}

type DtoLiteNodeRequest struct {
	Order   int            `json:"order"`
	Request DtoLiteRequest `json:"request"`
}
