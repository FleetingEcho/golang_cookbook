# 拓扑排序与最短路径

## 1. 拓扑排序（Kahn 算法 / BFS）

```go
// 入度表法（Kahn）
func TopologicalSort(n int, edges [][]int) []int {
    adj := make([][]int, n)
    inDegree := make([]int, n)

    for _, e := range edges {
        u, v := e[0], e[1]
        adj[u] = append(adj[u], v)
        inDegree[v]++
    }

    queue := []int{}
    for i := 0; i < n; i++ {
        if inDegree[i] == 0 {
            queue = append(queue, i)
        }
    }

    var result []int
    for len(queue) > 0 {
        u := queue[0]; queue = queue[1:]
        result = append(result, u)
        for _, v := range adj[u] {
            inDegree[v]--
            if inDegree[v] == 0 {
                queue = append(queue, v)
            }
        }
    }

    if len(result) != n { return nil } // 有环
    return result
}

// DFS 后序遍历（逆后序即为拓扑序）
func TopologicalSortDFS(n int, edges [][]int) []int {
    adj := make([][]int, n)
    for _, e := range edges {
        adj[e[0]] = append(adj[e[0]], e[1])
    }

    visited := make([]int8, n) // 0=白 1=灰 2=黑
    var result []int
    var dfs func(u int) bool
    dfs = func(u int) bool {
        if visited[u] == 1 { return false } // 环
        if visited[u] == 2 { return true }
        visited[u] = 1
        for _, v := range adj[u] {
            if !dfs(v) { return false }
        }
        visited[u] = 2
        result = append(result, u) // 后序加入
        return true
    }

    for i := 0; i < n; i++ {
        if visited[i] == 0 {
            if !dfs(i) { return nil }
        }
    }

    // 反转得到拓扑序
    for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
        result[i], result[j] = result[j], result[i]
    }
    return result
}
```

## 2. Dijkstra 最短路径

```go
type Edge struct{ To, Weight int }

func Dijkstra(n int, adj [][]Edge, start int) []int {
    dist := make([]int, n)
    for i := range dist { dist[i] = math.MaxInt }
    dist[start] = 0

    // 最小堆（Go 1.18+）
    pq := &MinHeap[distItem]{}
    heap.Init(pq)
    heap.Push(pq, distItem{start, 0})

    for pq.Len() > 0 {
        item := heap.Pop(pq).(distItem)
        u, d := item.node, item.dist
        if d > dist[u] { continue }

        for _, e := range adj[u] {
            nd := d + e.Weight
            if nd < dist[e.To] {
                dist[e.To] = nd
                heap.Push(pq, distItem{e.To, nd})
            }
        }
    }
    return dist
}

type distItem struct {
    node int
    dist int
}

type MinHeap []distItem
func (h MinHeap) Len() int { return len(h) }
func (h MinHeap) Less(i, j int) bool { return h[i].dist < h[j].dist }
func (h MinHeap) Swap(i, j int) { h[i], h[j] = h[j], h[i] }
func (h *MinHeap) Push(v any) { *h = append(*h, v.(distItem)) }
func (h *MinHeap) Pop() any { old := *h; n := len(old); v := old[n-1]; *h = old[:n-1]; return v }
func (h MinHeap) Len() int { return len(h) } // 名字冲突需注意
```

## 完整对照表

| 算法 | TS | Go |
|------|-----|-----|
| 拓扑 Kahn | `inDegree` 数组 + 队列 | 同 |
| 拓扑 DFS | 三色标记 + 后序反转 | 同 |
| Dijkstra | 优先队列 | heap.Interface |
| 邻接表 | `[][number, number][]` | `[][]Edge` |
