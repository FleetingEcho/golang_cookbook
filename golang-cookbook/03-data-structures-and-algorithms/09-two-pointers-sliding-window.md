# 双指针 — Two Pointers

## 1. 相向双指针（两数之和/回文）

```go
func TwoSum(nums []int, target int) []int {
    left, right := 0, len(nums)-1
    for left < right {
        sum := nums[left] + nums[right]
        if sum == target {
            return []int{left, right}
        } else if sum < target {
            left++
        } else {
            right--
        }
    }
    return nil
}

func IsPalindrome(s string) bool {
    left, right := 0, len(s)-1
    for left < right {
        if s[left] != s[right] { return false }
        left++
        right--
    }
    return true
}

// 三数之和
func ThreeSum(nums []int) [][]int {
    slices.Sort(nums)
    var result [][]int
    n := len(nums)
    for i := 0; i < n-2; i++ {
        if i > 0 && nums[i] == nums[i-1] { continue }
        left, right := i+1, n-1
        for left < right {
            sum := nums[i] + nums[left] + nums[right]
            if sum == 0 {
                result = append(result, []int{nums[i], nums[left], nums[right]})
                for left < right && nums[left] == nums[left+1] { left++ }
                for left < right && nums[right] == nums[right-1] { right-- }
                left++; right--
            } else if sum < 0 {
                left++
            } else {
                right--
            }
        }
    }
    return result
}
```

# 滑动窗口 — Sliding Window

## 2. 固定窗口

```go
func MaxSumFixedWindow(nums []int, k int) int {
    sum := 0
    for i := 0; i < k; i++ { sum += nums[i] }
    maxSum := sum
    for i := k; i < len(nums); i++ {
        sum += nums[i] - nums[i-k]
        maxSum = max(maxSum, sum)
    }
    return maxSum
}
```

## 3. 可变窗口（无重复/至少条件）

```go
// 无重复字符的最长子串
func LengthOfLongestSubstring(s string) int {
    lastIdx := [256]int{}   // 字符最后出现位置
    for i := range lastIdx { lastIdx[i] = -1 }
    left, maxLen := 0, 0
    for right := 0; right < len(s); right++ {
        if lastIdx[s[right]] >= left {
            left = lastIdx[s[right]] + 1
        }
        lastIdx[s[right]] = right
        maxLen = max(maxLen, right-left+1)
    }
    return maxLen
}

// 最小覆盖子串
func MinWindow(s string, t string) string {
    need := [128]int{}
    for _, ch := range t { need[ch]++ }
    left, count, start, minLen := 0, len(t), 0, len(s)+1

    for right := 0; right < len(s); right++ {
        if need[s[right]] > 0 { count-- }
        need[s[right]]--
        for count == 0 {
            if right-left+1 < minLen {
                start, minLen = left, right-left+1
            }
            need[s[left]]++
            if need[s[left]] > 0 { count++ }
            left++
        }
    }
    if minLen > len(s) { return "" }
    return s[start : start+minLen]
}
```

## 4. 快慢指针（链表相关见链表章节）

```go
// 移除元素（原地）
func RemoveElement(nums []int, val int) int {
    slow := 0
    for fast := 0; fast < len(nums); fast++ {
        if nums[fast] != val {
            nums[slow] = nums[fast]
            slow++
        }
    }
    return slow
}
```

## 完整对照表

| 模式 | TS | Go |
|------|-----|-----|
| 相向双指针 | `l=0,r=arr.length-1` | `left, right` |
| 快慢双指针 | `slow=0,fast=0` | 同 |
| 固定窗口 | `sum += arr[i] - arr[i-k]` | 同 |
| 可变窗口 | `while(cond) left++` | `for cond { left++ }` |
| 字符计数 | `Map<char, count>` | `[128]int` / `[256]int` |
