package dto

import "github.com/Rafael24595/go-api-core/src/domain"

type DtoCollection struct {
	Id        string     `json:"_id"`
	Name      string     `json:"name"`
	Timestamp int64      `json:"timestamp"`
	Context   DtoContext `json:"context"`
	Nodes     []DtoNode  `json:"nodes"`
	Owner     string     `json:"owner"`
	Modified  int64      `json:"modified"`
}

func ToCollection(dto *DtoCollection) *domain.Collection {
	return &domain.Collection{
		Id: dto.Id,
		Name: dto.Name,
		Timestamp: dto.Timestamp,
		Context: dto.Context.Id,
		Nodes: ToNodes(dto.Nodes),
		Owner: dto.Owner,
		Modified: dto.Modified,
	}
}

func ToNodes(dto []DtoNode) []domain.NodeReference {
	nodes := make([]domain.NodeReference, len(dto))

	for i := range dto {
		nodes[i] = domain.NodeReference{
			Order: dto[i].Order,
			Request: dto[i].Request.Id,
		}
	}

	return nodes
}