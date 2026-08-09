package validx

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/lcylpzls/errx"
)

// Validator 是校验入口,持有配置、规则缓存与自定义规则表。
// 所有方法并发安全。
type Validator struct {
	cfg        config
	cache      sync.Map // reflect.Type -> *structRules
	fieldCache sync.Map // 规则串 -> []Rule
	mu         sync.RWMutex
	custom     map[string]ValidationFunc
}

// ValidationFunc 是自定义校验函数签名。
// value 为被校验值,param 为规则参数(可为空),path 为字段路径;
// 返回 nil 表示通过,返回错误表示失败(非 errx 错误会被包装并携带字段信息)。
type ValidationFunc func(value any, param string, path string) error

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

// RegisterValidation 实例级注册自定义规则。
// 规则名须合法且不与内置规则冲突;重复注册同名规则以最后一次为准。
func (v *Validator) RegisterValidation(name string, fn ValidationFunc) error {
	if !validRuleName(name) {
		return errx.NewCodef(CodeInvalidRule, "非法规则名 %q", name)
	}
	if _, ok := builtinRules[name]; ok {
		return errx.NewCodef(CodeInvalidRule, "规则 %q 与内置规则冲突", name)
	}
	if fn == nil {
		return errx.NewCodef(CodeInvalidRule, "规则 %q 的校验函数不能为空", name)
	}
	v.mu.Lock()
	if v.custom == nil {
		v.custom = make(map[string]ValidationFunc)
	}
	v.custom[name] = fn
	v.mu.Unlock()
	return nil
}

// ValidateField 以规则串校验单个值(不支持 dive)。
// 字段路径固定为 "value";规则解析结果按串缓存。
func (v *Validator) ValidateField(value any, rules string) error {
	compiled, err := v.compileFieldRules(rules)
	if err != nil {
		return err
	}
	rv := reflect.ValueOf(value)
	for rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			break
		}
		rv = rv.Elem()
	}
	return v.applyRules(rv, "value", compiled, reflect.Value{})
}

// compileFieldRules 解析并缓存单字段规则串。
func (v *Validator) compileFieldRules(tag string) ([]Rule, error) {
	if cached, ok := v.fieldCache.Load(tag); ok {
		return cached.([]Rule), nil
	}
	rules, err := v.compileRules(tag)
	if err != nil {
		return nil, err
	}
	for _, r := range rules {
		if r.name == "dive" {
			return nil, errx.NewCode(CodeInvalidRule,
				"ValidateField 不支持 dive 规则")
		}
	}
	v.fieldCache.Store(tag, rules)
	return rules, nil
}

// customFn 读取自定义规则函数(并发安全)。
func (v *Validator) customFn(name string) (ValidationFunc, bool) {
	v.mu.RLock()
	defer v.mu.RUnlock()
	fn, ok := v.custom[name]
	return fn, ok
}

// Validate 校验结构体:成功返回 nil,失败返回 errx 错误
// (单字段错误直接返回,多字段错误为 errx 聚合)。
func (v *Validator) Validate(value any) error {
	rv := reflect.ValueOf(value)
	if !rv.IsValid() {
		return errx.NewCode(CodeValidationFailed, "校验对象不能为 nil")
	}
	for rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return errx.NewCode(CodeValidationFailed, "校验对象不能为空指针")
		}
		rv = rv.Elem()
	}
	switch rv.Kind() {
	case reflect.Struct:
		return v.validateStruct(rv, "")
	case reflect.Slice, reflect.Array:
		return v.validateCollection(rv)
	case reflect.Map:
		return v.validateMap(rv)
	default:
		return errx.NewCodef(CodeValidationFailed,
			"校验对象必须是结构体/切片/map,当前类型 %s", rv.Kind())
	}
}

// validateCollection 校验切片/数组:元素必须是结构体(解指针),逐个校验聚合。
func (v *Validator) validateCollection(rv reflect.Value) error {
	var errs []error
	for i := 0; i < rv.Len(); i++ {
		ev := derefValue(rv.Index(i))
		if ev.Kind() != reflect.Struct {
			return errx.NewCodef(CodeValidationFailed,
				"切片元素 [%d] 必须是结构体,当前类型 %s", i, ev.Kind())
		}
		if err := v.validateStruct(ev, fmt.Sprintf("[%d]", i)); err != nil {
			errs = append(errs, err)
		}
	}
	return joinErrs(errs)
}

// validateMap 校验 map:值必须是结构体(解指针),逐个校验聚合。
func (v *Validator) validateMap(rv reflect.Value) error {
	var errs []error
	iter := rv.MapRange()
	for iter.Next() {
		ev := derefValue(iter.Value())
		if ev.Kind() != reflect.Struct {
			return errx.NewCodef(CodeValidationFailed,
				"map 值 [%v] 必须是结构体,当前类型 %s", iter.Key(), ev.Kind())
		}
		if err := v.validateStruct(ev, fmt.Sprintf("[%v]", iter.Key())); err != nil {
			errs = append(errs, err)
		}
	}
	return joinErrs(errs)
}
