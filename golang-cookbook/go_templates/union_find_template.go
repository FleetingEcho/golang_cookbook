// 并查集模板
package main

import "fmt"

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
		d.parent[x] = d.Find(d.parent[x])
	}
	return d.parent[x]
}

func (d *DSU) Union(x, y int) bool {
	px, py := d.Find(x), d.Find(y)
	if px == py {
		return false
	}
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

// ---------- 应用示例 ----------

// 岛屿数量
func numIslands(grid [][]byte) int {
	if len(grid) == 0 {
		return 0
	}
	m, n := len(grid), len(grid[0])
	dsu := NewDSU(m * n)
	zeros := 0

	dirs := [][2]int{{0, 1}, {1, 0}}
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
				zeros++
			}
		}
	}
	return dsu.Count() - zeros
}

// 连通网络的操作次数
func makeConnected(n int, connections [][]int) int {
	dsu := NewDSU(n)
	extra := 0
	for _, conn := range connections {
		if !dsu.Union(conn[0], conn[1]) {
			extra++
		}
	}
	// n - count = 当前连通分量数（合并后）
	need := dsu.Count() - 1
	if extra >= need {
		return need
	}
	return -1
}

func main() {
	// 示例
	dsu := NewDSU(5)
	dsu.Union(0, 1)
	dsu.Union(1, 2)
	dsu.Union(3, 4)

	fmt.Println("Connected(0,2):", dsu.Connected(0, 2)) // true
	fmt.Println("Connected(0,3):", dsu.Connected(0, 3)) // false
	fmt.Println("Components:", dsu.Count())              // 2
}
