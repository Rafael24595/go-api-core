package test_collection

import (
	"testing"

	"github.com/Rafael24595/go-api-core/src/domain"
)

func TestSortRequests(t *testing.T) {
	collection := domain.NewFreeCollection("anonymous")
	collection.Nodes = []domain.NodeReference{
		{
			Order: 3,
			Item: "3",
		},
		{
			Order: 1,
			Item: "1",
		},
		{
			Order: 4,
			Item: "4",
		},
		{
			Order: 2,
			Item: "2",
		},
	}

	collection.SortRequests()

	for i, node := range collection.Nodes {
		expected := i + 1
		if node.Order != expected {
			t.Errorf("Found variable %d but %d expected", node.Order, expected)
		}
	}
}

func TestExistsRequest(t *testing.T) {
	collection := domain.NewFreeCollection("anonymous")
	collection.Nodes = []domain.NodeReference{
		{
			Order: 3,
			Item: "3",
		},
		{
			Order: 1,
			Item: "1",
		},
		{
			Order: 4,
			Item: "4",
		},
		{
			Order: 2,
			Item: "2",
		},
	}

	cursor := "2"
	if !collection.ExistsRequest(cursor) {
		t.Errorf("Request %s not found.", cursor)
	}

	cursor = "5"
	if collection.ExistsRequest(cursor) {
		t.Errorf("Request %s found.", cursor)
	}
}

func TestTakeRequest(t *testing.T) {
	collection := domain.NewFreeCollection("anonymous")
	collection.Nodes = []domain.NodeReference{
		{
			Order: 3,
			Item: "3",
		},
		{
			Order: 1,
			Item: "1",
		},
		{
			Order: 4,
			Item: "4",
		},
		{
			Order: 2,
			Item: "2",
		},
	}

	cursor := "2"
	_, exists := collection.TakeRequest(cursor)

	if !exists {
		t.Errorf("Request %s not found", cursor)
	}


	if len(collection.Nodes) > 3 || collection.ExistsRequest(cursor) {
		t.Errorf("Request %s found after take it.", cursor)
	}
}