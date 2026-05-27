# 测试 — Testing

## 1. 单元测试

```go
// file: math.go
package mathutil

func Add(a, b int) int { return a + b }

func Divide(a, b int) (int, error) {
    if b == 0 { return 0, errors.New("division by zero") }
    return a / b, nil
}

// file: math_test.go（与源码同包，文件名 _test.go）
package mathutil

import "testing"

// 基本测试
func TestAdd(t *testing.T) {
    got := Add(1, 2)
    want := 3
    if got != want {
        t.Errorf("Add(1,2) = %d; want %d", got, want)
    }
}

// 表格驱动测试（Go 标准模式）
func TestDivide(t *testing.T) {
    tests := []struct {
        name    string
        a, b    int
        want    int
        wantErr bool
    }{
        {"normal", 10, 2, 5, false},
        {"by zero", 10, 0, 0, true},
        {"negative", -6, 3, -2, false},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := Divide(tt.a, tt.b)
            if (err != nil) != tt.wantErr {
                t.Errorf("Divide() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if got != tt.want {
                t.Errorf("Divide() = %d, want %d", got, tt.want)
            }
        })
    }
}

// 子测试
func TestMath(t *testing.T) {
    t.Run("addition", func(t *testing.T) {
        if Add(1, 2) != 3 { t.Fail() }
    })
    t.Run("subtraction", func(t *testing.T) {
        if 3-1 != 2 { t.Fail() }
    })
}
```

## 2. 测试辅助函数

```go
// 通过/失败标记
func TestWithHelper(t *testing.T) {
    t.Log("this will show on verbose mode")
    // t.Fatal("stop now")     // 立即停止
    // t.Fatalf("format %d", 1)
    // t.Error("non-fatal")     // 继续执行
    // t.Errorf("format %d", 1)
}

// 临时目录
func TestTempDir(t *testing.T) {
    dir := t.TempDir() // 自动清理
    f, _ := os.CreateTemp(dir, "test*.txt")
    f.Write([]byte("hello"))
    f.Close()
}

// 跳过
func TestWithSkip(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping in short mode")
    }
}
```

## 3. 基准测试

```go
func BenchmarkAdd(b *testing.B) {
    for i := 0; i < b.N; i++ {
        Add(1, 2)
    }
}

// 运行：
// go test -bench=. -benchmem
// 输出：BenchmarkAdd-8   1000000000   0.3 ns/op   0 B/op   0 allocs/op
```

## 4. 运行测试

```bash
# 运行所有
go test ./...

# 运行当前包
go test -v

# 运行特定测试
go test -run TestAdd

# 基准测试
go test -bench=. -benchmem

# 覆盖度
go test -cover
go test -coverprofile=coverage.out
go tool cover -html=coverage.out

# 竞态检测
go test -race ./...

# 短模式
go test -short ./...
```

## 完整对照表

| 操作 | Jest/Vitest | Go testing |
|------|-------------|-----------|
| 测试文件 | `*.test.ts` | `*_test.go` |
| 定义测试 | `test('name', ()=>{})` | `func TestXxx(t *testing.T)` |
| 断言 | `expect(a).toBe(b)` | `if got != want { t.Errorf(...) }` |
| 表格驱动 | `.each` | `[]struct + t.Run` |
| 设置 | `beforeEach` | `t.Cleanup` / 普通函数 |
| 临时目录 | `tmpdir` 包 | `t.TempDir()` |
| 基准测试 | `vitest bench` | `func BenchmarkXxx(b *testing.B)` |
| 覆盖度 | `--coverage` | `-coverprofile` |
| 竞态 | 无 | `-race` |
