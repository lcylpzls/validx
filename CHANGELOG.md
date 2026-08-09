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
