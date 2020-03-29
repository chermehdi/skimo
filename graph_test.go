package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGraph_GetTopologicalOrder(t *testing.T) {
	links := []Link{
		{"a.h", "b.h"}, // a -> b
		{"a.h", "c.h"}, // a -> c
		{"c.h", "b.h"}, // c -> b
		{"b.h", "f.h"}, // b -> f
	}

	graph := NewGraph(links, NewSet())
	order := graph.GetTopologicalOrder("include")

	assert.Equal(t, 4, len(order))
	assert.Equal(t, "f.h", order[0])
	assert.Equal(t, "b.h", order[1])
	assert.Equal(t, "c.h", order[2])
	assert.Equal(t, "a.h", order[3])
}
