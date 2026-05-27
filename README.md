# Golang Cookbook

> A hands-on Go reference for TypeScript developers — language fundamentals, concurrency, data structures, algorithms, and engineering practices.

---

## Quick Navigation

### 01 — Language Fundamentals

| Topic | File |
|-------|------|
| Variables & Constants | [`01-variables`](./golang-cookbook/01-language/01-variables.md) |
| Primitives (12 numeric types) | [`02-primitives`](./golang-cookbook/01-language/02-primitives.md) |
| Functions & `defer` | [`03-functions`](./golang-cookbook/01-language/03-functions.md) |
| Control Flow (`if`/`for`/`switch`) | [`04-control-flow`](./golang-cookbook/01-language/04-control-flow.md) |
| Errors (no try/catch) | [`05-errors`](./golang-cookbook/01-language/05-errors.md) |
| Pointers | [`06-pointers`](./golang-cookbook/01-language/06-pointers.md) |
| Nil vs Zero Values | [`07-nil-and-zero`](./golang-cookbook/01-language/07-nil-and-zero.md) |
| Array vs Slice | [`08-array-vs-slice`](./golang-cookbook/01-language/08-array-vs-slice.md) |
| Structs (Objects) | [`09-object-vs-struct`](./golang-cookbook/01-language/09-object-vs-struct.md) |
| Tuples (not built-in) | [`10-tuple`](./golang-cookbook/01-language/10-tuple.md) |
| Maps | [`11-map`](./golang-cookbook/01-language/11-map.md) |
| Generics (Go 1.18+) | [`12-generics`](./golang-cookbook/01-language/12-generics.md) |
| Type Assertions | [`13-type-assertion`](./golang-cookbook/01-language/13-type-assertion.md) |

### 02 — Concurrency

| Topic | File |
|-------|------|
| Goroutines vs Promises | [`01-goroutine`](./golang-cookbook/02-concurrency/01-goroutine.md) |
| Channels | [`02-channel`](./golang-cookbook/02-concurrency/02-channel.md) |
| `select` — Multiplexing | [`03-select`](./golang-cookbook/02-concurrency/03-select.md) |
| WaitGroup | [`04-waitgroup`](./golang-cookbook/02-concurrency/04-waitgroup.md) |
| Mutex & Race Conditions | [`05-mutex`](./golang-cookbook/02-concurrency/05-mutex.md) |
| Context (timeout/cancel) | [`06-context`](./golang-cookbook/02-concurrency/06-context.md) |

### 03 — Data Structures & Algorithms

| Topic | File |
|-------|------|
| Stack & Queue | [`01-stack-queue`](./golang-cookbook/03-data-structures-and-algorithms/01-stack-queue.md) |
| Linked List | [`02-linked-list`](./golang-cookbook/03-data-structures-and-algorithms/02-linked-list.md) |
| Heap / Priority Queue | [`03-heap`](./golang-cookbook/03-data-structures-and-algorithms/03-heap.md) |
| Union-Find (DSU) | [`04-union-find`](./golang-cookbook/03-data-structures-and-algorithms/04-union-find.md) |
| Trie | [`05-trie`](./golang-cookbook/03-data-structures-and-algorithms/05-trie.md) |
| Binary Tree | [`06-binary-tree`](./golang-cookbook/03-data-structures-and-algorithms/06-binary-tree.md) |
| Graph (DFS/BFS/Dijkstra) | [`07-graph`](./golang-cookbook/03-data-structures-and-algorithms/07-graph.md) |
| Topological Sort | [`08-topological-sort`](./golang-cookbook/03-data-structures-and-algorithms/08-topological-sort.md) |
| Two Pointers & Sliding Window | [`09-two-pointers-sliding-window`](./golang-cookbook/03-data-structures-and-algorithms/09-two-pointers-sliding-window.md) |
| Binary Search & Backtracking | [`10-binary-search-backtracking`](./golang-cookbook/03-data-structures-and-algorithms/10-binary-search-backtracking.md) |
| DP Patterns (Knapsack/LIS/LCS/Kadane) | [`11-dp-patterns`](./golang-cookbook/03-data-structures-and-algorithms/11-dp-patterns.md) |
| Bit Manipulation | [`12-bit-manipulation`](./golang-cookbook/03-data-structures-and-algorithms/12-bit-manipulation.md) |

### 04 — Engineering Practices

| Topic | File |
|-------|------|
| Struct Methods & Interfaces | [`01-struct-methods-interfaces`](./golang-cookbook/04-engineering/01-struct-methods-interfaces.md) |
| JSON Serialization & Error Wrapping | [`02-json-errors`](./golang-cookbook/04-engineering/02-json-errors.md) |
| I/O & Module System | [`03-io-modules`](./golang-cookbook/04-engineering/03-io-modules.md) |
| Testing (`testing` package) | [`04-testing`](./golang-cookbook/04-engineering/04-testing.md) |

### Runnable Templates

| Template | File |
|----------|------|
| I/O (ACM/contest style) | [`io_template.go`](./golang-cookbook/go_templates/io_template.go) |
| Binary Tree | [`tree_template.go`](./golang-cookbook/go_templates/tree_template.go) |
| Graph (DFS/BFS/Topo/Dijkstra) | [`graph_template.go`](./golang-cookbook/go_templates/graph_template.go) |
| Union-Find | [`union_find_template.go`](./golang-cookbook/go_templates/union_find_template.go) |

```bash
# Run any template directly
go run golang-cookbook/go_templates/io_template.go
```

---

## Repository Structure

```
.
├── README.md                           ← you are here
├── go.work                             ← Go workspace (golang-cookbook + projects)
├── golang-cookbook/                    ← Go cookbook (35+ markdown files)
│   ├── 01-language/                    ← 13 topics
│   ├── 02-concurrency/                 ← 6 topics
│   ├── 03-data-structures-and-algorithms/  ← 12 topics
│   ├── 04-engineering/                 ← 4 topics
│   ├── go_templates/                   ← 4 runnable .go files
│   └── README.md
└── projects/                           ← full-stack Issue Tracker demo
    ├── server/                         ← Go API (chi + sqlite)
    └── web/                            ← React + Vite SPA
```

## Go Version Requirements

- **Go 1.18+** — generics
- **Go 1.21+** — `clear()`, `cmp`, `slices`, `maps`
- **Go 1.22+** — loop variable semantics, integer `range`
- **Go 1.23+** — range-over-func (used in `projects/server/`)

---

## License

MIT
