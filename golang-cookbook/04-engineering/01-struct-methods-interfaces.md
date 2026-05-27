# 结构体方法与接口 — Methods & Interfaces

## 1. 方法（Methods）

```go
type Rectangle struct {
    Width, Height float64
}

// 值接收者
func (r Rectangle) Area() float64 {
    return r.Width * r.Height
}

// 指针接收者（修改原对象）
func (r *Rectangle) Scale(factor float64) {
    r.Width *= factor
    r.Height *= factor
}

// 工厂函数（无构造函数）
func NewRectangle(w, h float64) *Rectangle {
    return &Rectangle{Width: w, Height: h}
}

// 使用
r := Rectangle{3, 4}
fmt.Println(r.Area())  // 12
r.Scale(2)
fmt.Println(r.Area())  // 48
```

## 2. 接口（Interface）Go 核心

```go
// 接口定义（隐式实现——关键差异！）
type Shape interface {
    Area() float64
    Perimeter() float64
}

// 不需要写 "implements"！
// Rectangle 只要实现了 Area() 和 Perimeter()，就自动满足 Shape

func (r Rectangle) Perimeter() float64 {
    return 2 * (r.Width + r.Height)
}

type Circle struct{ Radius float64 }

func (c Circle) Area() float64 {
    return math.Pi * c.Radius * c.Radius
}

func (c Circle) Perimeter() float64 {
    return 2 * math.Pi * c.Radius
}

// 多态
func printShapeInfo(s Shape) {
    fmt.Printf("Area=%.2f, Perimeter=%.2f\n", s.Area(), s.Perimeter())
}

// 使用
shapes := []Shape{
    Rectangle{3, 4},
    Circle{5},
}
for _, s := range shapes {
    printShapeInfo(s)
}
```

```typescript
// TypeScript
interface Shape {
    area(): number;
    perimeter(): number;
}

// 需要显式 implements
class Rectangle implements Shape {
    constructor(public width: number, public height: number) {}
    area(): number { return this.width * this.height; }
    perimeter(): number { return 2 * (this.width + this.height); }
}
```

## 3. 嵌入（Embedding）替代继承

```go
type Logger struct{}

func (l Logger) Log(msg string) {
    fmt.Println("[log]", msg)
}

type Server struct {
    Logger          // 嵌入 Logger——方法被提升
    Host string
    Port int
}

srv := Server{Host: "localhost", Port: 8080}
srv.Log("server started")     // Logger 的方法被提升
srv.Logger.Log("explicit")    // 也可以显式调用
```

## 4. 接口的常见用法

```go
// 1. 标准库接口：io.Reader / io.Writer
type Reader interface {
    Read(p []byte) (n int, err error)
}

// 2. 空接口：any（Go 1.18+）
var v any = 42

// 3. 接口组合
type ReadWriter interface {
    Reader
    Writer
}

// 4. 类型断言
var s Shape = Rectangle{3, 4}
if r, ok := s.(Rectangle); ok {
    fmt.Println(r.Width)
}
```

## 速记

```
type I interface { Method() T }  — 接口定义
type S struct { ... }            — 结构体
func (s S) Method() T { }        — 值接收者方法
func (s *S) Method() T { }       — 指针接收者方法

!  接口隐式实现 — 不需要写 implements
!  嵌入 ≠ 继承 — 没有多态，只有提升
!  nil 接口 ≠ nil 指针 — 类型信息不为空
