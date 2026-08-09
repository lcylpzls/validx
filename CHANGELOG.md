# 更新日志

本项目遵循[语义化版本](https://semver.org/lang/zh-CN/)。

## [Unreleased]

### 规划

- 完成调研、PRD、架构、API 草案、ADR 与迭代计划。

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
