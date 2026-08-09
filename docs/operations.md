# 运行手册

## 规则速查

| 规则 | 适用 | 说明 |
| --- | --- | --- |
| required | 全部 | 非零值(string 非空、数字非 0、容器/指针非 nil) |
| omitempty | 全部 | 空值跳过后续规则 |
| min / max / len | 数值 / 字符串 / 容器 | 数值大小、字符数、元素数 |
| gt / lt / gte / lte | 同上 | 严格/非严格比较 |
| email | string | net/mail 权威解析 |
| regexp= | string | 预编译正则 |
| oneof= | string / 数值 / bool | 枚举白名单(空格分隔) |
| alpha / alphanum / numeric | string | 字母 / 字母数字 / 数字 |
| boolean | string | strconv.ParseBool |
| uuid | string | 8-4-4-4-12 格式 |
| url | string | 绝对 URL(scheme + host) |
| ip | string | IPv4 / IPv6 |
| datetime=layout | string | Go 时间布局 |
| eqfield / nefield | 结构体字段 | 同结构体字段比较(编译期校验引用) |
| eq / ne | string / 数值 / bool | 常量比较(自动字符串化) |
| dive | 容器 | 进入元素应用后续规则 |

## 常见场景

### 请求参数校验

```go
v, _ := validx.New()
if err := v.Validate(req); err != nil {
	// errx 聚合,可直接返回 HTTP 层
}
```

### 单字段校验

```go
if err := v.ValidateField(input, "required,email"); err != nil {
	// 路径固定为 value
}
```

### 自定义规则

```go
v.RegisterValidation("even", func(value any, _ string, path string) error {
	if n, ok := value.(int); !ok || n%2 != 0 {
		return errx.Newf(errx.KindInvalid, validx.CodeValidationFailed, "必须为偶数").
			WithField("field", path).WithField("rule", "even")
	}
	return nil
})
```

## 注意事项

- `Validator` 并发安全,可在多个 goroutine 间共享;
- 规则解析结果按类型缓存,首次校验后修改结构体 tag 不生效;
- 失败路径构造 errx 结构化错误(含调用栈),热路径敏感场景可
  `errx.SetStackCapture(false)` 全局关闭;
- 动态正则(用户提供的 regexp 参数)需评估 ReDoS 风险。
