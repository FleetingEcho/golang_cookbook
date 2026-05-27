# 栈与队列 — Stack & Queue

## 1. 栈（Stack）— LIFO

```go
// Go — slice 就是栈
type Stack[T any] []T

func (s *Stack[T]) Push(v T) {
    *s = append(*s, v)
}

func (s *Stack[T]) Pop() (T, bool) {
    if len(*s) == 0 {
        var zero T
        return zero, false
    }
    v := (*s)[len(*s)-1]
    *s = (*s)[:len(*s)-1]
    return v, true
}

func (s *Stack[T]) Peek() (T, bool) {
    if len(*s) == 0 {
        var zero T
        return zero, false
    }
    return (*s)[len(*s)-1], true
}

func (s *Stack[T]) Len() int { return len(*s) }

// 使用
s := Stack[int]{}
s.Push(1)
s.Push(2)
v, _ := s.Pop() // 2
```

## 2. 队列（Queue）— FIFO

```go
// 方式 1：slice（效率低，Pop 需要 O(n) 移动）
type SimpleQueue[T any] []T
func (q *SimpleQueue[T]) Enqueue(v T) { *q = append(*q, v) }
func (q *SimpleQueue[T]) Dequeue() (T, bool) {
    if len(*q) == 0 {
        var zero T
        return zero, false
    }
    v := (*q)[0]
    *q = (*q)[1:] // O(n) 移动
    return v, true
}

// 方式 2：环形队列（推荐，O(1)）
type Queue[T any] struct {
    data []T
    head, tail, size, cap int
}

func NewQueue[T any](capacity int) *Queue[T] {
    return &Queue[T]{data: make([]T, capacity), cap: capacity}
}

func (q *Queue[T]) Enqueue(v T) bool {
    if q.size == q.cap { return false }
    q.data[q.tail] = v
    q.tail = (q.tail + 1) % q.cap
    q.size++
    return true
}

func (q *Queue[T]) Dequeue() (T, bool) {
    if q.size == 0 {
        var zero T
        return zero, false
    }
    v := q.data[q.head]
    q.head = (q.head + 1) % q.cap
    q.size--
    return v, true
}

func (q *Queue[T]) Len() int { return q.size }
```

## 3. 双端队列（Deque）

```go
type Deque[T any] struct {
    data []T
    head, tail, size, cap int
}

func NewDeque[T any](capacity int) *Deque[T] {
    return &Deque[T]{data: make([]T, capacity+1), cap: capacity+1}
}

func (d *Deque[T]) PushFront(v T) bool {
    if d.size == d.cap-1 { return false }
    d.head = (d.head - 1 + d.cap) % d.cap
    d.data[d.head] = v
    d.size++
    return true
}

func (d *Deque[T]) PushBack(v T) bool {
    if d.size == d.cap-1 { return false }
    d.data[d.tail] = v
    d.tail = (d.tail + 1) % d.cap
    d.size++
    return true
}

func (d *Deque[T]) PopFront() (T, bool) {
    if d.size == 0 {
        var zero T
        return zero, false
    }
    v := d.data[d.head]
    d.head = (d.head + 1) % d.cap
    d.size--
    return v, true
}

func (d *Deque[T]) PopBack() (T, bool) {
    if d.size == 0 {
        var zero T
        return zero, false
    }
    d.tail = (d.tail - 1 + d.cap) % d.cap
    return d.data[d.tail], true
}
```

## 完整对照表

| 操作 | TS Array | Go Stack/Queue |
|------|----------|---------------|
| 入栈 | `arr.push(v)` | `s.Push(v)` |
| 出栈 | `arr.pop()` | `s.Pop()` |
| 入队 | `arr.push(v)` | `q.Enqueue(v)` |
| 出队 | `arr.shift()` | `q.Dequeue()` |
| 查看栈顶 | `arr[arr.length-1]` | `s.Peek()` |
| 长度 | `arr.length` | `s.Len()` |
| 双端队列 | 无原生 | `Deque[T]` |
