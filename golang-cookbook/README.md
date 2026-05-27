# Golang Cookbook 🚀

> TypeScript 开发者快速上手 Go 的对照手册
> Go 版本：**1.22+** | 泛型：**Go 1.18+**

## 快速定位

| 如果你要找 | 去哪里 |
|-----------|--------|
| "Go 的 const 和 TS 的 const 有什么区别？" | [01-variables](./01-language/01-variables.md) |
| "Go 竟然有 12 种数值类型？" | [02-primitives](./01-language/02-primitives.md) |
| "defer 怎么用？" | [03-functions](./01-language/03-functions.md) |
| "Go 有 while 吗？" | [04-control-flow](./01-language/04-control-flow.md) |
| "try/catch 在 Go 里怎么搞？" | [05-errors](./01-language/05-errors.md) |
| "指针到底怎么回事？" | [06-pointers](./01-language/06-pointers.md) |
| "nil vs undefined" | [07-nil-and-zero](./01-language/07-nil-and-zero.md) |
| "Array vs Slice，什么时候用哪个？" | [08-array-vs-slice](./01-language/08-array-vs-slice.md) |
| "结构体怎么定义？" | [09-object-vs-struct](./01-language/09-object-vs-struct.md) |
| "元组呢？" | [10-tuple](./01-language/10-tuple.md) |
| "Go 的 Map 怎么用？" | [11-map](./01-language/11-map.md) |
| "泛型怎么写？" | [12-generics](./01-language/12-generics.md) |
| "类型断言 vs as" | [13-type-assertion](./01-language/13-type-assertion.md) |
| "goroutine 和 Promise 的区别？" | [02-concurrency/01-goroutine](./02-concurrency/01-goroutine.md) |
| "channel 的完整用法" | [02-concurrency/02-channel](./02-concurrency/02-channel.md) |
| "select 多路复用" | [02-concurrency/03-select](./02-concurrency/03-select.md) |
| "等待所有 goroutine 完成" | [02-concurrency/04-waitgroup](./02-concurrency/04-waitgroup.md) |
| "锁，数据竞争怎么办？" | [02-concurrency/05-mutex](./02-concurrency/05-mutex.md) |
| "context 超时控制" | [02-concurrency/06-context](./02-concurrency/06-context.md) |
| "栈、队列用 Go 怎么写？" | [03/01-stack-queue](./03-data-structures-and-algorithms/01-stack-queue.md) |
| "链表" | [03/02-linked-list](./03-data-structures-and-algorithms/02-linked-list.md) |
| "堆与优先队列" | [03/03-heap](./03-data-structures-and-algorithms/03-heap.md) |
| "并查集 DSU" | [03/04-union-find](./03-data-structures-and-algorithms/04-union-find.md) |
| "前缀树 Trie" | [03/05-trie](./03-data-structures-and-algorithms/05-trie.md) |
| "二叉树遍历模板" | [03/06-binary-tree](./03-data-structures-and-algorithms/06-binary-tree.md) |
| "图 DFS/BFS/Dijkstra" | [03/07-graph](./03-data-structures-and-algorithms/07-graph.md) |
| "拓扑排序与最短路径" | [03/08-topological-sort](./03-data-structures-and-algorithms/08-topological-sort.md) |
| "双指针与滑动窗口" | [03/09-two-pointers-sliding-window](./03-data-structures-and-algorithms/09-two-pointers-sliding-window.md) |
| "二分查找与回溯" | [03/10-binary-search-backtracking](./03-data-structures-and-algorithms/10-binary-search-backtracking.md) |
| "DP 模式（背包/LIS/LCS/Kadane）" | [03/11-dp-patterns](./03-data-structures-and-algorithms/11-dp-patterns.md) |
| "位操作模板" | [03/12-bit-manipulation](./03-data-structures-and-algorithms/12-bit-manipulation.md) |
| "结构体与方法、接口" | [04/01-struct-methods-interfaces](./04-engineering/01-struct-methods-interfaces.md) |
| "JSON 序列化 + Error 包装" | [04/02-json-errors](./04-engineering/02-json-errors.md) |
| "I/O 与模块系统" | [04/03-io-modules](./04-engineering/03-io-modules.md) |
| "测试（testing 包）" | [04/04-testing](./04-engineering/04-testing.md) |
| "可运行的刷题模板" | [templates/](./templates/) |

## 结构说明

```
golang-cookbook/
├── 01-language/          # 语言基础对照（13 个文件）
├── 02-concurrency/       # 并发（6 个文件）
├── 03-data-structures-and-algorithms/  # 数据结构与算法（12 个文件）
├── 04-engineering/       # 工程实践（4 个文件）
└── templates/            # 可运行的 .go 模板（4 个文件）
```

## 使用方式

```bash
# 阅读 .md 文件
# 运行 .go 模板
cd templates
go run io_template.go
go run tree_template.go
go run graph_template.go
```

所有 `.go` 文件都是完整的 `package main`，可直接 `go run`。

## Go 版本要求

- **Go 1.18+** — 泛型语法
- **Go 1.21+** — `clear()` / `cmp` / `slices` / `maps` 包
- **Go 1.22+** — 循环变量语义改进 / 整数 `range`

推荐使用 Go 1.22+。
