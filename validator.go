package validx

import (
	"reflect"
	"sync"

	"github.com/lcylpzls/errx"
)

// Validator 是校验入口,持有配置、规则缓存与自定义规则表。
// 所有方法并发安全。
type Validator struct {
	cfg   config
	cache sync.Map // reflect.Type -> *structRules
}

// New 创建校验器。配置非法时返回 VALIDX_INVALID_RULE。
func New(opts ...Option) (*Validator, error) {
	cfg := defaultConfig()
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	if err := validateConfig(cfg); err != nil {
		return nil, err
	}
	return &Validator{cfg: cfg}, nil
}

// Validate 校验结构体:成功返回 nil,失败返回 errx 错误
// (单字段错误直接返回,多字段错误为 errx 聚合)。
func (v *Validator) Validate(value any) error {
	rv := reflect.ValueOf(value)
	if !rv.IsValid() {
		return errx.New(errx.KindInvalid, CodeValidationFailed, "校验对象不能为 nil")
	}
	for rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return errx.New(errx.KindInvalid, CodeValidationFailed, "校验对象不能为空指针")
		}
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return errx.Newf(errx.KindInvalid, CodeValidationFailed,
			"校验对象必须是结构体,当前类型 %s", rv.Kind())
	}
	return v.validateStruct(rv, "")
}
