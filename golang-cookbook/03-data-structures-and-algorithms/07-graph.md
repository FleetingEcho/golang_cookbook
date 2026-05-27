# 图 — 邻接表/矩阵 + DFS/BFS

## 1. 图表示

```go
// 邻接表（最常用）
type Graph struct {
    n    int
    adj  [][]int // 无权重
}

func NewGraph(n int) *Graph {
    return &Graph{n: n, adj: make([][]int, n)}
}

func (g *Graph) AddEdge(u, v int) {
    g.adj[u] = append(g.adj[u], v)
}

func (g *Graph) AddUndirected(u, v int) {
    g.adj[u] = append(g.adj[u], v)
    g.adj[v] = append(g.adj[v], u)
}

// 带权邻接表
type Edge struct {
    To     int
    Weight int
}

type WeightedGraph struct {
    n    int
    adj  [][]Edge
}

func NewWeightedGraph(n int) *WeightedGraph {
    return &WeightedGraph{n: n, adj: make([][]Edge, n)}
}

func (g *WeightedGraph) AddEdge(u, v, w int) {
    g.adj[u] = append(g.adj[u], Edge{To: v, Weight: w})
}

// 邻接矩阵
type MatrixGraph struct {
    n   int
    mat [][]int // 0 = 无边, >0 = 边权
}

func NewMatrixGraph(n int) *MatrixGraph {
    mat := make([][]int, n)
    for i := range mat {
        mat[i] = make([]int, n)
    }
    return &MatrixGraph{n: n, mat: mat}
}
```

## 2. DFS

```go
func (g *Graph) DFS(start int) []int {
    visited := make([]bool, g.n)
    var result []int

    var dfs func(u int)
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

// 全连通分量遍历
func (g *Graph) DFSAll() [][]int {
    visited := make([]bool, g.n)
    var result [][]int

    for i := 0; i < g.n; i++ {
        if !visited[i] {
            var comp []int
            var dfs func(u int)
            dfs = func(u int) {
                visited[u] = true
                comp = append(comp, u)
                for _, v := range g.adj[u] {
                    if !visited[v] {
                        dfs(v)
                    }
                }
            }
            dfs(i)
            result = append(result, comp)
        }
    }
    return result
}
```

## 3. BFS

```go
func (g *Graph) BFS(start int) []int {
    visited := make([]bool, g.n)
    queue := []int{start}
    visited[start] = true
    var result []int

    for len(queue) > 0 {
        u := queue[0]; queue = queue[1:]
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

// BFS 找最短路径（无权）
func (g *Graph) ShortestPath(start, end int) int {
    if start == end { return 0 }
    dist := make([]int, g.n)
    for i := range dist { dist[i] = -1 }
    queue := []int{start}
    dist[start] = 0

    for len(queue) > 0 {
        u := queue[0]; queue = queue[1:]
        for _, v := range g.adj[u] {
            if dist[v] == -1 {
                dist[v] = dist[u] + 1
                if v == end { return dist[v] }
                queue = append(queue, v)
            }
        }
    }
    return -1 // 不可达
}
```

## 4. 泛型（字符串键）

```go
type GraphGeneric[T comparable] struct {
    adj map[T][]T
}

func NewGraphGeneric[T comparable]() *GraphGeneric[T] {
    return &GraphGeneric[T]{adj: make(map[T][]T)}
}

func (g *GraphGeneric[T]) AddEdge(u, v T) {
    g.adj[u] = append(g.adj[u], v)
}

func (g *GraphGeneric[T]) BFS(start T) []T {
    visited := make(map[T]bool)
    queue := []T{start}
    visited[start] = true
    var result []T
    for len(queue) > 0 {
        u := queue[0]; queue = queue[1:]
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
```

## 完整对照表

| 操作 | TS | Go |
|------|-----|-----|
| 邻接表 | `Map<number, number[]>` | `[][]int` |
| DFS 递归 | `function dfs(u)` | `func(u int)` 闭包 |
| BFS | `queue.push/shift` | 用 slice 模拟队列 |
| 路径 | `dist` 数组 | `dist` 数组 |
| 泛型图 | `Map<T, T[]>` | `GraphGeneric[T comparable]` |
