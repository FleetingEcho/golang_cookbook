# 堆与优先队列 — Heap & Priority Queue

```go
// Go 标准库 container/heap 实现的是最小堆
// 需要实现 heap.Interface（Len / Less / Swap / Push / Pop）

// 1. 通用最小堆（Go 1.18+ 泛型版本）
type MinHeap[T constraints.Ordered] []T

func (h MinHeap[T]) Len() int           { return len(h) }
func (h MinHeap[T]) Less(i, j int) bool  { return h[i] < h[j] }
func (h MinHeap[T]) Swap(i, j int)       { h[i], h[j] = h[j], h[i] }

func (h *MinHeap[T]) Push(v any) {
    *h = append(*h, v.(T))
}

func (h *MinHeap[T]) Pop() any {
    old := *h
    n := len(old)
    v := old[n-1]
    *h = old[:n-1]
    return v
}

// 辅助函数
func (h *MinHeap[T]) PushVal(v T) { heap.Push(h, v) }
func (h *MinHeap[T]) PopVal() T   { return heap.Pop(h).(T) }

// 使用
h := &MinHeap[int]{}
heap.Init(h)
h.PushVal(3)
h.PushVal(1)
h.PushVal(2)
fmt.Println(h.PopVal()) // 1

// 2. 最大堆（自定义 Less）
type MaxHeap[T constraints.Ordered] []T

func (h MaxHeap[T]) Len() int           { return len(h) }
func (h MaxHeap[T]) Less(i, j int) bool  { return h[i] > h[j] } // 反着写
func (h MaxHeap[T]) Swap(i, j int)       { h[i], h[j] = h[j], h[i] }
func (h *MaxHeap[T]) Push(v any)         { *h = append(*h, v.(T)) }
func (h *MaxHeap[T]) Pop() any {
    old := *h; n := len(old); v := old[n-1]
    *h = old[:n-1]; return v
}
func (h *MaxHeap[T]) PushVal(v T) { heap.Push(h, v) }
func (h *MaxHeap[T]) PopVal() T   { return heap.Pop(h).(T) }

// 3. 优先队列（自定义排序）
type Item struct {
    Value    string
    Priority int
}

type PriorityQueue []*Item

func (pq PriorityQueue) Len() int { return len(pq) }
func (pq PriorityQueue) Less(i, j int) bool {
    return pq[i].Priority < pq[j].Priority // 最小优先
}
func (pq PriorityQueue) Swap(i, j int) { pq[i], pq[j] = pq[j], pq[i] }
func (pq *PriorityQueue) Push(v any)   { *pq = append(*pq, v.(*Item)) }
func (pq *PriorityQueue) Pop() any {
    old := *pq; n := len(old); v := old[n-1]
    *pq = old[:n-1]; return v
}

// 4. 一行的 Top-K（用 sort.Slice 替代堆，小数据量足够）
func topK[T constraints.Ordered](data []T, k int) []T {
    sort.Slice(data, func(i, j int) bool { return data[i] > data[j] })
    return data[:min(k, len(data))]
}
```

## 完整对照表

| 操作 | TS | Go |
|------|-----|-----|
| 最小堆 | `new MinPriorityQueue()` | 实现 heap.Interface |
| 最大堆 | `new MaxPriorityQueue()` | 反写 Less |
| Top-K | `arr.sort().slice(0,k)` | sort + slice / 堆 |
| 优先队列 | 自定义 comparator | 自定义 struct + Less |
| 泛型 | `Heap<T>` | `Heap[T constraints.Ordered]` |
| 内建 | 无 | `container/heap` |
