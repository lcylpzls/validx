# validx 架构设计

> 版本:v0.1.0(规划稿) · 状态:评审中

## 1. 总体分层

```text
业务代码
├── validx.Validator(Validate / ValidateField / RegisterValidation)
├── 编译层(规则解析 + 类型缓存)
├── 遍历层(反射:结构体 / 嵌套 / slice / map / dive)
├── 规则层(内置规则 + 自定义注册)
└── 错误层(errx 聚合,字段路径 + 规则名)
```

## 2. 核心模块职责

| 模块 | 职责 |
| --- | --- |
| `validator.go` | `Validator`、`New`、`Validate`、`ValidateField` |
| `compile.go` | tag 解析、规则预编译、类型缓存(sync.Map) |
| `reflect.go` | 反射遍历(嵌套 / dive / 指针解引用) |
| `rules.go` | 内置规则实现 |
| `registry.go` | 自定义规则注册 |
| `errors.go` | `VALIDX_*` 错误码与聚合 |
| `options.go` | `WithTagName`、配置校验 |

## 3. 编译层

- 首次校验某类型时解析全部字段 tag 并缓存:
  `sync.Map[reflect.Type]fieldRules`;
- 规则按 tag 逗号分隔:`required,min=3`;
- 非法规则返回 `VALIDX_INVALID_RULE` 并**不缓存**,
  保证配置错误每次暴露;
- `dive` 之后的规则绑定到元素类型,与当前字段规则分离。

## 4. 遍历层

- 指针:解引用(空指针由 required 判定);
- 结构体:逐字段取 tag 规则并校验;
- slice / map:遇到 `dive` 递归元素(元素为结构体时递归解析);
- 嵌套结构体:字段无显式规则时,若为结构体则自动递归;
- `-` tag 跳过。

## 5. 规则层

- 每条规则:名称 + 参数 + 预编译状态(regexp 预编译);
- 内置规则见 PRD 4.2;格式类复用标准库解析器;
- 自定义规则签名:`func(value any, param string, path string) error`,
  返回 errx 错误或 nil;返回普通 error 时包装为校验失败。

## 6. 错误模型

- `VALIDX_INVALID_RULE`:规则语法/参数非法(配置错误);
- `VALIDX_VALIDATION_FAILED`:字段校验失败,
  携带 field(路径)与 rule 字段;
- 多字段失败用 `errx.Join` 聚合,单字段错误透传;
- 路径格式:`Field` / `Field.Sub` / `Field[0]` / `Field[key]`。

## 7. 目标目录结构

```text
validx/
├── README.md
├── CHANGELOG.md
├── go.mod           # module github.com/lcylpzls/validx
├── validator.go
├── compile.go
├── reflect.go
├── rules.go
├── registry.go
├── errors.go
├── options.go
├── docs/
├── examples/
└── bench_test.go
```

## 8. 依赖策略

- 仅标准库 + errx,零第三方。
