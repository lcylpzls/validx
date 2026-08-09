# validx API 参考

> 状态:**v1.0.0 API 已冻结**。新增规则与能力以次版本发布,
> 破坏性变更仅随主版本;任何修改须经 apidiff 对比并记录 CHANGELOG。

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
func (v *Validator) Validate(value any) error // 结构体 / 切片 / map(元素为结构体)
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

## 7. 规则总表(v0.9.0,41 条)

| 类别 | 规则 |
| --- | --- |
| 必填 | required / omitempty / required_if / required_unless / excluded_if |
| 数值与长度 | min / max / len / gt / lt / gte / lte(整数与小数参数) |
| 格式 | email / regexp / uuid / url / ip / ipv4 / ipv6 / hostname / fqdn / port / mac / base64 / json / hexadecimal / semver / datetime / boolean / numeric / alpha / alphanum |
| 字符串 | contains / excludes / startswith / endswith |
| 枚举 | oneof |
| 跨字段 | eqfield / nefield / eq / ne |
| 容器 | dive(元素规则) |
| 跳过 | `-` |

## 8. 稳定性说明

- v0.1.0 – v0.8.0 累计 8 个发布版本,API 仅做向后兼容扩展;
- v1.0.0 已冻结 API;新规则以次版本发布;
- 每个版本发布前执行:100% 覆盖率、race、staticcheck、fuzz、
  govulncheck、apidiff 对比与三平台 CI。
