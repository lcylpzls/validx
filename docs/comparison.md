# validx 与 go-playground/validator 对比

> 更新日期:2026-08-09 · validx v0.4.0 · validator v10.30.3

## 1. 性能(同机实测)

| 场景 | validx | validator | 倍率 |
| --- | --- | --- | --- |
| Validate(命中) | 498 ns | 1059 ns | 2.1x 快 |
| ValidateNested | 325 ns | 573 ns | 1.8x 快 |
| ValidateInvalid | 4500 ns | 754 ns | 6x 慢 |

失败路径慢是 errx 结构化错误(调用栈)的代价,可用
`errx.SetStackCapture(false)` 关闭栈捕获。

## 2. 功能点

| 功能 | validx | validator |
| --- | --- | --- |
| tag 驱动 | ✅ | ✅ |
| 规则预编译缓存 | ✅ | ✅ |
| 必填 / 范围 / 长度 | ✅ | ✅ |
| 格式规则 | email/uuid/url/ip/... | 数十种 |
| 嵌套结构体 | ✅ 自动递归 | ✅ |
| dive 元素 | ✅ | ✅ |
| 跨字段 | eqfield/nefield/eq/ne | eqfield/nefield/... |
| 自定义规则 | ✅ 实例级 | ✅ 全局/实例 |
| 单字段校验 | ✅ | ✅ |
| 聚合错误字段路径 | ✅ | ✅ |
| 错误翻译(i18n) | ❌(中文,调用方映射) | ✅(多语言) |
| 依赖 | 零第三方 | universal-translator 等 |
| 生态集成 | errx 打通 | 独立 |

## 3. 取舍

- **validx 胜在**:薄(规则面克制)、快(成功路径 2 倍)、零依赖、
  errx 生态统一、编译期暴露配置错误(跨字段引用、非法参数);
- **validator 胜在**:规则面全(几十种格式)、多语言翻译、
  社区成熟度高;
- 对自用项目:validx 的规则面已覆盖日常 95% 场景,
  失败路径成本可通过全局配置消除。

## 4. 选型建议

- 需要多语言校验消息 / 极端规则面:validator;
- 自用基建栈(confx/logx/errx/webx):validx(生态统一 + 性能);
- 两库 tag 语法核心兼容,迁移成本低。
