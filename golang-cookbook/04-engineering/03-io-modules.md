# I/O 与模块系统

## 1. io.Reader / io.Writer

```go
// Go 最核心的接口
// type Reader interface { Read(p []byte) (n int, err error) }
// type Writer interface { Write(p []byte) (n int, err error) }

// 从文件读取
f, _ := os.Open("file.txt")
defer f.Close()

data, _ := io.ReadAll(f)              // 全部读取
io.Copy(os.Stdout, f)                 // 复制到标准输出

// 字符串
r := strings.NewReader("hello world")
buf := make([]byte, 5)
r.Read(buf) // buf = "hello"

// 组合
r = io.MultiReader(strings.NewReader("a"), strings.NewReader("b"))
data, _ = io.ReadAll(r) // "ab"

// bufio
scanner := bufio.NewScanner(os.Stdin)
for scanner.Scan() {
    fmt.Println(scanner.Text()) // 逐行读取
}
```

## 2. Go Modules

```bash
# 创建模块
go mod init example.com/myproject

# 添加依赖
go get github.com/gin-gonic/gin@v1.9.0

# 整理依赖
go mod tidy

# 升级
go get -u ./...
go get -u github.com/gin-gonic/gin

# 替换（本地开发）
# go.mod:
replace example.com/mylib => ../mylib

# vendor
go mod vendor
```

```go
// go.mod 示例
module github.com/user/project

go 1.22

require (
    github.com/gin-gonic/gin v1.9.0
    golang.org/x/sync v0.1.0
)

// 导出规则：大写字母开头的标识符是公开的
// 小写字母开头是包私有的
```

## 3. internal 包

```go
// 目录结构：
// project/
//   internal/
//     auth/     → 仅 project 内的包可导入
//   cmd/
//     server/   → main 包

// internal 包只能被同一 module 的父目录导入
// import "project/internal/auth"  // ✅
// 外部项目无法导入此包
```

## 4. main 函数与命令行参数

```go
package main

import (
    "flag"
    "fmt"
    "os"
)

func main() {
    // 方式 1：os.Args
    fmt.Println(os.Args[0]) // 程序路径

    // 方式 2：flag（推荐）
    name := flag.String("name", "world", "name to greet")
    port := flag.Int("port", 8080, "port number")
    flag.Parse()
    fmt.Printf("Hello %s, port %d\n", *name, *port)
}
```
