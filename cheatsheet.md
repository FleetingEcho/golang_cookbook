# Go 语言速查表 — Language Fundamentals

> TypeScript → Go 快速对照，覆盖 `01-language/` 全部 13 个文件的核心内容。
> 每个章节标题可点击跳转到对应的完整文档。

---

## [1. 变量声明 →](golang-cookbook/01-language/01-variables.md)

| 操作 | TypeScript | Go |
|------|-----------|-----|
| 声明 + 初始化 | `let x = 1` | `var x = 1` 或 `x := 1`（函数内） |
| 显式类型 | `let x: number = 1` | `var x int = 1` |
| 常量 | `const x = 1`（运行时） | `const x = 1`（编译时，仅标量/string/bool） |
| 多变量 | `let a=1, b=2` | `a, b := 1, 2` |
| 未初始化 | `let x: number` → `undefined` ❌ | `var x int` → `0` ✅ 零值安全 |
| 丢弃值 | 无语言支持 | `_`（blank identifier） |

```go
// := 至少有一个新变量才合法
x := 1
x, y := 2, 3  // ✅ y 是新变量

// iota — 自增枚举
const (
    A = iota // 0
    B        // 1
    C        // 2
)
```

---

## [2. 基本类型 →](golang-cookbook/01-language/02-primitives.md)

```
TS  number      = Go float64 + int/uint 组合
TS  string      = Go string（不可变）
TS  Uint8Array  = Go []byte
TS  unknown     = Go any（= interface{}）
```

### Go 数值体系（12 种）

```
int / int8 / int16 / int32 / int64         有符号
uint / uint8 / uint16 / uint32 / uint64    无符号
byte  = uint8                               二进制数据
rune  = int32                               Unicode 码点
float32 / float64                           浮点
complex64 / complex128                      复数
```

```go
// 类型转换必须显式
var a int = 3
var b int64 = 4
c := int64(a) + b  // ✅ 必须转
// c := a + b      // ❌ mismatched types

// string <-> []byte
b := []byte("hello")       // string → []byte
s := string(b)             // []byte → string

// string <-> int
n, _ := strconv.Atoi("42")      // → int
s := strconv.Itoa(42)           // → string
```

---

## [3. 函数 →](golang-cookbook/01-language/03-functions.md)

```go
func add(a, b int) int                // 基本
func div(a, b int) (int, error)       // 多返回值（最常用模式）
func sum(nums ...int) int             // 变参
func add(a, b int) (sum int) {        // 命名返回值
    sum = a + b
    return                            // 裸 return
}

// 泛型函数
func Map[T, U any](s []T, f func(T) U) []U

// 闭包
func makeCounter() func() int {
    count := 0
    return func() int {
        count++
        return count - 1
    }
}
```

### defer — 延迟执行（LIFO）

```go
f, _ := os.Open("file")
defer f.Close()       // 函数返回时执行

// defer 参数在注册时求值
x := 1
defer fmt.Println(x)  // 打印 1（不是函数返回时的 x）
x = 2

// defer 可修改命名返回值
func example() (n int) {
    defer func() { n++ }()
    return 5  // → 实际返回 6
}
```

---

## [4. 控制流 →](golang-cookbook/01-language/04-control-flow.md)

```go
// if — 可带短语句
if err := doSomething(); err != nil {
    return err
}

// for — Go 只有 for（没有 while/do-while）
for i := 0; i < n; i++ {}     // C 风格
for condition {}               // 相当于 while
for {}                         // 无限循环
for i, v := range items {}     // range 迭代

// switch — 不用 break，自动匹配第一个
switch x {
case 1:
    println("one")
case 2, 3:
    println("two or three")
default:
    println("other")
}

// switch 可代替 if-else 链
switch {
case x > 0:
    println("positive")
case x < 0:
    println("negative")
default:
    println("zero")
}
```

---

## [5. 指针 →](golang-cookbook/01-language/06-pointers.md)

```go
x := 42
p := &x            // 取地址
fmt.Println(*p)    // 解引用 → 42
*p = 100           // 修改原值
fmt.Println(x)     // 100

// 不可寻址的值（常见陷阱）
m := map[int]Point{0: {1, 2}}
// m[0].X = 3     // ❌ 不能取 map 元素的地址
p := m[0]          // ✅ 要读出 → 修改 → 写回
p.X = 3
m[0] = p
```

---

## [6. nil 与零值 →](golang-cookbook/01-language/07-nil-and-zero.md)

| 类型 | 零值 | 说明 |
|------|------|------|
| `bool` | `false` | |
| `int`/`float` | `0` | |
| `string` | `""` | 不是 `undefined` |
| `pointer` | `nil` | |
| `slice` | `nil` | 可安全 `append` |
| `map` | `nil` | **写入会 panic**，必须 `make` |
| `channel` | `nil` | 会永久阻塞 |
| `interface` | `nil` | |
| `struct` | 每个字段递归零值 | |

```go
var m map[string]int
// m["key"] = 1  // ❌ panic: assignment to nil map
m = make(map[string]int)
m["key"] = 1      // ✅

var s []int
s = append(s, 1)  // ✅ nil slice 可 append
```

---

## [7. Array vs Slice →](golang-cookbook/01-language/08-array-vs-slice.md)

```go
// 数组 — 值类型，固定长度
var arr [3]int = [3]int{1, 2, 3}
arr2 := [...]int{1, 2, 3}     // 编译器推断长度
arr == arr2                     // ✅ 数组可直接比较

// 切片 — 引用类型，动态长度
var sl []int                   // nil slice
sl = []int{1, 2, 3}            // 字面量
sl = make([]int, 5)            // len=5, cap=5
sl = make([]int, 3, 10)        // len=3, cap=10
sl = append(sl, 4)             // 追加
sl = sl[:2]                    // 截取

// 切片是底层数组的视图
a := []int{1, 2, 3, 4, 5}
b := a[1:4]                    // [2, 3, 4]
b[0] = 99                      // a 也被修改 → [1, 99, 3, 4, 5]
```

