# validx

自研 tag 驱动参数校验库:声明式规则、嵌套结构体、dive 元素校验、
errx 聚合错误,零第三方依赖。

> 当前状态:**v0.9.0 实现完成,待 CI 验证与发布**。

## 定位

validx 不是表达式引擎,不解决「复杂条件逻辑」的问题;它解决
每个入口都要重复的部分:

- `validate` tag 声明必填 / 范围 / 长度 / 格式 / 枚举;
- 嵌套结构体与 slice / map 元素自动校验;
- 校验错误统一 errx 聚合,携带字段路径与规则名;
- 规则预编译缓存,热路径零重复反射。

## 特性

- 内置规则:required / omitempty / min / max / len / email / regexp /
  oneof / dive / alpha / alphanum / numeric / boolean / uuid / url / ip /
  datetime / gt / lt / gte / lte;
- 自定义规则:`RegisterValidation(name, fn)` 实例级注册;
- 单字段校验:`ValidateField(value, rules)`;
- 跨字段规则:`eqfield` / `nefield` / `eq` / `ne`;
- 嵌套结构体自动递归,slice / map 元素经 dive 校验;
- 批量校验:`Validate` 直接接受切片 / map(元素为结构体);
- 小数参数:`min=1.5` 等数值规则支持浮点比较;
- 条件必填:`required_if` / `required_unless`;
- 条件禁止:`excluded_if`;
- 网络规则:`ipv4` / `ipv6` / `hostname` / `fqdn` / `port`;
- 格式规则:`base64` / `json` / `hexadecimal` / `mac` / `semver`;
- 字符串规则:`contains` / `excludes` / `startswith` / `endswith`;
- 聚合错误统一 errx(字段路径 + 规则名)。

## 性能

- 规则参数编译期预解析,热路径零重复解析;
- 简单结构体验证 < 0.5µs,成功路径为 go-playground/validator 的 2 倍;
- 详见 [docs/performance.md](docs/performance.md)。

## 快速上手

```go
type User struct {
	Name  string   `validate:"required,min=2,max=32"`
	Email string   `validate:"required,email"`
	Age   int      `validate:"min=0,max=150"`
	Roles []string `validate:"dive,oneof=admin user guest"`
}

v, err := validx.New()
if err != nil {
	panic(err)
}
if err := v.Validate(user); err != nil {
	// errx 聚合错误:字段路径 + 规则名
}
```

## 质量门槛

- 语句覆盖率 100%,race、vet、staticcheck、fuzz、govulncheck 全绿;
- 三平台 CI(ubuntu / windows / macos);
- 简单结构体验证基准目标 < 1µs。

## 稳定性承诺(自 v1.0.0 生效)

- 本库遵循[语义化版本](https://semver.org/lang/zh-CN/);
- v1.0.0 起公开 API 冻结:新增规则与能力以次版本发布,
  破坏性变更仅随主版本;
- 每个版本发布前执行:100% 覆盖率、race、staticcheck、fuzz、
  govulncheck、apidiff 对比与三平台 CI。

## 文档

- [docs/README.md](docs/README.md) — 文档索引
- [docs/validation-research.md](docs/validation-research.md) — 校验领域调研手册
- [docs/operations.md](docs/operations.md) — 规则速查与常见场景
- [docs/comparison.md](docs/comparison.md) — 与 validator 全维度对比
- [examples/basic](examples/basic) — 用户注册(跨字段确认)
- [examples/order](examples/order) — 订单参数(dive 元素)

## 贡献与安全

- [CONTRIBUTING.md](CONTRIBUTING.md) — 开发流程与质量门槛
- [SECURITY.md](SECURITY.md) — 安全说明与漏洞报告

## License

MIT © [lcylpzls](https://github.com/lcylpzls)
