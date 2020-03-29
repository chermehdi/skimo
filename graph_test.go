package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func indexOf(arr []string, target string) int {
	for i, val := range arr {
		if val == target {
			return i
		}
	}
	return -1
}

func TestGraph_GetTopologicalOrder2(t *testing.T) {
	links := []Link{
		{"main.cpp", "include/strings.h"},                   // a -> b
		{"main.cpp", "include/lib/graphs.h"},                // a -> c
		{"include/lib/graphs.h", "include/lib/constants.h"}, // c -> b
	}
	graph := NewGraph(links, NewSet())
	order := graph.GetTopologicalOrder("include")

	assert.Equal(t, 3, len(order))
	assert.True(t, indexOf(order, "include/lib/graphs.h") > indexOf(order, "include/lib/constants.h"))
}

func TestGraph_GetTopologicalOrder(t *testing.T) {
	links := []Link{
		{"a.h", "b.h"}, // a -> b
		{"a.h", "c.h"}, // a -> c
		{"c.h", "b.h"}, // c -> b
		{"b.h", "f.h"}, // b -> f
	}

	graph := NewGraph(links, NewSet())
	order := graph.GetTopologicalOrder("include")

	fmt.Println(order)
	assert.Equal(t, 4, len(order))
	assert.Equal(t, "f.h", order[0])
	assert.Equal(t, "b.h", order[1])
	assert.Equal(t, "c.h", order[2])
	assert.Equal(t, "a.h", order[3])
}
