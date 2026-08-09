# 性能基准

## 方法

```powershell
go test -run '^$' -bench . -benchmem -benchtime=1s .
```

CI 的 bench job 记录每次 main 推送的基准日志(artifact),不设硬性门禁。

## 参考数据(v0.4.0,Windows / AMD Ryzen 5 7600)

| Benchmark | ns/op | B/op | allocs/op |
| --- | --- | --- | --- |
| Validate(5 字段命中) | 498 | 200 | 10 |
| ValidateInvalid(含失败) | 4500 | 2481 | 38 |
| ValidateNested(订单+元素) | 325 | 120 | 6 |

简单结构体验证稳定在 0.5µs 内,满足 <1µs 目标。

## 与 go-playground/validator 对比(临时工程,未入库)

对比代码位于本地 `.bench-compare/`(已 gitignore):

```powershell
cd .bench-compare
go test -run '^$' -bench . -benchmem -benchtime=1s .
```

实测(同机,validator v10.30.3):

| 场景 | validx | validator | 倍率 |
| --- | --- | --- | --- |
| Validate(命中) | 498 ns | 1059 ns | 2.1x 快 |
| ValidateNested | 325 ns | 573 ns | 1.8x 快 |
| ValidateInvalid | 4500 ns | 754 ns | 6x 慢 |

解读:

- **成功路径**领先约 2 倍:规则参数编译期预解析(整数/枚举/正则),
  运行时零重复解析;validator 的规则分派与缓存更重;
- **失败路径**较慢是刻意的生态成本:validx 每个字段错误构造
  errx 结构化错误(携带调用栈与字段),validator 返回轻量结构;
  业务可通过 `errx.SetStackCapture(false)` 全局关闭栈捕获,
  失败路径显著下降;
- 失败通常不在热路径,该取舍换取 errx 生态统一。

## v0.3.0 → v0.4.0 优化

- 规则参数编译期预解析(min/max/len/gt/lt/gte/lte 整数、
  oneof 枚举拆分、regexp 预编译),消除运行时重复解析;
- Validate 命中 566 ns → 498 ns(-12%),分配 296 B → 200 B。
