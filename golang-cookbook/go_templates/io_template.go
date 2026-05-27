// 标准输入输出模板（ACM / 刷题模式）
// 使用：go run io_template.go < input.txt
// 或手动输入，Ctrl+D 结束
package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
)

// ---------- 输入工具 ----------
var scanner = bufio.NewScanner(os.Stdin)

func init() {
	scanner.Split(bufio.ScanWords)
}

func readInt() int {
	scanner.Scan()
	n, _ := strconv.Atoi(scanner.Text())
	return n
}

func readInt64() int64 {
	scanner.Scan()
	n, _ := strconv.ParseInt(scanner.Text(), 10, 64)
	return n
}

func readFloat() float64 {
	scanner.Scan()
	f, _ := strconv.ParseFloat(scanner.Text(), 64)
	return f
}

func readString() string {
	scanner.Scan()
	return scanner.Text()
}

func readInts(n int) []int {
	nums := make([]int, n)
	for i := range nums {
		nums[i] = readInt()
	}
	return nums
}

func readIntLines(n int) [][]int {
	lines := make([][]int, n)
	for i := range lines {
		lines[i] = readInts(n)
	}
	return lines
}

// ---------- 输出工具 ----------
var writer = bufio.NewWriter(os.Stdout)

func printf(format string, args ...any) {
	fmt.Fprintf(writer, format, args...)
}

func println(args ...any) {
	fmt.Fprintln(writer, args...)
}

// ---------- 常用算法函数 ----------
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
func abs(a int) int {
	if a < 0 {
		return -a
	}
	return a
}

// ---------- 主函数 ----------
func main() {
	defer writer.Flush()

	// 示例：两数之和
	a := readInt()
	b := readInt()
	println(a + b)

	// 多行示例
	n := readInt()
	nums := readInts(n)
	sum := 0
	for _, v := range nums {
		sum += v
	}
	println(sum)
}
