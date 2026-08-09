package validx

import "github.com/lcylpzls/errx"

// 错误码定义:validx 各失败场景的错误码。
const (
	// CodeInvalidRule 规则语法或参数非法(配置错误)。
	CodeInvalidRule errx.Code = "VALIDX_INVALID_RULE"
	// CodeValidationFailed 字段校验失败。
	CodeValidationFailed errx.Code = "VALIDX_VALIDATION_FAILED"
)

func init() {
	errx.RegisterCode(CodeInvalidRule, "校验规则语法或参数非法")
	errx.RegisterCodeKind(CodeInvalidRule, errx.KindInvalid)
	errx.RegisterCode(CodeValidationFailed, "字段校验失败")
	errx.RegisterCodeKind(CodeValidationFailed, errx.KindInvalid)
}
