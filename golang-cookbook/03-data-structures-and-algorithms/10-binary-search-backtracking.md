# 二分查找与回溯

## 1. 二分查找

```go
// 标准二分
func BinarySearch(nums []int, target int) int {
    left, right := 0, len(nums)-1
    for left <= right {
        mid := left + (right-left)/2
        if nums[mid] == target {
            return mid
        } else if nums[mid] < target {
            left = mid + 1
        } else {
            right = mid - 1
        }
    }
    return -1
}

// 左边界（第一个 >= target）
func LowerBound(nums []int, target int) int {
    left, right := 0, len(nums)-1
    for left <= right {
        mid := left + (right-left)/2
        if nums[mid] >= target {
            right = mid - 1
        } else {
            left = mid + 1
        }
    }
    return left
}

// 右边界（最后一个 <= target）
func UpperBound(nums []int, target int) int {
    left, right := 0, len(nums)-1
    for left <= right {
        mid := left + (right-left)/2
        if nums[mid] <= target {
            left = mid + 1
        } else {
            right = mid - 1
        }
    }
    return right
}

// 旋转数组搜索
func RotatedSearch(nums []int, target int) int {
    left, right := 0, len(nums)-1
    for left <= right {
        mid := left + (right-left)/2
        if nums[mid] == target { return mid }
        if nums[left] <= nums[mid] { // 左半有序
            if target >= nums[left] && target < nums[mid] {
                right = mid - 1
            } else {
                left = mid + 1
            }
        } else { // 右半有序
            if target > nums[mid] && target <= nums[right] {
                left = mid + 1
            } else {
                right = mid - 1
            }
        }
    }
    return -1
}

// 泛型版本
func BinarySearchGeneric[T constraints.Ordered](data []T, target T) int {
    left, right := 0, len(data)-1
    for left <= right {
        mid := left + (right-left)/2
        if data[mid] == target { return mid }
        if data[mid] < target { left = mid + 1 } else { right = mid - 1 }
    }
    return -1
}
```

## 2. 回溯（Backtracking）

```go
// 全排列
func Permute(nums []int) [][]int {
    var result [][]int
    used := make([]bool, len(nums))
    var backtrack func(path []int)
    backtrack = func(path []int) {
        if len(path) == len(nums) {
            perm := make([]int, len(path))
            copy(perm, path)
            result = append(result, perm)
            return
        }
        for i := 0; i < len(nums); i++ {
            if used[i] { continue }
            used[i] = true
            backtrack(append(path, nums[i]))
            used[i] = false
        }
    }
    backtrack([]int{})
    return result
}

// 组合
func Combine(n, k int) [][]int {
    var result [][]int
    var backtrack func(start int, path []int)
    backtrack = func(start int, path []int) {
        if len(path) == k {
            comb := make([]int, len(path))
            copy(comb, path)
            result = append(result, comb)
            return
        }
        for i := start; i <= n; i++ {
            backtrack(i+1, append(path, i))
        }
    }
    backtrack(1, []int{})
    return result
}

// 子集
func Subsets(nums []int) [][]int {
    var result [][]int
    var backtrack func(start int, path []int)
    backtrack = func(start int, path []int) {
        subset := make([]int, len(path))
        copy(subset, path)
        result = append(result, subset)
        for i := start; i < len(nums); i++ {
            backtrack(i+1, append(path, nums[i]))
        }
    }
    backtrack(0, []int{})
    return result
}

// N 皇后
func SolveNQueens(n int) [][]string {
    var result [][]string
    board := make([][]byte, n)
    for i := range board {
        board[i] = make([]byte, n)
        for j := range board[i] { board[i][j] = '.' }
    }
    cols := make([]bool, n)
    diag1 := make([]bool, 2*n-1) // r + c
    diag2 := make([]bool, 2*n-1) // r - c + n - 1

    var backtrack func(r int)
    backtrack = func(r int) {
        if r == n {
            sol := make([]string, n)
            for i, row := range board { sol[i] = string(row) }
            result = append(result, sol)
            return
        }
        for c := 0; c < n; c++ {
            if cols[c] || diag1[r+c] || diag2[r-c+n-1] { continue }
            board[r][c] = 'Q'
            cols[c], diag1[r+c], diag2[r-c+n-1] = true, true, true
            backtrack(r + 1)
            board[r][c] = '.'
            cols[c], diag1[r+c], diag2[r-c+n-1] = false, false, false
        }
    }
    backtrack(0)
    return result
}
```

## 完整对照表

| 模式 | TS | Go |
|------|-----|-----|
| 二分 | `l<=r` `mid=Math.floor((l+r)/2)` | `left+(right-left)/2` |
| 左边界 | `while(l<=r)` | `for left <= right` |
| 回溯模板 | `backtrack(path,start)` | `var backtrack func(start int, path []T)` |
| 剪枝 | `if (used[i]) continue` | 同 |
| 深拷贝 | `[...path]` | `copy(make([]T,len(path)), path)` |
