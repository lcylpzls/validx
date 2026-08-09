# validx API 设计草案

> 本文是评审稿;v0.1.0 冻结时确认。

## 1. 快速上手

```go
type User struct {
	Name  string `validate:"required,min=2,max=20"`
	Email string `validate:"required,email"`
	Age   int    `validate:"min=0,max=150"`
	Tags  []string `validate:"dive,oneof=admin user guest"`
}

v, err := validx.New()
if err != nil {
	panic(err)
}
if err := v.Validate(user); err != nil {
	// err 为 errx 聚合错误,字段路径 + 规则名
}
```

## 2. 核心类型

```go
type Validator struct { /* 未导出 */ }

func New(opts ...Option) (*Validator, error)
func (v *Validator) Validate(value any) error
func (v *Validator) ValidateField(value any, rules string) error // v0.2.0
func (v *Validator) RegisterValidation(name string, fn ValidationFunc) error // v0.2.0
```

## 3. 配置

```go
func WithTagName(name string) Option // 默认 "validate",空串非法
```

## 4. 规则语法

- 逗号分隔规则,顺序执行,首个失败即记入该字段错误并停止;
- `name=param` 参数形式;`oneof=a b c` 空格分隔枚举;
- `dive` 之后规则作用于 slice / map 元素;
- `-` 跳过整个字段。

## 5. 错误

```go
// 错误码
VALIDX_INVALID_RULE         // 规则语法/参数非法(配置错误)
VALIDX_VALIDATION_FAILED    // 字段校验失败

// 字段错误携带:
field   // 路径,如 Profile.Address.City / Items[0].Name / Meta[key]
rule    // 失败规则名
```

## 6. 迭代范围

| 版本 | 内容 |
| --- | --- |
| v0.1.0 | required / omitempty / min / max / len / email / regexp / oneof / dive / 嵌套 / 聚合错误 |
| v0.2.0 | RegisterValidation / ValidateField / alpha / alphanum / numeric / boolean / uuid / url / ip / datetime / gt / lt / gte / lte |
| v0.3.0 | eqfield / nefield / eq / ne / 示例 |
| v0.4.0 | 性能优化 / 对比基准 / fuzz |
| v0.5.0 | 工业级打磨(治理 / 文档 / vuln / apidiff) |
