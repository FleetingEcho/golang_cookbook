# Golang Cookbook & Issue Tracker

> A monorepo containing a Go learning cookbook and a full-stack Issue Tracker demo (Go backend + React frontend).

## Quick Links

| Section | Description | Go To |
|---------|-------------|-------|
| 📖 **Go Cookbook** | TS→Go language reference, concurrency, algorithms, engineering practices | [`golang-cookbook/`](./golang-cookbook/) |
| 🖥️ **Issue Tracker API** | Go backend (chi + sqlite), ported from Rust/Axum | [`projects/server/`](./projects/server/) |
| 🌐 **Issue Tracker Frontend** | React + Vite SPA | [`projects/web/`](./projects/web/) |
| 🧩 **Go Templates** | Runnable `package main` templates for coding practice | [`golang-cookbook/go_templates/`](./golang-cookbook/go_templates/) |

---

## 📖 Go Cookbook

A side-by-side reference for TypeScript developers learning Go. Covers 35+ topics across 4 chapters:

```
golang-cookbook/
├── 01-language/          # 13 files — variables, primitives, functions, control flow,
│                         # errors, pointers, nil, array vs slice, structs, maps,
│                         # generics, type assertions
├── 02-concurrency/       # 6 files — goroutines, channels, select, WaitGroup, mutex, context
├── 03-data-structures-and-algorithms/  # 12 files — stack, queue, linked list, heap,
│                                        # union-find, trie, binary tree, graph (DFS/BFS/Dijkstra),
│                                        # topological sort, sliding window, DP patterns, bit ops
├── 04-engineering/       # 4 files — struct methods, interfaces, JSON, testing
└── go_templates/         # 4 runnable .go files for coding practice
```

## 🖥️ Issue Tracker API

A full CRUD backend rewritten from Rust/Axum to Go.

**Tech**: Go 1.23+, [chi/v5](https://github.com/go-chi/chi), [modernc.org/sqlite](https://gitlab.com/cznic/sqlite) (zero CGO), `log/slog`

```bash
cd projects/server
make seed    # reset DB + seed 20 issues, 14 labels, comments
make run     # start on http://127.0.0.1:3001
```

**Features**:
- Issues CRUD with pagination, filtering, search, label association
- Comments CRUD
- Labels CRUD with issue attachment/detachment
- File upload/download (10MB limit, path-traversal safe storage)
- Middleware: API key auth, request ID, CORS, structured logging
- 6 integration tests with in-memory SQLite

## 🌐 Issue Tracker Frontend

A React SPA that consumes the API.

**Tech**: React 19, Vite, TypeScript

```bash
cd projects/web
npm install
npm run dev   # start on http://127.0.0.1:5173
```

## Go Workspace

The root [`go.work`](./go.work) file ties together all Go modules in this repo:

```go 1.23.0

use (
    ./golang-cookbook/go_templates
    ./projects/server
)
```

---

## License

MIT
