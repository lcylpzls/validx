# 更新日志

本项目遵循[语义化版本](https://semver.org/lang/zh-CN/)。

## [Unreleased]

### 规划

- 完成调研、PRD、架构、API 草案、ADR 与迭代计划。

## [v1.1.1] - 2026-08-10

### 修复

- examples/basic、examples/order 示例模块 go.mod 与最新依赖对齐
  （go mod tidy），修复 main CI 示例构建失败。

## [v1.1.0] - 2026-08-10

### 变更

- 家族测试底座接入：全部测试改用语义等价的 testx 断言
  （含 Require* 致命断言）；
- 测试依赖新增 `testx v1.2.0`，errx 同步升级 v1.4.0。

### 质量

- 根包语句覆盖率 100%；race / vet / staticcheck 全绿。

## [v0.1.0] - 2026-08-09

### 新增

- 核心校验:required / omitempty / min / max / len / email / regexp /
  oneof / dive;
- 嵌套结构体自动递归,slice / map 元素经 dive 校验;
- 规则预编译缓存(sync.Map),非法规则不缓存;
- 聚合错误统一 errx(字段路径 + 规则名);
- 配置:WithTagName 自定义 tag 名;
- 零第三方依赖;覆盖率 100%,race / vet / staticcheck / fuzz / vuln 全绿。

## [v0.2.0] - 2026-08-09

### 新增

- RegisterValidation:实例级自定义规则注册(参数可选,可覆盖,不与内置冲突);
- ValidateField:单字段校验(规则串缓存,不支持 dive);
- 规则扩充:alpha / alphanum / numeric / boolean / uuid / url / ip /
  datetime / gt / lt / gte / lte;
- 自定义规则错误:errx 透传,普通错误自动包装字段信息;
- 覆盖率 100%,race / vet / staticcheck / fuzz / vuln 全绿。

## [v0.3.0] - 2026-08-09

### 新增

- 跨字段规则:eqfield / nefield(同结构体字段比较,编译期校验引用),
  eq / ne(常量比较,数值/布尔自动字符串化);
- 跨字段支持指针字段自动解引用比较;
- dive 元素使用跨字段规则时明确报 VALIDX_INVALID_RULE;
- 示例:用户注册(跨字段确认)、订单参数(dive 元素);
- CI 新增 examples job;
- 覆盖率 100%,race / vet / staticcheck / fuzz / vuln 全绿。

## [v0.4.0] - 2026-08-09

### 性能

- 规则参数编译期预解析:min/max/len/gt/lt/gte/lte 整数、
  oneof 枚举拆分、regexp 预编译,消除运行时重复解析;
- Validate 命中 566 ns → 498 ns(-12%),分配 296 B → 200 B;
- 与 go-playground/validator 对比基准:成功路径 2 倍领先
  (docs/performance.md);失败路径成本为 errx 结构化错误
  (可 SetStackCapture(false) 关闭)。

### 质量

- 覆盖率 100%,race / vet / staticcheck / fuzz / vuln 全绿。

## [v0.5.0] - 2026-08-09

### 治理与文档

- SECURITY.md、CODEOWNERS、CONTRIBUTING、issue/PR 模板;
- operations(规则速查/场景)、quality(质量门槛)、release(发布流程)、
  comparison(与 validator 全维度对比)文档;
- 规则新增规范写入贡献指南(编译期预解析、三路径测试);
- 覆盖率 100%,race / vet / staticcheck / fuzz / vuln 全绿。

## [v0.6.0] - 2026-08-09

### 修复

- dive 元素为 nil 指针时 required 失效(现正确失败);
- float 截断比较 bug:1.9 误通过 max=1(现按浮点精确比较);
- 数值规则参数支持小数(min=1.5 / gt=2.5 等)。

### 新增

- Validate 直接接受切片 / map:元素须为结构体,逐个校验并聚合,
  路径含索引 / key;
- 错误消息带字段路径(字段 Name:为必填项);
- Version 常量。

### 质量

- 覆盖率 100%,race / vet / staticcheck / fuzz / vuln 全绿。

## [v0.7.0] - 2026-08-09

### 新增

- 条件必填:required_if / required_unless(跨字段,编译期校验引用);
- 格式规则:base64 / json / hexadecimal / mac / semver;
- 字符串规则:contains / excludes / startswith / endswith;
- 条件必填不适用于 dive 元素(明确报错);
- 覆盖率 100%,race / vet / staticcheck / fuzz / vuln 全绿。

## [v0.8.0] - 2026-08-09

### 新增

- 条件禁止:excluded_if(字段等于指定值时当前字段必须为空);
- 网络规则:ipv4 / ipv6 / hostname(RFC 1123)/ fqdn / port;
- 规则面累计 41 个,覆盖日常校验场景;
- 覆盖率 100%,race / vet / staticcheck / fuzz / vuln 全绿。

## [v0.9.0] - 2026-08-09

### 正式版预备

- 规则总表文档化(41 条,见 docs/api-design.md);
- 稳定性承诺写入 README(自 v1.0.0 生效);
- 新增 FuzzValidateBatch 批量校验 fuzz(CI 三目标);
- 基准复测与 41 规则性能确认;
- 覆盖率 100%,race / vet / staticcheck / fuzz / vuln 全绿。

## [v1.0.0] - 2026-08-09

### 正式版

- 公开 API 冻结,遵循语义化版本;
- Version 常量更新为 v1.0.0;
- README 稳定性承诺正式生效;
- docs/api-design.md 升级为正式 API 参考(41 条规则总表);
- 全量回归:100% 覆盖率、race、staticcheck、fuzz ×3、govulncheck、
  apidiff 对比 v0.9.0、三平台 CI。

### 版本历程

- v0.1.0 – v0.9.0:核心、自定义规则、跨字段、性能、治理、
  修复与规则扩充共 9 个迭代版本;
- v1.0.0:正式版,API 冻结。
