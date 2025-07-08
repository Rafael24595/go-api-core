package dto

import "github.com/Rafael24595/go-api-core/src/domain"

type DtoNodeCollection struct {
	Order      int           `json:"order"`
	Collection DtoCollection `json:"collection"`
}

func ToCollectionNodes(dto []DtoNodeCollection) []domain.NodeReference {
	nodes := make([]domain.NodeReference, len(dto))

	for i := range dto {
		nodes[i] = domain.NodeReference{
			Order: dto[i].Order,
			Item:  dto[i].Collection.Id,
		}
	}

	return nodes
}

type DtoLiteNodeCollection struct {
	Order      int               `json:"order"`
	Collection DtoLiteCollection `json:"collection"`
}
