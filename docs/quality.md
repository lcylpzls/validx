# 质量保障

## 门槛(每个版本强制)

- 语句覆盖率 100%;
- `go vet` / `staticcheck` 零告警;
- `go test -race` 全绿;
- fuzz 目标至少 2 个(规则解析、校验路径);
- 三平台 CI(ubuntu / windows / macos)× Go 1.26;
- govulncheck 零告警;go.mod tidy 无漂移;
- 示例全部可构建并通过 vet。

## 测试策略

- 每种规则:成功 / 失败 / 类型不适用三条路径;
- 嵌套与 dive:路径格式(点、索引、map key)断言;
- 跨字段:编译期引用校验与运行时防御分支;
- 自定义规则:注册冲突、覆盖、并发、错误包装;
- 并发:缓存竞争与注册/校验并行(race);
- 回归种子:CI fuzz 发现的边界输入入库(testdata/)。

## 性能

见 [performance.md](performance.md)。规则参数编译期预解析,
禁止在运行时重复解析;成功路径基准与 validator 对比纳入 CI 记录。

## API 兼容性

- <1.0.0 允许有意的破坏性变更,须在 CHANGELOG 说明;
- 破坏性变更按家族约定走 minor 版本,并在 CHANGELOG 显著标注。
