# 并查集 — Union-Find / DSU

## 并查集（泛型版）

```go
type DSU struct {
    parent []int
    rank   []int
    count  int // 连通分量数
}

func NewDSU(n int) *DSU {
    parent := make([]int, n)
    rank := make([]int, n)
    for i := range parent {
        parent[i] = i
        rank[i] = 1
    }
    return &DSU{parent: parent, rank: rank, count: n}
}

func (d *DSU) Find(x int) int {
    if d.parent[x] != x {
        d.parent[x] = d.Find(d.parent[x]) // 路径压缩
    }
    return d.parent[x]
}

func (d *DSU) Union(x, y int) bool {
    px, py := d.Find(x), d.Find(y)
    if px == py { return false }

    // 按秩合并
    if d.rank[px] < d.rank[py] {
        px, py = py, px
    }
    d.parent[py] = px
    if d.rank[px] == d.rank[py] {
        d.rank[px]++
    }
    d.count--
    return true
}

func (d *DSU) Connected(x, y int) bool {
    return d.Find(x) == d.Find(y)
}

func (d *DSU) Count() int { return d.count }

// 泛型版本（适用于非整数元素）
type DSUGeneric[T comparable] struct {
    parent map[T]T
    rank   map[T]int
    count  int
}

func NewDSUGeneric[T comparable]() *DSUGeneric[T] {
    return &DSUGeneric[T]{
        parent: make(map[T]T),
        rank:   make(map[T]int),
    }
}

func (d *DSUGeneric[T]) Add(x T) {
    if _, ok := d.parent[x]; !ok {
        d.parent[x] = x
        d.rank[x] = 1
        d.count++
    }
}

func (d *DSUGeneric[T]) Find(x T) T {
    if d.parent[x] != x {
        d.parent[x] = d.Find(d.parent[x])
    }
    return d.parent[x]
}

func (d *DSUGeneric[T]) Union(x, y T) bool {
    d.Add(x); d.Add(y)
    px, py := d.Find(x), d.Find(y)
    if px == py { return false }
    if d.rank[px] < d.rank[py] {
        px, py = py, px
    }
    d.parent[py] = px
    if d.rank[px] == d.rank[py] {
        d.rank[px]++
    }
    d.count--
    return true
}
```

## 应用：岛屿数量

```go
func numIslands(grid [][]byte) int {
    if len(grid) == 0 { return 0 }
    m, n := len(grid), len(grid[0])
    dsu := NewDSU(m * n)

    dirs := [][2]int{{0, 1}, {1, 0}} // 右、下
    for i := 0; i < m; i++ {
        for j := 0; j < n; j++ {
            if grid[i][j] == '1' {
                for _, d := range dirs {
                    ni, nj := i+d[0], j+d[1]
                    if ni < m && nj < n && grid[ni][nj] == '1' {
                        dsu.Union(i*n+j, ni*n+nj)
                    }
                }
            } else {
                dsu.count-- // 水域不计入连通分量
            }
        }
    }
    return dsu.Count()
}
```

## 完整对照表

| 操作 | TS | Go |
|------|-----|-----|
| 创建 | `class DSU { parent: number[] }` | `NewDSU(n)` |
| 查找 | `find(x)` 递归/迭代 | `Find(x)` 路径压缩 |
| 合并 | `union(x, y)` | `Union(x, y)` → bool |
| 连通 | `find(x) === find(y)` | `Connected(x, y)` |
| 分量 | 追踪 count | `Count()` |
| 泛型 | `Map<K, K>` | `DSUGeneric[T comparable]` |
