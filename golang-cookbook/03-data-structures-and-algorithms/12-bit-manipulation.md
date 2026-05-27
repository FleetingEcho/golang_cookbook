# 位操作 — Bit Manipulation

## 常用模板

```go
// 检查第 k 位是否为 1（从 0 开始）
func checkBit(n, k int) bool { return n>>k&1 == 1 }

// 设置第 k 位为 1
func setBit(n, k int) int { return n | 1<<k }

// 清除第 k 位
func clearBit(n, k int) int { return n &^ (1 << k) }

// 翻转第 k 位
func toggleBit(n, k int) int { return n ^ 1<<k }

// 获取最低位的 1
func lowbit(n int) int { return n & -n }

// 清除最低位的 1
func removeLowbit(n int) int { return n & (n - 1) }

// 计算 1 的个数
func popCount(n int) int {
    count := 0
    for n != 0 {
        n &= n - 1
        count++
    }
    return count
}

// 内置方法
// bits.OnesCount(uint(n))

// 是否是 2 的幂
func isPowerOfTwo(n int) bool { return n > 0 && n&(n-1) == 0 }

// 枚举子集
func subsets(nums []int) [][]int {
    n := len(nums)
    result := make([][]int, 0, 1<<n)
    for mask := 0; mask < 1<<n; mask++ {
        subset := []int{}
        for i := 0; i < n; i++ {
            if mask>>i&1 != 0 {
                subset = append(subset, nums[i])
            }
        }
        result = append(result, subset)
    }
    return result
}

// 枚举子集的子集（S 的子集）
func submasks(S int) []int {
    var result []int
    for sub := S; sub > 0; sub = (sub - 1) & S {
        result = append(result, sub)
    }
    result = append(result, 0)
    return result
}

// 两数交换（无临时变量，面试技巧）
func swap(a, b int) (int, int) {
    a ^= b
    b ^= a
    a ^= b
    return a, b
}

// 异或的应用：找唯一不重复的数
func singleNumber(nums []int) int {
    xor := 0
    for _, v := range nums { xor ^= v }
    return xor
}
```

## 完整对照表

| 操作 | TS | Go |
|------|-----|-----|
| 检查位 | `(n>>k & 1) !== 0` | `n>>k&1 == 1` |
| 设置位 | `n \|= 1<<k` | `n \|= 1<<k` |
| 清除位 | `n &= ~(1<<k)` | `n &^= 1<<k` |
| 最低位 | `n & -n` | `n & -n` |
| 置零最低位 | `n & (n-1)` | `n & (n-1)` |
| 1 个数 | 手动 | `bits.OnesCount` |
| 整数范围 | `2^31-1` | `math.MaxInt`/`math.MaxInt64` |
