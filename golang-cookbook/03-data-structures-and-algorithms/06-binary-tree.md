# 二叉树 — Binary Tree

## 1. 定义

```go
type TreeNode[T any] struct {
    Val   T
    Left  *TreeNode[T]
    Right *TreeNode[T]
}

// 创建
func NewTree[T any](vals []T, null T) *TreeNode[T] {
    if len(vals) == 0 { return nil }
    root := &TreeNode[T]{Val: vals[0]}
    queue := []*TreeNode[T]{root}
    i := 1
    for len(queue) > 0 && i < len(vals) {
        node := queue[0]; queue = queue[1:]
        if vals[i] != null {
            node.Left = &TreeNode[T]{Val: vals[i]}
            queue = append(queue, node.Left)
        }
        i++
        if i < len(vals) && vals[i] != null {
            node.Right = &TreeNode[T]{Val: vals[i]}
            queue = append(queue, node.Right)
        }
        i++
    }
    return root
}
```

## 2. 四种遍历

```go
// 前序：根 → 左 → 右
func Preorder[T any](root *TreeNode[T]) []T {
    var result []T
    var dfs func(*TreeNode[T])
    dfs = func(n *TreeNode[T]) {
        if n == nil { return }
        result = append(result, n.Val) // 根
        dfs(n.Left)                    // 左
        dfs(n.Right)                   // 右
    }
    dfs(root)
    return result
}

// 中序：左 → 根 → 右
func Inorder[T any](root *TreeNode[T]) []T {
    var result []T
    var dfs func(*TreeNode[T])
    dfs = func(n *TreeNode[T]) {
        if n == nil { return }
        dfs(n.Left)
        result = append(result, n.Val)
        dfs(n.Right)
    }
    dfs(root)
    return result
}

// 后序：左 → 右 → 根
func Postorder[T any](root *TreeNode[T]) []T {
    var result []T
    var dfs func(*TreeNode[T])
    dfs = func(n *TreeNode[T]) {
        if n == nil { return }
        dfs(n.Left)
        dfs(n.Right)
        result = append(result, n.Val)
    }
    dfs(root)
    return result
}

// 层序（BFS）
func LevelOrder[T any](root *TreeNode[T]) [][]T {
    if root == nil { return nil }
    var result [][]T
    queue := []*TreeNode[T]{root}
    for len(queue) > 0 {
        n := len(queue)
        level := make([]T, n)
        for i := 0; i < n; i++ {
            node := queue[0]; queue = queue[1:]
            level[i] = node.Val
            if node.Left != nil { queue = append(queue, node.Left) }
            if node.Right != nil { queue = append(queue, node.Right) }
        }
        result = append(result, level)
    }
    return result
}

// 前序（迭代，栈）
func PreorderIter[T any](root *TreeNode[T]) []T {
    var result []T
    stack := []*TreeNode[T]{}
    cur := root
    for cur != nil || len(stack) > 0 {
        for cur != nil {
            result = append(result, cur.Val)
            stack = append(stack, cur)
            cur = cur.Left
        }
        cur = stack[len(stack)-1]; stack = stack[:len(stack)-1]
        cur = cur.Right
    }
    return result
}
```

## 3. 常见操作

```go
// 最大深度
func MaxDepth[T any](root *TreeNode[T]) int {
    if root == nil { return 0 }
    return 1 + max(MaxDepth(root.Left), MaxDepth(root.Right))
}

// 对称检查
func IsSymmetric[T any](root *TreeNode[T]) bool {
    var check func(l, r *TreeNode[T]) bool
    check = func(l, r *TreeNode[T]) bool {
        if l == nil && r == nil { return true }
        if l == nil || r == nil { return false }
        return l.Val == r.Val && check(l.Left, r.Right) && check(l.Right, r.Left)
    }
    return check(root.Left, root.Right)
}

// 直径
func Diameter[T any](root *TreeNode[T]) int {
    ans := 0
    var dfs func(*TreeNode[T]) int
    dfs = func(n *TreeNode[T]) int {
        if n == nil { return 0 }
        l, r := dfs(n.Left), dfs(n.Right)
        ans = max(ans, l+r)
        return 1 + max(l, r)
    }
    dfs(root)
    return ans
}

// 最低公共祖先
func LCA[T any](root, p, q *TreeNode[T]) *TreeNode[T] {
    if root == nil || root == p || root == q { return root }
    left := LCA(root.Left, p, q)
    right := LCA(root.Right, p, q)
    if left != nil && right != nil { return root }
    if left != nil { return left }
    return right
}
```

## 完整对照表

| 操作 | TS | Go |
|------|-----|-----|
| 节点 | `class TreeNode { val; left; right }` | `struct { Val T; Left, Right *TreeNode[T] }` |
| 前序 | 根左右 | 递归 / 栈迭代 |
| 中序 | 左根右 | 递归 / 栈迭代 |
| 后序 | 左右根 | 递归 / 双栈 |
| 层序 | BFS 队列 | 队列 |
| 深度 | 递归 max | 递归 1+max |
| LCA | 递归 | 递归 |
| 泛型 | `TreeNode<T>` | `TreeNode[T any]` |
