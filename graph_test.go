package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGraph_GetTopologicalOrder(t *testing.T) {
	links := []Link{
		{"a", "b"}, // a -> b
		{"a", "c"}, // a -> c
		{"c", "b"}, // c -> b
		{"b", "f"}, // b -> f
	}

	graph := NewGraph(links, NewSet())
	order := graph.GetTopologicalOrder("include")

	assert.Equal(t, 4, len(order))
	assert.Equal(t, "include/a/c/b/f", order[0])
	assert.Equal(t, "include/a/c/b", order[1])
	assert.Equal(t, "include/a/c", order[2])
	assert.Equal(t, "include/a", order[3])
}
