// 二叉树模板
package main

import "fmt"

type TreeNode struct {
	Val   int
	Left  *TreeNode
	Right *TreeNode
}

// 从层序数组创建二叉树（-1 表示 nil）
func buildTree(vals []int) *TreeNode {
	if len(vals) == 0 {
		return nil
	}
	root := &TreeNode{Val: vals[0]}
	queue := []*TreeNode{root}
	i := 1
	for len(queue) > 0 && i < len(vals) {
		node := queue[0]
		queue = queue[1:]
		if vals[i] != -1 {
			node.Left = &TreeNode{Val: vals[i]}
			queue = append(queue, node.Left)
		}
		i++
		if i < len(vals) && vals[i] != -1 {
			node.Right = &TreeNode{Val: vals[i]}
			queue = append(queue, node.Right)
		}
		i++
	}
	return root
}

// 前序
func preorder(root *TreeNode) []int {
	var result []int
	var dfs func(*TreeNode)
	dfs = func(n *TreeNode) {
		if n == nil {
			return
		}
		result = append(result, n.Val)
		dfs(n.Left)
		dfs(n.Right)
	}
	dfs(root)
	return result
}

// 中序
func inorder(root *TreeNode) []int {
	var result []int
	var dfs func(*TreeNode)
	dfs = func(n *TreeNode) {
		if n == nil {
			return
		}
		dfs(n.Left)
		result = append(result, n.Val)
		dfs(n.Right)
	}
	dfs(root)
	return result
}

// 后序
func postorder(root *TreeNode) []int {
	var result []int
	var dfs func(*TreeNode)
	dfs = func(n *TreeNode) {
		if n == nil {
			return
		}
		dfs(n.Left)
		dfs(n.Right)
		result = append(result, n.Val)
	}
	dfs(root)
	return result
}

// 层序
func levelOrder(root *TreeNode) [][]int {
	if root == nil {
		return nil
	}
	var result [][]int
	queue := []*TreeNode{root}
	for len(queue) > 0 {
		n := len(queue)
		level := make([]int, n)
		for i := 0; i < n; i++ {
			node := queue[0]
			queue = queue[1:]
			level[i] = node.Val
			if node.Left != nil {
				queue = append(queue, node.Left)
			}
			if node.Right != nil {
				queue = append(queue, node.Right)
			}
		}
		result = append(result, level)
	}
	return result
}

func main() {
	// 示例：构建二叉树 [3,9,20,-1,-1,15,7]
	vals := []int{3, 9, 20, -1, -1, 15, 7}
	root := buildTree(vals)

	fmt.Println("前序:", preorder(root))
	fmt.Println("中序:", inorder(root))
	fmt.Println("后序:", postorder(root))
	fmt.Println("层序:", levelOrder(root))
}
