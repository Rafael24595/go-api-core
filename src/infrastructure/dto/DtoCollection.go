package dto

import "github.com/Rafael24595/go-api-core/src/domain"

type DtoCollection struct {
	Id        string                  `json:"_id"`
	Name      string                  `json:"name"`
	Timestamp int64                   `json:"timestamp"`
	Context   DtoContext              `json:"context"`
	Nodes     []DtoNodeRequest        `json:"nodes"`
	Owner     string                  `json:"owner"`
	Modified  int64                   `json:"modified"`
	Status    domain.StatusCollection `json:"status"`
}

func FromCollection(collection *domain.Collection, ctx *DtoContext, nodes []DtoNodeRequest) *DtoCollection{
	return &DtoCollection{
		Id:        collection.Id,
		Name:      collection.Name,
		Timestamp: collection.Timestamp,
		Context:   *ctx,
		Nodes:     nodes,
		Owner:     collection.Owner,
		Modified:  collection.Modified,
	}
}

func ToCollection(dto *DtoCollection) *domain.Collection {
	return &domain.Collection{
		Id:        dto.Id,
		Name:      dto.Name,
		Timestamp: dto.Timestamp,
		Context:   dto.Context.Id,
		Nodes:     ToRequestNodes(dto.Nodes),
		Owner:     dto.Owner,
		Modified:  dto.Modified,
		Status:    dto.Status,
	}
}
