package dto

import (
	"github.com/Rafael24595/go-api-core/src/domain"
	"github.com/Rafael24595/go-api-core/src/domain/action"
)

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

func FromNodeRequest(nodes []action.NodeRequest) []DtoNodeRequest {
	dtos := make([]DtoNodeRequest, len(nodes))

	for i := range nodes {
		dtos[i] = DtoNodeRequest{
			Order: nodes[i].Order,
			Request: *FromRequest(&nodes[i].Request),
		}
	}

	return dtos
}
