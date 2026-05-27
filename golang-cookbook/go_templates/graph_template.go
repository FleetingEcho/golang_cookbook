// 图模板 — 邻接表 + DFS + BFS + Dijkstra
package main

import (
	"container/heap"
	"fmt"
	"math"
)

// ---------- 无权图 ----------
type Graph struct {
	n   int
	adj [][]int
}

func NewGraph(n int) *Graph {
	return &Graph{n: n, adj: make([][]int, n)}
}

func (g *Graph) AddEdge(u, v int) {
	g.adj[u] = append(g.adj[u], v)
}

func (g *Graph) AddUndirected(u, v int) {
	g.AddEdge(u, v)
	g.AddEdge(v, u)
}

func (g *Graph) DFS(start int) []int {
	visited := make([]bool, g.n)
	var result []int
	var dfs func(int)
	dfs = func(u int) {
		visited[u] = true
		result = append(result, u)
		for _, v := range g.adj[u] {
			if !visited[v] {
				dfs(v)
			}
		}
	}
	dfs(start)
	return result
}

func (g *Graph) BFS(start int) []int {
	visited := make([]bool, g.n)
	queue := []int{start}
	visited[start] = true
	var result []int

	for len(queue) > 0 {
		u := queue[0]
		queue = queue[1:]
		result = append(result, u)
		for _, v := range g.adj[u] {
			if !visited[v] {
				visited[v] = true
				queue = append(queue, v)
			}
		}
	}
	return result
}

// ---------- 带权图 + Dijkstra ----------
type Edge struct {
	To     int
	Weight int
}

type WeightedGraph struct {
	n   int
	adj [][]Edge
}

func NewWeightedGraph(n int) *WeightedGraph {
	return &WeightedGraph{n: n, adj: make([][]Edge, n)}
}

func (g *WeightedGraph) AddEdge(u, v, w int) {
	g.adj[u] = append(g.adj[u], Edge{To: v, Weight: w})
}

// Dijkstra 最短路径
type Item struct {
	node int
	dist int
}
type PriorityQueue []Item

func (pq PriorityQueue) Len() int            { return len(pq) }
func (pq PriorityQueue) Less(i, j int) bool  { return pq[i].dist < pq[j].dist }
func (pq PriorityQueue) Swap(i, j int)       { pq[i], pq[j] = pq[j], pq[i] }
func (pq *PriorityQueue) Push(v any)         { *pq = append(*pq, v.(Item)) }
func (pq *PriorityQueue) Pop() any {
	old := *pq
	n := len(old)
	v := old[n-1]
	*pq = old[:n-1]
	return v
}

func (g *WeightedGraph) Dijkstra(start int) []int {
	dist := make([]int, g.n)
	for i := range dist {
		dist[i] = math.MaxInt
	}
	dist[start] = 0

	pq := &PriorityQueue{}
	heap.Init(pq)
	heap.Push(pq, Item{start, 0})

	for pq.Len() > 0 {
		item := heap.Pop(pq).(Item)
		u, d := item.node, item.dist
		if d > dist[u] {
			continue
		}
		for _, e := range g.adj[u] {
			nd := d + e.Weight
			if nd < dist[e.To] {
				dist[e.To] = nd
				heap.Push(pq, Item{e.To, nd})
			}
		}
	}
	return dist
}

func main() {
	// 示例：无权图 BFS
	g := NewGraph(5)
	g.AddUndirected(0, 1)
	g.AddUndirected(0, 2)
	g.AddUndirected(1, 3)
	g.AddUndirected(2, 4)

	fmt.Println("DFS:", g.DFS(0))
	fmt.Println("BFS:", g.BFS(0))

	// 示例：Dijkstra
	wg := NewWeightedGraph(5)
	wg.AddEdge(0, 1, 4)
	wg.AddEdge(0, 2, 1)
	wg.AddEdge(2, 1, 2)
	wg.AddEdge(1, 3, 1)
	wg.AddEdge(2, 3, 5)

	dist := wg.Dijkstra(0)
	fmt.Println("Dijkstra from 0:", dist)
}
