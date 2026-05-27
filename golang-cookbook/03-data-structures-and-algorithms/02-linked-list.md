# 链表 — Linked List

## 1. 单链表

```go
type ListNode[T any] struct {
    Val  T
    Next *ListNode[T]
}

// 创建
func NewList[T any](vals ...T) *ListNode[T] {
    dummy := &ListNode[T]{}
    cur := dummy
    for _, v := range vals {
        cur.Next = &ListNode[T]{Val: v}
        cur = cur.Next
    }
    return dummy.Next
}

// 遍历
func Traverse[T any](head *ListNode[T]) []T {
    var result []T
    for cur := head; cur != nil; cur = cur.Next {
        result = append(result, cur.Val)
    }
    return result
}

// 反转
func Reverse[T any](head *ListNode[T]) *ListNode[T] {
    var prev *ListNode[T]
    cur := head
    for cur != nil {
        next := cur.Next
        cur.Next = prev
        prev = cur
        cur = next
    }
    return prev
}

// 快慢指针找中点
func Middle[T any](head *ListNode[T]) *ListNode[T] {
    slow, fast := head, head
    for fast != nil && fast.Next != nil {
        slow = slow.Next
        fast = fast.Next.Next
    }
    return slow
}

// 合并两个有序链表
func Merge[T constraints.Ordered](l1, l2 *ListNode[T]) *ListNode[T] {
    dummy := &ListNode[T]{}
    cur := dummy
    for l1 != nil && l2 != nil {
        if l1.Val < l2.Val {
            cur.Next = l1
            l1 = l1.Next
        } else {
            cur.Next = l2
            l2 = l2.Next
        }
        cur = cur.Next
    }
    if l1 != nil { cur.Next = l1 }
    if l2 != nil { cur.Next = l2 }
    return dummy.Next
}

// 是否有环
func HasCycle[T any](head *ListNode[T]) bool {
    slow, fast := head, head
    for fast != nil && fast.Next != nil {
        slow = slow.Next
        fast = fast.Next.Next
        if slow == fast { return true }
    }
    return false
}
```

## 2. 双链表

```go
type DoublyListNode[T any] struct {
    Val  T
    Prev *DoublyListNode[T]
    Next *DoublyListNode[T]
}

type DoublyList[T any] struct {
    Head, Tail *DoublyListNode[T]
    Size       int
}

func (l *DoublyList[T]) Append(v T) {
    node := &DoublyListNode[T]{Val: v}
    if l.Head == nil {
        l.Head, l.Tail = node, node
    } else {
        node.Prev = l.Tail
        l.Tail.Next = node
        l.Tail = node
    }
    l.Size++
}

func (l *DoublyList[T]) Remove(node *DoublyListNode[T]) {
    if node.Prev != nil {
        node.Prev.Next = node.Next
    } else {
        l.Head = node.Next
    }
    if node.Next != nil {
        node.Next.Prev = node.Prev
    } else {
        l.Tail = node.Prev
    }
    l.Size--
}

func (l *DoublyList[T]) MoveToFront(node *DoublyListNode[T]) {
    l.Remove(node)
    node.Prev = nil
    node.Next = l.Head
    if l.Head != nil {
        l.Head.Prev = node
    }
    l.Head = node
    if l.Tail == nil {
        l.Tail = node
    }
    l.Size++
}
```

## 完整对照表

| 操作 | TS | Go |
|------|-----|-----|
| 节点 | `class Node { val; next }` | `type ListNode[T] struct { Val T; Next *ListNode[T] }` |
| 反转 | 递归/迭代 | 三指针迭代 |
| 快慢指针 | `fast = fast.next.next` | 同上 |
| 合并 | 递归/迭代 | dummy 节点 + 迭代 |
| LRU 缓存 | `Map` + 链表 | 双链表 + `map` |
| 泛型 | `Node<T>` | `ListNode[T any]` |
