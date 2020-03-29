package main

import (
	"path"
)

type Set struct {
	members map[string]bool
}

func NewSet() Set {
	return Set{
		members: make(map[string]bool),
	}
}

func (set Set) Has(path string) bool {
	_, exist := set.members[path]
	return exist
}

func (set Set) Add(path string) {
	set.members[path] = true
}

func (set Set) Remove(path string) {
	delete(set.members, path)
}

type Graph struct {
	seen   Set            // #includes that are already been processed
	degree map[string]int // The degree of a given path
	adj    map[string][]string
}

func (g Graph) GetTopologicalOrder(root string) []string {
	// TODO(chermehdi): what if there is a cycle?
	order := make([]string, 0)
	for k := range g.adj {
		if g.degree[k] == 0 && !g.seen.Has(k) {
			g.getTopologicalOrder(k, path.Join(root, k), &order)
		}
	}
	return order
}

func (g Graph) getTopologicalOrder(current string, currentPath string, order *[]string) {
	g.seen.Add(current)
	for _, v := range g.adj[current] {
		g.degree[v]--
		if g.degree[v] == 0 && !g.seen.Has(v) {
			g.getTopologicalOrder(v, path.Join(currentPath, v), order)
		}
	}
	*order = append(*order, currentPath)
}

func NewGraph(links []Link, seen Set) Graph {
	adj := make(map[string][]string)
	degree := make(map[string]int)

	for _, link := range links {
		adj[link.from] = make([]string, 0)
		adj[link.to] = make([]string, 0)
		degree[link.to] = 0
		degree[link.from] = 0
	}

	for _, link := range links {
		adj[link.from] = append(adj[link.from], link.to)
		degree[link.to]++
	}
	return Graph{
		seen:   seen,
		degree: degree,
		adj:    adj,
	}
}
