# validx 产品需求(PRD)

> 版本:v1.0.0(正式版) · 状态:已发布

## 1. 背景与动机

业务代码的参数校验重复且分散:

- 每个入口手写 if 判断,规则与字段分离,难维护;
- 校验错误格式不统一,HTTP 层难聚合返回;
- 与 errx / webx 生态不打通。

结论:**做一个 tag 驱动、薄而克制、零第三方的结构体校验库**,
业务代码用 `validate` tag 声明规则,错误统一 errx 聚合。

## 2. 目标

1. tag 驱动:`validate:"required,min=3"`,字段就近可见;
2. 常用规则:必填 / 范围 / 长度 / 格式 / 枚举;
3. 嵌套结构体 / slice / map 自动校验(dive);
4. 聚合错误带字段路径与规则名,统一 errx;
5. 规则预编译缓存,热路径零重复反射;
6. 自定义规则实例级注册,零第三方依赖。

## 3. 非目标(明确不做)

- 不做表达式引擎(CEL);
- 不做多语言翻译框架(错误消息中文,调用方自行映射);
- 不做 web 框架绑定(webx 通过文档适配);
- 不兼容 go-playground/validator 的全部 API;
- 不做全局注册表。

## 4. 能力需求

### 4.1 核心

- `New(opts ...Option) (*Validator, error)`;
- `Validate(value any) error`:结构体校验,成功返回 nil,
  失败返回 errx 聚合错误;
- 嵌套 struct 自动递归;slice / map 元素经 `dive` 递归;
- `-` 跳过字段。

### 4.2 规则(v0.1.0)

- `required`:非零值(string 非空、数字非 0、指针/切片/map 非 nil);
- `omitempty`:空值跳过后续规则;
- `min=` / `max=` / `len=`:数值大小、字符串/切片/map 长度;
- `email`:net/mail 权威解析;
- `regexp=`:正则匹配;
- `oneof=a b c`:枚举白名单;
- `dive`:进入切片/map 元素应用后续规则。

### 4.3 自定义与单字段(v0.2.0)

- `RegisterValidation(name, fn)` 实例级注册;
- `ValidateField(value any, rules string) error`;
- 补充规则:alpha / alphanum / numeric / boolean / uuid / url / ip /
  datetime / gt / lt / gte / lte。

### 4.4 跨字段(v0.3.0)

- `eqfield=Password` / `nefield=Name` / `eq=` / `ne=`;
- 示例(用户注册、订单参数)。

### 4.5 观测与配置

- `WithTagName(name)` 自定义 tag 名(默认 validate);
- 规则解析错误返回 `VALIDX_INVALID_RULE`(配置错误,启动即暴露)。

## 5. 非功能需求

- **性能**:规则预编译缓存,简单结构体验证 < 1µs;
- **质量**:语句覆盖率 100%、race、staticcheck、vet、fuzz、三平台 CI;
- **依赖**:标准库 + errx,零第三方。

## 6. 验收标准

v0.1.0 发布时:

1. 全部内置规则与嵌套/dive 全路径测试;
2. 聚合错误字段路径与规则名断言;
3. 非法规则返回 VALIDX_INVALID_RULE;
4. 100% 语句覆盖率,race / staticcheck / vet 全绿;
5. 基准与 validator 对比基线写入文档。
