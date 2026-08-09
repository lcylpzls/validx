# 贡献指南

感谢参与 validx 的打磨。请遵循以下约定。

## 环境与语言

- 开发机为 Windows,执行命令一律使用 PowerShell;
- 所有日志、注释与文档使用简体中文;
- 目标 Go 版本见 go.mod(当前 1.26)。

## 开发流程

1. 从 `docs/iteration-plan.md` 选择或提出迭代条目;
2. 在分支上实现,代码风格对齐现有文件(薄封装、显式命名、默认安全);
3. 本地验证:

```powershell
go vet ./...
go run honnef.co/go/tools/cmd/staticcheck@2026.1 ./...
go test -race -coverprofile=coverage.out ./...
go tool cover -func coverage.out
```

要求语句覆盖率 100%;新增规则必须同步更新
`docs/api-design.md`、README、CHANGELOG 与性能文档。

## 规则新增规范

- 规则名小写字母;参数语义明确;
- 编译期预解析参数(禁止运行时重复解析);
- 非字符串格式规则复用标准库解析器;
- 每种规则需覆盖成功 / 失败 / 类型不适用三条路径。

## 提交规范

- 提交信息以版本或主题开头,简述变更与验证结果;
- 涉及行为变更时在 CHANGELOG 记录;
- 破坏性变更仅允许在 <1.0.0 版本内,且需在 PR 说明。
