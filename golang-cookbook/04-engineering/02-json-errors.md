# JSON 序列化与 Error 包装

## 1. JSON 序列化

```go
import "encoding/json"

type User struct {
    ID    int    `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email,omitempty"` // 空值跳过
    Role  string `json:"role"`
    Token string `json:"-"`               // 忽略该字段
}

// 序列化
user := User{ID: 1, Name: "Alice", Role: "admin"}
data, err := json.Marshal(user)
// {"id":1,"name":"Alice","role":"admin"}

data, _ := json.MarshalIndent(user, "", "  ") // 美化输出

// 反序列化
var u User
json.Unmarshal(data, &u)

// 未知字段处理
type Config struct {
    Host string `json:"host"`
    Port int    `json:"port"`
}

// 允许未知字段（默认 UnownField 会报错）
decoder := json.NewDecoder(strings.NewReader(jsonStr))
decoder.DisallowUnknownFields() // 默认行为：遇到未知字段报错
```

## 2. 自定义 JSON 编解码

```go
type CustomTime struct {
    time.Time
}

const layout = "2006-01-02"

func (ct CustomTime) MarshalJSON() ([]byte, error) {
    return json.Marshal(ct.Format(layout))
}

func (ct *CustomTime) UnmarshalJSON(data []byte) error {
    var s string
    if err := json.Unmarshal(data, &s); err != nil {
        return err
    }
    t, err := time.Parse(layout, s)
    if err != nil { return err }
    ct.Time = t
    return nil
}
```

## 3. Error 包装

```go
// Go 1.13+ 错误链
import "errors"

var ErrNotFound = errors.New("not found")

func GetUser(id int) (*User, error) {
    if id <= 0 {
        return nil, fmt.Errorf("get user %d: %w", id, ErrNotFound)
    }
    return &User{ID: id}, nil
}

// 检查
_, err := GetUser(-1)
if errors.Is(err, ErrNotFound) {
    fmt.Println("not found, handle gracefully")
}

// 自定义错误类型
type ValidationError struct {
    Field string
    Value any
    Err   error
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation: %s = %v: %v", e.Field, e.Value, e.Err)
}

func (e *ValidationError) Unwrap() error { return e.Err }

// errors.As 检查
var valErr *ValidationError
if errors.As(err, &valErr) {
    fmt.Println("field:", valErr.Field)
}
```