---

## [8. Struct（对象）→](golang-cookbook/01-language/09-object-vs-struct.md)

```go
type Person struct {
    Name string
    Age  int
}

// 声明
p := Person{Name: "Alice", Age: 30}
p := Person{"Alice", 30}        // 按字段顺序（不推荐）
var p Person                    // 零值：Name="", Age=0

// 方法（Go 的方法 = 带接收者的函数）
func (p Person) Greet() string {
    return "Hi, I'm " + p.Name
}
func (p *Person) Birthday() {
    p.Age++                      // 指针接收者可修改
}

// 嵌入（不是继承）
type Employee struct {
    Person                       // 嵌入 → 字段和方法被提升
    Title string
}
e := Employee{Person: Person{Name: "Bob"}, Title: "Engineer"}
fmt.Println(e.Name)              // "Bob" — 提升字段
```

---

## [9. Map →](golang-cookbook/01-language/11-map.md)

```go
m := make(map[string]int)
m := map[string]int{"a": 1, "b": 2}

v := m["a"]        // 1
v, ok := m["c"]    // 0, false — ok 表示 key 是否存在
delete(m, "a")     // 删除

// 遍历（顺序随机！）
for k, v := range m {
    fmt.Println(k, v)
}
```

---

## [10. 接口 (interface) →](golang-cookbook/01-language/09-object-vs-struct.md) <!-- 接口和 struct 在同一文件 -->

```go
// 定义
type Writer interface {
    Write([]byte) (int, error)
}

// 隐式实现 — 不需要 `implements` 关键字
type MyWriter struct{}
func (w MyWriter) Write(data []byte) (int, error) {
    return len(data), nil
}

// 空接口 = any（Go 1.18+）
var x any
x = 42
x = "hello"
s, ok := x.(string)  // 类型断言

// 类型 switch
switch v := x.(type) {
case int:
    fmt.Println("int:", v)
case string:
    fmt.Println("string:", v)
default:
    fmt.Println("unknown")
}
```

---

## [11. 泛型 →](golang-cookbook/01-language/12-generics.md)

```go
// 泛型函数
func Identity[T any](v T) T { return v }

// 泛型类型
type Stack[T any] struct {
    items []T
}
func (s *Stack[T]) Push(v T) {
    s.items = append(s.items, v)
}

// 约束
func Min[T ~int | ~float64](a, b T) T {
    if a < b { return a }
    return b
}

// comparable — 内置的可比较约束
func Contains[T comparable](s []T, v T) bool {
    for _, item := range s {
        if item == v { return true }
    }
    return false
}
```

| TS 泛型 | Go 泛型 |
|---------|---------|
| `T extends Constraint` | `T Constraint`（interface 做约束） |
| 方法可加额外类型参数 ✅ | 方法不能有自己的类型参数 ❌ |
| `class Box<T>` | `type Box[T any] struct{}` |

---

## [12. 错误处理 →](golang-cookbook/01-language/05-errors.md)

```go
// Go 没有 try/catch — error 是返回值
func doSomething() (int, error) {
    if bad {
        return 0, errors.New("something went wrong")
    }
    return 42, nil
}

// 调用方必须检查 error
result, err := doSomething()
if err != nil {
    log.Fatal(err)
}

// 自定义错误
type ValidationError struct {
    Field string
    Msg   string
}
func (e *ValidationError) Error() string {
    return fmt.Sprintf("%s: %s", e.Field, e.Msg)
}

// errors.Is / errors.As（Go 1.13+）
if errors.Is(err, io.EOF) { /* 特定错误值 */ }
var ve *ValidationError
if errors.As(err, &ve) { /* 特定错误类型 */ }

// panic/recover — 仅用于"不可能发生"的情况
defer func() {
    if r := recover(); r != nil {
        log.Println("recovered:", r)
    }
}()
```

---

## [13. 类型断言与类型转换 →](golang-cookbook/01-language/13-type-assertion.md)

```go
// 类型断言：interface → 具体类型
var x any = "hello"
s := x.(string)        // → "hello"（类型不对则 panic）
s, ok := x.(string)    // → "hello", true（安全）

// 类型转换：数值类型之间
var i int = 42
var f float64 = float64(i)
var u uint = uint(i)

// string ↔ []rune
r := []rune("你好")     // → [20320, 22909]
s := string([]rune{20320, 22909}) // → "你好"
```

---

## 高频陷阱速记

```go
// 1. 循环变量捕获（Go < 1.22）
for i := 0; i < 3; i++ {
    go func() { println(i) }() // ❌ 都打印 2
}
// 1.22+ 已修复，或用传参

// 2. map 元素不可取地址
m := map[int]Point{0: {1, 2}}
// m[0].X = 3  // ❌

// 3. nil map 写入 panic
var m map[string]int
// m["k"] = 1  // ❌

// 4. 切片截取共享底层数组
a := []int{1, 2, 3, 4, 5}
b := a[1:3]
b[0] = 99  // a[1] 也变成 99

// 5. int 溢出不报错
var x int32 = 2_000_000_000
x += 1_000_000_000  // 变成负数，不 panic

// 6. len("你好") = 6（UTF-8 字节数）
// utf8.RuneCountInString("你好") = 2

// 7. range 遍历 map 顺序随机

// 8. 未使用的变量/import 编译报错
```
