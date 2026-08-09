# validx 迭代计划与质量门槛

## 1. 迭代阶段

### P0 项目骨架

- go.mod(module github.com/lcylpzls/validx,go 1.26)、目录、CI
  (三平台 + staticcheck + govulncheck + tidy + apidiff)、错误码。

### P1 v0.1.0 核心

- required / omitempty / min / max / len / email / regexp / oneof;
- 嵌套结构体、slice/map dive、`-` 跳过;
- 规则预编译缓存、聚合错误、配置校验。

### P2 v0.2.0 自定义与规则扩充

- RegisterValidation / ValidateField;
- alpha / alphanum / numeric / boolean / uuid / url / ip /
  datetime / gt / lt / gte / lte。

### P3 v0.3.0 跨字段与示例

- eqfield / nefield / eq / ne;
- examples(用户注册、订单参数)。

### P4 v0.4.0 性能

- 编译缓存与反射路径优化;
- 与 validator 对比基准、fuzz(规则解析 / 校验)。

### P5 v0.5.0 工业级打磨

- 治理文件(SECURITY / CONTRIBUTING / PR 模板)、运行文档;
- 性能文档与决策收尾。

### P6 自我审查与提升(计划完成后)

- 计划阶段完成后,持续自审:API 一致性、边界行为、性能、
  文档完备性、竞品对比;
- 发现改进点即推进次版本并发布;
- 直至维护者(Codex)判断可工业化,停下并征询用户是否发布 1.0.0。

### P6 执行记录

- v0.6.0:修复 dive 元素 nil 指针绕过 required;float 截断比较 bug
  (1.9 误通过 max=1)+ 小数参数支持;Validate 支持切片/map 批量;
  错误消息带字段路径;Version 常量。
- v0.7.0:条件必填(required_if / required_unless,编译期校验引用);
  格式规则(base64 / json / hexadecimal / mac / semver);
  字符串规则(contains / excludes / startswith / endswith)。
- v0.8.0:条件禁止(excluded_if);网络规则(ipv4 / ipv6 / hostname /
  fqdn / port);规则面累计 41 个,覆盖日常校验场景。

## 2. 质量门槛(每阶段强制)

- 语句覆盖率 100%;`go vet` / `staticcheck` 零告警;
- `go test -race` 全绿;fuzz 至少 2 个目标(规则解析、校验路径);
- 三平台 CI × Go 1.26;govulncheck 零告警;
- go.mod tidy 漂移检查;apidiff 对比上一 tag;
- 所有日志、注释、文档使用简体中文。

## 3. 性能目标

| 场景 | 目标 |
| --- | --- |
| 简单结构体验证(5 字段) | < 1µs |
| 嵌套结构体 | 与单层同量级(缓存生效) |
| 规则解析失败 | 不缓存,重复暴露 |
